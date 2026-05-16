package extras

import (
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// --- DLP_KEY_CHAIN ---

func TestBindDlpKeyChain(t *testing.T) {
	got, err := BindDlpKeyChain(value.StringValue("projects/p/locations/l/keyRings/r/cryptoKeys/k"), value.BytesValue([]byte("wrapped-bytes")))
	if err != nil {
		t.Fatalf("BindDlpKeyChain: %v", err)
	}
	b, _ := got.ToBytes()
	if !strings.Contains(string(b), "projects/p") {
		t.Fatalf("envelope missing resource: %s", string(b))
	}

	got, _ = BindDlpKeyChain(nil, value.BytesValue(nil))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDlpKeyChain(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- DLP_DETERMINISTIC_ENCRYPT / DECRYPT ---

func TestBindDlpDeterministicEncryptDecrypt(t *testing.T) {
	keyEnv, _ := BindDlpKeyChain(value.StringValue("rsrc"), value.BytesValue([]byte("k")))
	pt := value.StringValue("secret")
	surrogate := value.StringValue("PERSON")

	ct, err := BindDlpDeterministicEncrypt(keyEnv, pt, surrogate)
	if err != nil {
		t.Fatalf("BindDlpDeterministicEncrypt: %v", err)
	}
	// Determinism: two calls with same inputs produce same output.
	ct2, _ := BindDlpDeterministicEncrypt(keyEnv, pt, surrogate)
	cs, _ := ct.ToString()
	cs2, _ := ct2.ToString()
	if cs != cs2 {
		t.Fatalf("deterministic encrypt unstable")
	}

	// Decrypt round-trip.
	out, err := BindDlpDeterministicDecrypt(keyEnv, ct, surrogate)
	if err != nil {
		t.Fatalf("BindDlpDeterministicDecrypt: %v", err)
	}
	s, _ := out.ToString()
	if s != "secret" {
		t.Fatalf("decrypt = %q, want 'secret'", s)
	}
}

func TestBindDlpDeterministicEncryptEmptySurrogate(t *testing.T) {
	keyEnv, _ := BindDlpKeyChain(value.StringValue("rsrc"), value.BytesValue([]byte("k")))
	ct, err := BindDlpDeterministicEncrypt(keyEnv, value.StringValue("data"), value.StringValue(""))
	if err != nil {
		t.Fatalf("BindDlpDeterministicEncrypt: %v", err)
	}
	cs, _ := ct.ToString()
	if strings.Contains(cs, "(") {
		t.Fatalf("empty surrogate must not produce a prefix: %q", cs)
	}
	out, _ := BindDlpDeterministicDecrypt(keyEnv, ct, value.StringValue(""))
	s, _ := out.ToString()
	if s != "data" {
		t.Fatalf("decrypt = %q, want 'data'", s)
	}
}

func TestBindDlpDeterministicEncryptContext(t *testing.T) {
	keyEnv, _ := BindDlpKeyChain(value.StringValue("rsrc"), value.BytesValue([]byte("k")))
	a, _ := BindDlpDeterministicEncrypt(keyEnv, value.StringValue("x"), value.StringValue("S"), value.StringValue("ctx1"))
	b, _ := BindDlpDeterministicEncrypt(keyEnv, value.StringValue("x"), value.StringValue("S"), value.StringValue("ctx2"))
	as, _ := a.ToString()
	bs, _ := b.ToString()
	if as == bs {
		t.Fatalf("different contexts should produce different ciphertext")
	}
}

func TestBindDlpDeterministicDecryptErrors(t *testing.T) {
	if _, err := BindDlpDeterministicDecrypt(value.BytesValue([]byte("k")), value.StringValue("garbage!@#"), value.StringValue("")); err == nil {
		t.Fatalf("invalid base64 should error")
	}
	if _, err := BindDlpDeterministicDecrypt(value.BytesValue([]byte("k")), value.StringValue("aGVsbG8="), value.StringValue("")); err == nil {
		t.Fatalf("too-short ciphertext should error")
	}
}

func TestBindDlpDeterministicEncryptArity(t *testing.T) {
	if _, err := BindDlpDeterministicEncrypt(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindDlpDeterministicDecrypt(); err == nil {
		t.Fatalf("arity error expected")
	}
	got, _ := BindDlpDeterministicEncrypt(nil, value.StringValue("a"), value.StringValue("b"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

// --- OBJ.MAKE_REF / FETCH_METADATA / GET_ACCESS_URL / GET_READ_URL ---

func TestBindObjMakeRefURI(t *testing.T) {
	got, err := BindObjMakeRef(value.StringValue("gs://bucket/path"))
	if err != nil {
		t.Fatalf("BindObjMakeRef: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "gs://bucket/path") {
		t.Fatalf("ref missing uri: %s", s)
	}
}

func TestBindObjMakeRefWithAuthorizer(t *testing.T) {
	got, _ := BindObjMakeRef(value.StringValue("gs://b/p"), value.StringValue("svc-acct"))
	s, _ := got.ToString()
	if !strings.Contains(s, "svc-acct") {
		t.Fatalf("ref missing authorizer: %s", s)
	}
}

func TestBindObjMakeRefJSONInput(t *testing.T) {
	got, _ := BindObjMakeRef(value.StringValue(`{"uri":"gs://b/p","extra":1}`))
	s, _ := got.ToString()
	if !strings.Contains(s, "gs://b/p") {
		t.Fatalf("ref missing uri: %s", s)
	}

	got, _ = BindObjMakeRef(value.StringValue(`{"uri":"gs://b/p"}`), value.StringValue("svc"))
	s, _ = got.ToString()
	if !strings.Contains(s, "svc") {
		t.Fatalf("ref missing authorizer: %s", s)
	}
}

func TestBindObjMakeRefArity(t *testing.T) {
	if _, err := BindObjMakeRef(); err == nil {
		t.Fatalf("arity error expected")
	}
	got, _ := BindObjMakeRef(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindObjFetchMetadata(t *testing.T) {
	ref, _ := BindObjMakeRef(value.StringValue("gs://bucket/path"))
	got, err := BindObjFetchMetadata(ref)
	if err != nil {
		t.Fatalf("BindObjFetchMetadata: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "gs://bucket/path") {
		t.Fatalf("metadata missing uri: %s", s)
	}
	if !strings.Contains(s, "md5_hash") {
		t.Fatalf("metadata missing md5_hash: %s", s)
	}

	got, _ = BindObjFetchMetadata(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindObjFetchMetadata(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindObjGetAccessUrlAndReadUrl(t *testing.T) {
	ref, _ := BindObjMakeRef(value.StringValue("gs://bucket/file"))
	got, err := BindObjGetAccessUrl(ref)
	if err != nil {
		t.Fatalf("BindObjGetAccessUrl: %v", err)
	}
	s, _ := got.ToString()
	if !strings.HasPrefix(s, "https://storage.googleapis.local/") {
		t.Fatalf("access url = %q", s)
	}

	got, err = BindObjGetReadUrl(ref)
	if err != nil {
		t.Fatalf("BindObjGetReadUrl: %v", err)
	}
	s2, _ := got.ToString()
	if s != s2 {
		t.Fatalf("read url should equal access url")
	}

	got, _ = BindObjGetAccessUrl(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindObjGetAccessUrl(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- AI.* ---

func TestBindAiIf(t *testing.T) {
	got, err := BindAiIf(value.StringValue("predicate"), value.IntValue(1), value.IntValue(2))
	if err != nil {
		t.Fatalf("BindAiIf: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 2 {
		t.Fatalf("AI.IF stub want else-branch value 2, got %d", n)
	}
	if _, err := BindAiIf(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindAiClassify(t *testing.T) {
	arr := &value.ArrayValue{Values: []value.Value{
		value.StringValue("cat1"), value.StringValue("cat2"),
	}}
	got, err := BindAiClassify(value.StringValue("input"), arr)
	if err != nil {
		t.Fatalf("BindAiClassify: %v", err)
	}
	s, _ := got.ToString()
	if s != "cat1" {
		t.Fatalf("AI.CLASSIFY want 'cat1', got %q", s)
	}

	// Empty array → NULL.
	got, _ = BindAiClassify(value.StringValue("input"), &value.ArrayValue{Values: nil})
	if got != nil {
		t.Fatalf("AI.CLASSIFY empty array want NULL")
	}
	got, _ = BindAiClassify(value.StringValue("input"), nil)
	if got != nil {
		t.Fatalf("AI.CLASSIFY NULL array want NULL")
	}
	// Non-array second arg passes through.
	got, _ = BindAiClassify(value.StringValue("input"), value.IntValue(42))
	n, _ := got.ToInt64()
	if n != 42 {
		t.Fatalf("non-array passthrough want 42, got %d", n)
	}

	if _, err := BindAiClassify(value.StringValue("only")); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindAiScore(t *testing.T) {
	got, err := BindAiScore(value.StringValue("x"))
	if err != nil {
		t.Fatalf("BindAiScore: %v", err)
	}
	f, _ := got.ToFloat64()
	if f != 0.5 {
		t.Fatalf("AI.SCORE stub = %v, want 0.5", f)
	}
	if _, err := BindAiScore(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindMlPredict(t *testing.T) {
	got, err := BindMlPredict(value.StringValue("model"), value.StringValue("features"))
	if err != nil {
		t.Fatalf("BindMlPredict: %v", err)
	}
	s, _ := got.ToString()
	if s != "features" {
		t.Fatalf("ML.PREDICT passthrough want 'features', got %q", s)
	}
	if _, err := BindMlPredict(value.StringValue("only")); err == nil {
		t.Fatalf("arity error expected")
	}
}
