// Package extras backs the BigQuery / Spanner extension functions
// that have no upstream signature ID in the GoogleSQL builtin catalog:
//
//   - DLP_KEY_CHAIN / DLP_DETERMINISTIC_ENCRYPT / DLP_DETERMINISTIC_DECRYPT
//   - OBJ.MAKE_REF / OBJ.FETCH_METADATA / OBJ.GET_ACCESS_URL / OBJ.GET_READ_URL
//   - AI.IF / AI.CLASSIFY / AI.SCORE / ML.PREDICT
//
// All three families wrap external services in production
// (Cloud DLP, GCS, Vertex AI). googlesqlite is an in-process backend,
// so the implementations are local stand-ins:
//
//   - DLP_* uses a deterministic AES-GCM tied to the wrapped-key
//     payload; the same (key, plaintext, surrogate, context) tuple
//     always yields the same ciphertext, mirroring the production
//     semantic guarantee.
//   - OBJ.* operates on a small JSON-shaped ObjectRef carried as a
//     STRING. The metadata / access-URL outputs are derived
//     deterministically from the URI so round-trips work without an
//     actual GCS backend.
//   - AI.IF / AI.CLASSIFY / AI.SCORE / ML.PREDICT return deterministic
//     placeholder values (false-branch / first category / score 0 /
//     identity passthrough). Callers that need real ML behaviour must
//     route through their own backend before reaching SQL.
package extras

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// ----- DLP -----

// BindDlpKeyChain builds a serialized BYTES key payload from the
// Cloud KMS resource name + the wrapped-key bytes. The output is a
// fixed JSON envelope so DLP_DETERMINISTIC_ENCRYPT / _DECRYPT can
// recover the resource name + key material deterministically.
func BindDlpKeyChain(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("DLP_KEY_CHAIN: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	resource, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	wrapped, err := args[1].ToBytes()
	if err != nil {
		return nil, err
	}
	envelope := struct {
		Resource string `json:"resource"`
		Wrapped  string `json:"wrapped"`
	}{
		Resource: resource,
		Wrapped:  base64.StdEncoding.EncodeToString(wrapped),
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

func dlpDeriveKey(keyBytes []byte, context string) []byte {
	mac := hmac.New(sha256.New, keyBytes)
	_, _ = mac.Write([]byte(context))
	sum := mac.Sum(nil)
	return sum[:32]
}

func dlpNonce(key []byte, surrogate, context, plaintext string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(surrogate))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(context))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(plaintext))
	return mac.Sum(nil)[:12]
}

// BindDlpDeterministicEncrypt encrypts STRING `plaintext` using an
// AES-GCM deterministic scheme keyed by the DLP_KEY_CHAIN envelope.
// Output shape: `<surrogate>(<n>):<base64-ciphertext>` when surrogate
// is non-empty, otherwise just `<base64-ciphertext>`.
func BindDlpDeterministicEncrypt(args ...value.Value) (value.Value, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("DLP_DETERMINISTIC_ENCRYPT: invalid number of arguments: got %d, want between 3 and 4", len(args))
	}
	if helper.ExistsNull(args[:3]) {
		return nil, nil
	}
	keyEnv, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	plaintext, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	surrogate, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	context := ""
	if len(args) == 4 && args[3] != nil {
		if s, err := args[3].ToString(); err == nil {
			context = s
		}
	}
	key := dlpDeriveKey(keyEnv, context)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := dlpNonce(key, surrogate, context, plaintext)
	sealed := gcm.Seal(nil, nonce, []byte(plaintext), []byte(surrogate))
	out := append(nonce[:], sealed...)
	encoded := base64.StdEncoding.EncodeToString(out)
	if surrogate == "" {
		return value.StringValue(encoded), nil
	}
	return value.StringValue(fmt.Sprintf("%s(%d):%s", surrogate, len(encoded), encoded)), nil
}

// BindDlpDeterministicDecrypt is the inverse of
// BindDlpDeterministicEncrypt.
func BindDlpDeterministicDecrypt(args ...value.Value) (value.Value, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("DLP_DETERMINISTIC_DECRYPT: invalid number of arguments: got %d, want between 3 and 4", len(args))
	}
	if helper.ExistsNull(args[:3]) {
		return nil, nil
	}
	keyEnv, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	ciphertext, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	surrogate, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	context := ""
	if len(args) == 4 && args[3] != nil {
		if s, err := args[3].ToString(); err == nil {
			context = s
		}
	}
	if surrogate != "" {
		prefix := fmt.Sprintf("%s(", surrogate)
		idx := strings.Index(ciphertext, prefix)
		if idx == 0 {
			// strip "<surrogate>(<n>):"
			if colon := strings.Index(ciphertext, "):"); colon > 0 {
				ciphertext = ciphertext[colon+2:]
			}
		}
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("DLP_DETERMINISTIC_DECRYPT: base64: %w", err)
	}
	if len(raw) < 12 {
		return nil, fmt.Errorf("DLP_DETERMINISTIC_DECRYPT: ciphertext too short")
	}
	nonce := raw[:12]
	body := raw[12:]
	key := dlpDeriveKey(keyEnv, context)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	out, err := gcm.Open(nil, nonce, body, []byte(surrogate))
	if err != nil {
		return nil, fmt.Errorf("DLP_DETERMINISTIC_DECRYPT: %w", err)
	}
	return value.StringValue(string(out)), nil
}

