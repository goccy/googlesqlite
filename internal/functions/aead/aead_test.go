package aead

import (
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// newKeyset builds a fresh AES-GCM keyset and returns the bytes form.
// While the key material is non-deterministic, every subsequent step
// in the test path round-trips through the same keyset, so the test
// outcome is deterministic.
func newKeyset(t *testing.T) value.Value {
	t.Helper()
	got, err := BindKeysNewKeyset(value.StringValue("AEAD_AES_GCM_256"))
	if err != nil {
		t.Fatalf("BindKeysNewKeyset: %v", err)
	}
	return got
}

// --- BindKeysNewKeyset ---

func TestBindKeysNewKeyset(t *testing.T) {
	got, err := BindKeysNewKeyset(value.StringValue("AEAD_AES_GCM_256"))
	if err != nil {
		t.Fatalf("BindKeysNewKeyset: %v", err)
	}
	b, _ := got.ToBytes()
	if len(b) == 0 {
		t.Fatalf("expected keyset bytes")
	}
	if !strings.Contains(string(b), "AEAD_AES_GCM_256") {
		t.Fatalf("keyset missing algorithm marker: %s", string(b))
	}
}

func TestBindKeysNewKeysetDeterministicAlg(t *testing.T) {
	got, err := BindKeysNewKeyset(value.StringValue("DETERMINISTIC_AEAD_AES_SIV_CMAC_256"))
	if err != nil {
		t.Fatalf("BindKeysNewKeyset (det): %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil keyset")
	}
}

func TestBindKeysNewKeysetUnsupported(t *testing.T) {
	if _, err := BindKeysNewKeyset(value.StringValue("UNKNOWN_ALG")); err == nil {
		t.Fatalf("unsupported algorithm should error")
	}
}

func TestBindKeysNewKeysetNullAndArity(t *testing.T) {
	got, _ := BindKeysNewKeyset(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysNewKeyset(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindKeysKeysetLength ---

func TestBindKeysKeysetLength(t *testing.T) {
	ks := newKeyset(t)
	got, err := BindKeysKeysetLength(ks)
	if err != nil {
		t.Fatalf("BindKeysKeysetLength: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 1 {
		t.Fatalf("fresh keyset length = %d, want 1", n)
	}
}

func TestBindKeysKeysetLengthNullAndArity(t *testing.T) {
	got, _ := BindKeysKeysetLength(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysKeysetLength(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindKeysKeysetLength(value.BytesValue([]byte("not-json"))); err == nil {
		t.Fatalf("invalid keyset bytes should error")
	}
}

// --- BindKeysKeysetToJson / BindKeysKeysetFromJson ---

func TestBindKeysKeysetJsonRoundtrip(t *testing.T) {
	ks := newKeyset(t)
	jsonForm, err := BindKeysKeysetToJson(ks)
	if err != nil {
		t.Fatalf("BindKeysKeysetToJson: %v", err)
	}
	s, _ := jsonForm.ToString()
	if !strings.Contains(s, "AEAD_AES_GCM_256") {
		t.Fatalf("JSON form missing algorithm: %s", s)
	}
	roundtrip, err := BindKeysKeysetFromJson(jsonForm)
	if err != nil {
		t.Fatalf("BindKeysKeysetFromJson: %v", err)
	}
	rtBytes, _ := roundtrip.ToBytes()
	if string(rtBytes) != s {
		t.Fatalf("FROM_JSON did not round-trip: got %s want %s", string(rtBytes), s)
	}
}

func TestBindKeysKeysetJsonNullAndArity(t *testing.T) {
	got, _ := BindKeysKeysetToJson(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	got, _ = BindKeysKeysetFromJson(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysKeysetToJson(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindKeysKeysetFromJson(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindKeysAddKeyFromRawBytes ---

func TestBindKeysAddKeyFromRawBytes(t *testing.T) {
	ks := newKeyset(t)
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	got, err := BindKeysAddKeyFromRawBytes(
		ks,
		value.StringValue("AEAD_AES_GCM_256"),
		value.BytesValue(raw),
	)
	if err != nil {
		t.Fatalf("BindKeysAddKeyFromRawBytes: %v", err)
	}
	n, _ := BindKeysKeysetLength(got)
	got2, _ := n.ToInt64()
	if got2 != 2 {
		t.Fatalf("after add: length = %d, want 2", got2)
	}
}

func TestBindKeysAddKeyFromRawBytesEmptyKeyset(t *testing.T) {
	// Empty keyset + add → 1 key
	raw := make([]byte, 32)
	got, err := BindKeysAddKeyFromRawBytes(
		value.BytesValue([]byte{}),
		value.StringValue("AEAD_AES_GCM_256"),
		value.BytesValue(raw),
	)
	if err != nil {
		t.Fatalf("BindKeysAddKeyFromRawBytes empty: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil keyset")
	}
}

func TestBindKeysAddKeyFromRawBytesNullAndArity(t *testing.T) {
	got, _ := BindKeysAddKeyFromRawBytes(nil, value.StringValue("AEAD_AES_GCM_256"), value.BytesValue(nil))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysAddKeyFromRawBytes(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindKeysRotateKeyset ---

func TestBindKeysRotateKeyset(t *testing.T) {
	ks := newKeyset(t)
	got, err := BindKeysRotateKeyset(ks)
	if err != nil {
		t.Fatalf("BindKeysRotateKeyset: %v", err)
	}
	n, _ := BindKeysKeysetLength(got)
	v, _ := n.ToInt64()
	if v != 2 {
		t.Fatalf("after rotate: length = %d, want 2", v)
	}
}

func TestBindKeysRotateKeysetCustomAlg(t *testing.T) {
	ks := newKeyset(t)
	got, err := BindKeysRotateKeyset(ks, value.StringValue("AEAD_AES_GCM_256"))
	if err != nil {
		t.Fatalf("BindKeysRotateKeyset alg: %v", err)
	}
	n, _ := BindKeysKeysetLength(got)
	v, _ := n.ToInt64()
	if v != 2 {
		t.Fatalf("after rotate: length = %d, want 2", v)
	}
}

func TestBindKeysRotateKeysetNullAndArity(t *testing.T) {
	got, _ := BindKeysRotateKeyset(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysRotateKeyset(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindAeadEncrypt / BindAeadDecrypt* ---

func TestBindAeadEncryptDecryptString(t *testing.T) {
	ks := newKeyset(t)
	plaintext := value.StringValue("hello world")
	aad := value.StringValue("context")

	ct, err := BindAeadEncrypt(ks, plaintext, aad)
	if err != nil {
		t.Fatalf("BindAeadEncrypt: %v", err)
	}

	pt, err := BindAeadDecryptString(ks, ct, aad)
	if err != nil {
		t.Fatalf("BindAeadDecryptString: %v", err)
	}
	s, _ := pt.ToString()
	if s != "hello world" {
		t.Fatalf("round-trip = %q, want 'hello world'", s)
	}
}

func TestBindAeadEncryptDecryptBytes(t *testing.T) {
	ks := newKeyset(t)
	plaintext := value.BytesValue([]byte{1, 2, 3, 4})
	aad := value.BytesValue(nil)

	ct, err := BindAeadEncrypt(ks, plaintext, aad)
	if err != nil {
		t.Fatalf("BindAeadEncrypt: %v", err)
	}
	pt, err := BindAeadDecryptBytes(ks, ct, aad)
	if err != nil {
		t.Fatalf("BindAeadDecryptBytes: %v", err)
	}
	b, _ := pt.ToBytes()
	if len(b) != 4 || b[3] != 4 {
		t.Fatalf("round-trip = %v, want [1,2,3,4]", b)
	}
}

func TestBindAeadDecryptMismatchedAAD(t *testing.T) {
	ks := newKeyset(t)
	ct, err := BindAeadEncrypt(ks, value.StringValue("secret"), value.StringValue("aad1"))
	if err != nil {
		t.Fatalf("BindAeadEncrypt: %v", err)
	}
	if _, err := BindAeadDecryptString(ks, ct, value.StringValue("aad2")); err == nil {
		t.Fatalf("decrypt with wrong AAD should error")
	}
}

func TestBindAeadEncryptErrors(t *testing.T) {
	if _, err := BindAeadEncrypt(); err == nil {
		t.Fatalf("arity error expected")
	}
	got, _ := BindAeadEncrypt(nil, value.StringValue("a"), value.StringValue("b"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindAeadEncrypt(value.BytesValue([]byte("not-json")), value.StringValue("a"), value.StringValue("b")); err == nil {
		t.Fatalf("invalid keyset should error")
	}
}

func TestBindAeadDecryptShortCiphertext(t *testing.T) {
	ks := newKeyset(t)
	if _, err := BindAeadDecryptBytes(ks, value.BytesValue([]byte{1, 2, 3}), value.StringValue("")); err == nil {
		t.Fatalf("short ciphertext should error")
	}
}

// --- BindDeterministicEncrypt / BindDeterministicDecrypt* ---

func TestBindDeterministicEncryptIsStable(t *testing.T) {
	ks := newKeyset(t)
	plaintext := value.StringValue("payload")
	aad := value.StringValue("aad")

	a, err := BindDeterministicEncrypt(ks, plaintext, aad)
	if err != nil {
		t.Fatalf("BindDeterministicEncrypt: %v", err)
	}
	b, err := BindDeterministicEncrypt(ks, plaintext, aad)
	if err != nil {
		t.Fatalf("BindDeterministicEncrypt: %v", err)
	}
	ab, _ := a.ToBytes()
	bb, _ := b.ToBytes()
	if string(ab) != string(bb) {
		t.Fatalf("deterministic encrypt should be stable across calls")
	}

	// Round-trip via DETERMINISTIC_DECRYPT_STRING.
	pt, err := BindDeterministicDecryptString(ks, a, aad)
	if err != nil {
		t.Fatalf("BindDeterministicDecryptString: %v", err)
	}
	s, _ := pt.ToString()
	if s != "payload" {
		t.Fatalf("round-trip = %q, want 'payload'", s)
	}
}

func TestBindDeterministicDecryptBytes(t *testing.T) {
	ks := newKeyset(t)
	plaintext := value.BytesValue([]byte{9, 8, 7})
	aad := value.BytesValue([]byte("aad"))

	ct, err := BindDeterministicEncrypt(ks, plaintext, aad)
	if err != nil {
		t.Fatalf("BindDeterministicEncrypt: %v", err)
	}
	pt, err := BindDeterministicDecryptBytes(ks, ct, aad)
	if err != nil {
		t.Fatalf("BindDeterministicDecryptBytes: %v", err)
	}
	b, _ := pt.ToBytes()
	if len(b) != 3 || b[2] != 7 {
		t.Fatalf("round-trip = %v, want [9,8,7]", b)
	}
}

func TestBindDeterministicEncryptErrors(t *testing.T) {
	if _, err := BindDeterministicEncrypt(); err == nil {
		t.Fatalf("arity error expected")
	}
	got, _ := BindDeterministicEncrypt(nil, value.StringValue("a"), value.StringValue("b"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

// --- KeysetChain / NewWrappedKeyset / RewrapKeyset / RotateWrappedKeyset ---

func TestBindKeysKeysetChain(t *testing.T) {
	got, err := BindKeysKeysetChain(value.BytesValue([]byte("kms-ref")), value.BytesValue([]byte("wrapped")))
	if err != nil {
		t.Fatalf("BindKeysKeysetChain: %v", err)
	}
	st, _ := got.ToStruct()
	if len(st.Keys) != 1 || st.Keys[0] != "keyset" {
		t.Fatalf("want struct with 'keyset' field, got %v", st.Keys)
	}

	// NULL inputs propagate.
	got, _ = BindKeysKeysetChain(nil, value.BytesValue([]byte("w")))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysKeysetChain(value.BytesValue(nil)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindKeysNewWrappedKeyset(t *testing.T) {
	got, err := BindKeysNewWrappedKeyset(value.BytesValue([]byte("kms")), value.StringValue("AEAD_AES_GCM_256"))
	if err != nil {
		t.Fatalf("BindKeysNewWrappedKeyset: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil keyset")
	}

	got, _ = BindKeysNewWrappedKeyset(nil, value.StringValue("AEAD_AES_GCM_256"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysNewWrappedKeyset(value.BytesValue(nil)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindKeysRewrapKeyset(t *testing.T) {
	want := value.BytesValue([]byte("payload"))
	got, err := BindKeysRewrapKeyset(value.BytesValue([]byte("a")), value.BytesValue([]byte("b")), want)
	if err != nil {
		t.Fatalf("BindKeysRewrapKeyset: %v", err)
	}
	gb, _ := got.ToBytes()
	if string(gb) != "payload" {
		t.Fatalf("expected passthrough, got %q", string(gb))
	}

	got, _ = BindKeysRewrapKeyset(nil, value.BytesValue(nil), value.BytesValue(nil))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysRewrapKeyset(value.BytesValue(nil)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindKeysRotateWrappedKeyset(t *testing.T) {
	want := value.BytesValue([]byte("payload"))
	got, err := BindKeysRotateWrappedKeyset(value.BytesValue([]byte("a")), want)
	if err != nil {
		t.Fatalf("BindKeysRotateWrappedKeyset: %v", err)
	}
	gb, _ := got.ToBytes()
	if string(gb) != "payload" {
		t.Fatalf("expected passthrough, got %q", string(gb))
	}

	got, _ = BindKeysRotateWrappedKeyset(nil, value.BytesValue(nil))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindKeysRotateWrappedKeyset(value.BytesValue(nil)); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- keysetFromArg (STRUCT path) ---

func TestKeysetFromArgStruct(t *testing.T) {
	ks := newKeyset(t)
	b, _ := ks.ToBytes()
	st := &value.StructValue{
		Keys:   []string{"keyset"},
		Values: []value.Value{value.BytesValue(b)},
	}
	got, err := BindKeysKeysetLength(st)
	if err != nil {
		t.Fatalf("BindKeysKeysetLength struct: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 1 {
		t.Fatalf("want 1, got %d", n)
	}
}

func TestKeysetFromArgStructMissingField(t *testing.T) {
	st := &value.StructValue{
		Keys:   []string{"other"},
		Values: []value.Value{value.StringValue("x")},
	}
	if _, err := BindKeysKeysetLength(st); err == nil {
		t.Fatalf("STRUCT without 'keyset' field should error")
	}
}
