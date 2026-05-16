// Package aead implements the BigQuery AEAD encryption functions
// (AEAD.ENCRYPT, AEAD.DECRYPT_*, DETERMINISTIC_ENCRYPT, KEYS.*).
//
// The BigQuery production runtime delegates to Tink with Cloud KMS for
// keyset wrapping. googlesqlite is a local execution backend, so we
// implement a pure-Go AEAD compatible with the public surface:
//
//   - Keysets are JSON documents shaped as
//     {"keys":[{"id":N,"algorithm":"AEAD_AES_GCM_256",
//     "key":"<base64-32-bytes>","primary":true}]}
//     This is interchange-compatible with Tink's KeysetJson form for the
//     subset of fields googlesqlite needs.
//   - Ciphertext layout: 4-byte big-endian key id || 12-byte nonce ||
//     AES-GCM ciphertext + 16-byte tag. The 4-byte prefix mirrors
//     Tink's "TINK" key-id-prefix output type so consumers can pick the
//     matching key at decrypt time.
//   - Deterministic variants derive the nonce as HMAC-SHA256(key,
//     additional_data || plaintext)[:12] so the same (key, plaintext,
//     additional_data) tuple always emits the same ciphertext.
package aead

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// keyset is the JSON-encoded keyset shape persisted in the BYTES
// surface of KEYS.NEW_KEYSET. Fields use snake_case to match Tink's
// JSON format and to remain forward-compatible with future fields.
type keyset struct {
	Keys []keysetKey `json:"keys"`
}

type keysetKey struct {
	ID        uint32 `json:"id"`
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
	Primary   bool   `json:"primary"`
}

const (
	algAESGCM      = "AEAD_AES_GCM_256"
	algAESSIVDET   = "DETERMINISTIC_AEAD_AES_SIV_CMAC_256"
	tinkPrefixSize = 4
	gcmNonceSize   = 12
	gcmKeySize     = 32
)

func loadKeyset(b []byte) (*keyset, error) {
	if len(b) == 0 {
		return nil, errors.New("keyset is empty")
	}
	var ks keyset
	if err := json.Unmarshal(b, &ks); err != nil {
		return nil, fmt.Errorf("keyset parse: %w", err)
	}
	if len(ks.Keys) == 0 {
		return nil, errors.New("keyset has no keys")
	}
	return &ks, nil
}

func (ks *keyset) primary() *keysetKey {
	for i := range ks.Keys {
		if ks.Keys[i].Primary {
			return &ks.Keys[i]
		}
	}
	return &ks.Keys[0]
}

func (ks *keyset) findID(id uint32) *keysetKey {
	for i := range ks.Keys {
		if ks.Keys[i].ID == id {
			return &ks.Keys[i]
		}
	}
	return nil
}

func decodeKey(k *keysetKey) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(k.Key)
	if err != nil {
		return nil, fmt.Errorf("key base64: %w", err)
	}
	return raw, nil
}