// ----- ObjectRef -----

// BindObjMakeRef builds an ObjectRef carried as a STRING containing a
// JSON envelope {uri, authorizer?, version?, details?}. The
// JSON-input form (`OBJ.MAKE_REF(json)`) and re-authorize form
// (`OBJ.MAKE_REF(objectref, authorizer)`) are detected by the first
// argument's runtime shape.
func BindObjMakeRef(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("OBJ.MAKE_REF: missing argument")
	}
	if args[0] == nil {
		return nil, nil
	}
	first, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	envelope := map[string]any{}
	if strings.HasPrefix(first, "{") {
		// JSON form or existing ObjectRef.
		_ = json.Unmarshal([]byte(first), &envelope)
		if len(args) >= 2 && args[1] != nil {
			if auth, err := args[1].ToString(); err == nil {
				envelope["authorizer"] = auth
			}
		}
	} else {
		envelope["uri"] = first
		if len(args) >= 2 && args[1] != nil {
			if auth, err := args[1].ToString(); err == nil && auth != "" {
				envelope["authorizer"] = auth
			}
		}
	}
	for i := 2; i < len(args); i++ {
		// Ignore further positional args for the stub form.
		_ = i
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

// BindObjFetchMetadata returns a deterministic metadata STRUCT
// (encoded as JSON) derived from the ObjectRef's URI.
func BindObjFetchMetadata(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("OBJ.FETCH_METADATA: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	envelope := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &envelope)
	uri, _ := envelope["uri"].(string)
	mac := hmac.New(sha256.New, []byte("googlesqlite-objref"))
	_, _ = mac.Write([]byte(uri))
	digest := mac.Sum(nil)
	out := map[string]any{
		"uri":          uri,
		"content_type": "application/octet-stream",
		"md5_hash":     base64.StdEncoding.EncodeToString(digest[:16]),
		"size":         int64(binary.BigEndian.Uint32(digest[16:20])),
		"updated":      int64(binary.BigEndian.Uint32(digest[20:24])),
	}
	b, _ := json.Marshal(out)
	return value.JsonValue(string(b)), nil
}

// BindObjGetAccessUrl returns a deterministic placeholder signed-URL
// derived from the ObjectRef.
func BindObjGetAccessUrl(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("OBJ.GET_ACCESS_URL: missing argument")
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	envelope := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &envelope)
	uri, _ := envelope["uri"].(string)
	return value.StringValue("https://storage.googleapis.local/" + strings.TrimPrefix(uri, "gs://")), nil
}

// BindObjGetReadUrl is an alias of OBJ.GET_ACCESS_URL for the stub
// runtime; production semantics differ only in the URL signature TTL.
func BindObjGetReadUrl(args ...value.Value) (value.Value, error) {
	return BindObjGetAccessUrl(args...)
}

// ----- Spanner ML / AI -----

// BindAiIf returns the `else_value` branch — the stub never
// classifies a predicate as true. Deterministic so query plans are
// stable.
func BindAiIf(args ...value.Value) (value.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("AI.IF: invalid number of arguments: got %d, want at least 3", len(args))
	}
	return args[2], nil
}

// BindAiClassify returns the first candidate category, or NULL when
// the categories array is empty.
func BindAiClassify(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("AI.CLASSIFY: invalid number of arguments: got %d, want at least 2", len(args))
	}
	if args[1] == nil {
		return nil, nil
	}
	if av, ok := args[1].(*value.ArrayValue); ok {
		if len(av.Values) > 0 {
			return av.Values[0], nil
		}
		return nil, nil
	}
	return args[1], nil
}

// BindAiScore returns a deterministic score of 0.5. Sentinel for the
// stub; consumers that need real scoring must route through their own
// model harness.
func BindAiScore(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("AI.SCORE: missing argument")
	}
	return value.FloatValue(0.5), nil
}

// BindMlPredict returns the input features unchanged. Sufficient for
// the stub surface (the function exists for SQL parses); production
// ML.PREDICT would call into a hosted model.
func BindMlPredict(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ML.PREDICT: invalid number of arguments: got %d, want at least 2", len(args))
	}
	return args[1], nil
}