func generateAESGCMKey() ([]byte, error) {
	key := make([]byte, gcmKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func serializeKeyset(ks *keyset) ([]byte, error) {
	return json.Marshal(ks)
}

// keysetFromArg extracts a keyset BYTES payload from a function
// argument. STRUCT-shaped keysets (the KEYS.KEYSET_CHAIN form) carry
// the resolved payload in a `keyset` STRING field; we accept that
// path too.
func keysetFromArg(v value.Value) ([]byte, error) {
	if v == nil {
		return nil, errors.New("keyset is NULL")
	}
	switch x := v.(type) {
	case value.BytesValue:
		return []byte(x), nil
	case value.StringValue:
		return []byte(string(x)), nil
	case *value.StructValue:
		for i, k := range x.Keys {
			if k == "keyset" && i < len(x.Values) && x.Values[i] != nil {
				return keysetFromArg(x.Values[i])
			}
		}
		return nil, errors.New("keyset STRUCT has no `keyset` field")
	}
	s, err := v.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func plaintextFromArg(v value.Value) ([]byte, error) {
	switch x := v.(type) {
	case value.BytesValue:
		return []byte(x), nil
	}
	s, err := v.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// BindKeysNewKeyset returns a fresh keyset of the requested algorithm.
func BindKeysNewKeyset(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("KEYS.NEW_KEYSET: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	alg, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	switch alg {
	case algAESGCM, algAESSIVDET:
		// supported
	default:
		return nil, fmt.Errorf("KEYS.NEW_KEYSET: unsupported algorithm %q", alg)
	}
	key, err := generateAESGCMKey()
	if err != nil {
		return nil, err
	}
	var idBuf [4]byte
	if _, err := rand.Read(idBuf[:]); err != nil {
		return nil, err
	}
	id := binary.BigEndian.Uint32(idBuf[:])
	ks := &keyset{Keys: []keysetKey{{
		ID:        id,
		Algorithm: alg,
		Key:       base64.StdEncoding.EncodeToString(key),
		Primary:   true,
	}}}
	out, err := serializeKeyset(ks)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(out), nil
}

// BindKeysAddKeyFromRawBytes wraps externally-provided key material.
func BindKeysAddKeyFromRawBytes(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("KEYS.ADD_KEY_FROM_RAW_BYTES: invalid number of arguments: got %d, want 3", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	ksBytes, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	alg, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	raw, err := args[2].ToBytes()
	if err != nil {
		return nil, err
	}
	var ks keyset
	if len(ksBytes) > 0 {
		if err := json.Unmarshal(ksBytes, &ks); err != nil {
			return nil, err
		}
	}
	for i := range ks.Keys {
		ks.Keys[i].Primary = false
	}
	var idBuf [4]byte
	if _, err := rand.Read(idBuf[:]); err != nil {
		return nil, err
	}
	id := binary.BigEndian.Uint32(idBuf[:])
	ks.Keys = append(ks.Keys, keysetKey{
		ID:        id,
		Algorithm: alg,
		Key:       base64.StdEncoding.EncodeToString(raw),
		Primary:   true,
	})
	out, err := serializeKeyset(&ks)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(out), nil
}

// BindKeysKeysetLength returns the number of keys in the keyset.
func BindKeysKeysetLength(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("KEYS.KEYSET_LENGTH: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := keysetFromArg(args[0])
	if err != nil {
		return nil, err
	}
	ks, err := loadKeyset(b)
	if err != nil {
		return nil, err
	}
	return value.IntValue(int64(len(ks.Keys))), nil
}

// BindKeysKeysetToJson returns the keyset as a JSON STRING (the
// keyset's BYTES form is already JSON, so this is an identity).
func BindKeysKeysetToJson(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("KEYS.KEYSET_TO_JSON: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := keysetFromArg(args[0])
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

// BindKeysKeysetFromJson parses a keyset JSON STRING back to BYTES.
func BindKeysKeysetFromJson(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("KEYS.KEYSET_FROM_JSON: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return value.BytesValue([]byte(s)), nil
}

// BindKeysRotateKeyset adds a fresh primary key to the keyset.
func BindKeysRotateKeyset(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("KEYS.ROTATE_KEYSET: missing argument")
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	b, err := keysetFromArg(args[0])
	if err != nil {
		return nil, err
	}
	ks, err := loadKeyset(b)
	if err != nil {
		return nil, err
	}
	alg := algAESGCM
	if len(args) >= 2 && args[1] != nil {
		if s, err := args[1].ToString(); err == nil {
			alg = s
		}
	}
	for i := range ks.Keys {
		ks.Keys[i].Primary = false
	}
	key, err := generateAESGCMKey()
	if err != nil {
		return nil, err
	}
	var idBuf [4]byte
	if _, err := rand.Read(idBuf[:]); err != nil {
		return nil, err
	}
	id := binary.BigEndian.Uint32(idBuf[:])
	ks.Keys = append(ks.Keys, keysetKey{
		ID:        id,
		Algorithm: alg,
		Key:       base64.StdEncoding.EncodeToString(key),
		Primary:   true,
	})
	out, err := serializeKeyset(ks)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(out), nil
}

// BindKeysKeysetChain resolves a wrapped-keyset chain into a STRUCT
// the AEAD.* functions can accept. Without Cloud KMS we treat the
// wrapped material as a passthrough.
func BindKeysKeysetChain(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("KEYS.KEYSET_CHAIN: missing argument")
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	wrapped, err := args[1].ToBytes()
	if err != nil {
		return nil, err
	}
	return &value.StructValue{
		Keys:   []string{"keyset"},
		Values: []value.Value{value.BytesValue(wrapped)},
	}, nil
}

// BindKeysNewWrappedKeyset is a no-op stub: without a KMS wrapper we
// return the same keyset.
func BindKeysNewWrappedKeyset(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("KEYS.NEW_WRAPPED_KEYSET: missing argument")
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	alg, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	return BindKeysNewKeyset(value.StringValue(alg))
}

// BindKeysRewrapKeyset is similarly a no-op without a KMS layer.
func BindKeysRewrapKeyset(args ...value.Value) (value.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("KEYS.REWRAP_KEYSET: missing argument")
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	return args[2], nil
}

// BindKeysRotateWrappedKeyset is similarly a passthrough rotation.
func BindKeysRotateWrappedKeyset(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("KEYS.ROTATE_WRAPPED_KEYSET: missing argument")
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	return args[1], nil
}

// BindAeadEncrypt implements AEAD.ENCRYPT.
func BindAeadEncrypt(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("AEAD.ENCRYPT: invalid number of arguments: got %d, want 3", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	ksBytes, err := keysetFromArg(args[0])
	if err != nil {
		return nil, err
	}
	ks, err := loadKeyset(ksBytes)
	if err != nil {
		return nil, err
	}
	primary := ks.primary()
	if primary.Algorithm != algAESGCM {
		return nil, fmt.Errorf("AEAD.ENCRYPT: primary key is %s, not %s", primary.Algorithm, algAESGCM)
	}
	key, err := decodeKey(primary)
	if err != nil {
		return nil, err
	}
	plaintext, err := plaintextFromArg(args[1])
	if err != nil {
		return nil, err
	}
	aad, err := plaintextFromArg(args[2])
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcmNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	out := make([]byte, 0, tinkPrefixSize+gcmNonceSize+len(plaintext)+gcm.Overhead())
	var prefix [tinkPrefixSize]byte
	binary.BigEndian.PutUint32(prefix[:], primary.ID)
	out = append(out, prefix[:]...)
	out = append(out, nonce...)
	out = gcm.Seal(out, nonce, plaintext, aad)
	return value.BytesValue(out), nil
}

// BindDeterministicEncrypt implements the SIV-style determinstic
// encrypt: same plaintext + AAD → same ciphertext.
func BindDeterministicEncrypt(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("DETERMINISTIC_ENCRYPT: invalid number of arguments: got %d, want 3", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	ksBytes, err := keysetFromArg(args[0])
	if err != nil {
		return nil, err
	}
	ks, err := loadKeyset(ksBytes)
	if err != nil {
		return nil, err
	}
	primary := ks.primary()
	key, err := decodeKey(primary)
	if err != nil {
		return nil, err
	}
	plaintext, err := plaintextFromArg(args[1])
	if err != nil {
		return nil, err
	}
	aad, err := plaintextFromArg(args[2])
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(aad)
	_, _ = mac.Write(plaintext)
	digest := mac.Sum(nil)
	nonce := digest[:gcmNonceSize]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, tinkPrefixSize+gcmNonceSize+len(plaintext)+gcm.Overhead())
	var prefix [tinkPrefixSize]byte
	binary.BigEndian.PutUint32(prefix[:], primary.ID)
	out = append(out, prefix[:]...)
	out = append(out, nonce...)
	out = gcm.Seal(out, nonce, plaintext, aad)
	return value.BytesValue(out), nil
}

// decryptCommon factors the shared AEAD.DECRYPT_BYTES /
// AEAD.DECRYPT_STRING / DETERMINISTIC_DECRYPT_* flow.
func decryptCommon(name string, args []value.Value) ([]byte, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("%s: invalid number of arguments: got %d, want 3", name, len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	ksBytes, err := keysetFromArg(args[0])
	if err != nil {
		return nil, err
	}
	ks, err := loadKeyset(ksBytes)
	if err != nil {
		return nil, err
	}
	ct, err := args[1].ToBytes()
	if err != nil {
		return nil, err
	}
	if len(ct) < tinkPrefixSize+gcmNonceSize {
		return nil, fmt.Errorf("%s: ciphertext too short", name)
	}
	id := binary.BigEndian.Uint32(ct[:tinkPrefixSize])
	k := ks.findID(id)
	if k == nil {
		k = ks.primary()
	}
	key, err := decodeKey(k)
	if err != nil {
		return nil, err
	}
	aad, err := plaintextFromArg(args[2])
	if err != nil {
		return nil, err
	}
	nonce := ct[tinkPrefixSize : tinkPrefixSize+gcmNonceSize]
	body := ct[tinkPrefixSize+gcmNonceSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	out, err := gcm.Open(nil, nonce, body, aad)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}

// BindAeadDecryptBytes implements AEAD.DECRYPT_BYTES.
func BindAeadDecryptBytes(args ...value.Value) (value.Value, error) {
	out, err := decryptCommon("AEAD.DECRYPT_BYTES", args)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return value.BytesValue(out), nil
}

// BindAeadDecryptString implements AEAD.DECRYPT_STRING.
func BindAeadDecryptString(args ...value.Value) (value.Value, error) {
	out, err := decryptCommon("AEAD.DECRYPT_STRING", args)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return value.StringValue(string(out)), nil
}

// BindDeterministicDecryptBytes implements DETERMINISTIC_DECRYPT_BYTES.
func BindDeterministicDecryptBytes(args ...value.Value) (value.Value, error) {
	out, err := decryptCommon("DETERMINISTIC_DECRYPT_BYTES", args)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return value.BytesValue(out), nil
}

// BindDeterministicDecryptString implements DETERMINISTIC_DECRYPT_STRING.
func BindDeterministicDecryptString(args ...value.Value) (value.Value, error) {
	out, err := decryptCommon("DETERMINISTIC_DECRYPT_STRING", args)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return value.StringValue(string(out)), nil
}
