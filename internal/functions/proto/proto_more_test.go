package proto

import (
	"encoding/base64"
	"math"
	"testing"
	gotime "time"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/goccy/googlesqlite/internal/value"
)

// buildVarintField returns wire bytes for one varint field
// (tag, value).
func buildVarintField(tag int, v uint64) []byte {
	out := protowire.AppendTag(nil, protowire.Number(tag), protowire.VarintType)
	return protowire.AppendVarint(out, v)
}

// buildBytesField returns wire bytes for one length-delimited field
// (tag, payload).
func buildBytesField(tag int, payload []byte) []byte {
	out := protowire.AppendTag(nil, protowire.Number(tag), protowire.BytesType)
	out = protowire.AppendVarint(out, uint64(len(payload)))
	out = append(out, payload...)
	return out
}

// --- BindGetProtoField ---

func TestBindGetProtoFieldInt(t *testing.T) {
	raw := buildVarintField(1, 42)
	got, err := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("int64"),
		value.StringValue(""),
	)
	if err != nil {
		t.Fatalf("BindGetProtoField: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 42 {
		t.Fatalf("want 42, got %d", n)
	}
}

func TestBindGetProtoFieldString(t *testing.T) {
	raw := buildBytesField(2, []byte("hello"))
	got, err := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(2),
		value.StringValue("string"),
		value.StringValue(""),
	)
	if err != nil {
		t.Fatalf("BindGetProtoField: %v", err)
	}
	s, _ := got.ToString()
	if s != "hello" {
		t.Fatalf("want 'hello', got %q", s)
	}
}

func TestBindGetProtoFieldBool(t *testing.T) {
	raw := buildVarintField(1, 1)
	got, _ := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("bool"),
		value.StringValue(""),
	)
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want true")
	}
}

func TestBindGetProtoFieldFloat(t *testing.T) {
	raw := protowire.AppendTag(nil, 1, protowire.Fixed32Type)
	raw = protowire.AppendFixed32(raw, math.Float32bits(2.5))
	got, _ := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("float"),
		value.StringValue(""),
	)
	f, _ := got.ToFloat64()
	if f != 2.5 {
		t.Fatalf("want 2.5, got %v", f)
	}
}

func TestBindGetProtoFieldDouble(t *testing.T) {
	raw := protowire.AppendTag(nil, 1, protowire.Fixed64Type)
	raw = protowire.AppendFixed64(raw, math.Float64bits(3.25))
	got, _ := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("double"),
		value.StringValue(""),
	)
	f, _ := got.ToFloat64()
	if f != 3.25 {
		t.Fatalf("want 3.25, got %v", f)
	}
}

func TestBindGetProtoFieldBytes(t *testing.T) {
	raw := buildBytesField(3, []byte{1, 2, 3})
	got, _ := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(3),
		value.StringValue("bytes"),
		value.StringValue(""),
	)
	b, _ := got.ToBytes()
	if len(b) != 3 || b[2] != 3 {
		t.Fatalf("want [1 2 3], got %v", b)
	}
}

func TestBindGetProtoFieldMessage(t *testing.T) {
	inner := buildVarintField(1, 7)
	raw := buildBytesField(2, inner)
	got, _ := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(2),
		value.StringValue("message"),
		value.StringValue(""),
	)
	b, _ := got.ToBytes()
	if len(b) == 0 {
		t.Fatalf("expected nested message bytes")
	}
}

// TestBindGetProtoFieldMissingNoDefault: field absent and no default →
// SQL NULL.
func TestBindGetProtoFieldMissingNoDefault(t *testing.T) {
	raw := buildVarintField(1, 1)
	got, err := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(99),
		value.StringValue("int64"),
		value.StringValue(""),
	)
	if err != nil {
		t.Fatalf("BindGetProtoField: %v", err)
	}
	if got != nil {
		t.Fatalf("want NULL for missing field, got %v", got)
	}
}

// TestBindGetProtoFieldMissingWithDefault: field absent and a default
// is supplied → default decoded.
func TestBindGetProtoFieldMissingWithDefault(t *testing.T) {
	raw := buildVarintField(1, 1)
	defB64 := base64.StdEncoding.EncodeToString([]byte("123"))
	got, _ := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(99),
		value.StringValue("int64"),
		value.StringValue(defB64),
	)
	n, _ := got.ToInt64()
	if n != 123 {
		t.Fatalf("want 123 (from default), got %d", n)
	}
}

// TestBindGetProtoFieldDefaultBoolFloatString cover the remaining
// decodeDefault kinds.
func TestBindGetProtoFieldDefaultBoolFloatString(t *testing.T) {
	raw := buildVarintField(2, 1)

	defTrue := base64.StdEncoding.EncodeToString([]byte("true"))
	got, _ := BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(99),
		value.StringValue("bool"),
		value.StringValue(defTrue),
	)
	if b, _ := got.ToBool(); !b {
		t.Errorf("bool default want true")
	}

	defF := base64.StdEncoding.EncodeToString([]byte("2.5"))
	got, _ = BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(99),
		value.StringValue("float"),
		value.StringValue(defF),
	)
	if f, _ := got.ToFloat64(); f != 2.5 {
		t.Errorf("float default want 2.5, got %v", f)
	}

	defS := base64.StdEncoding.EncodeToString([]byte("abc"))
	got, _ = BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(99),
		value.StringValue("string"),
		value.StringValue(defS),
	)
	if s, _ := got.ToString(); s != "abc" {
		t.Errorf("string default want 'abc', got %q", s)
	}

	got, _ = BindGetProtoField(
		value.BytesValue(raw),
		value.IntValue(99),
		value.StringValue("bytes"),
		value.StringValue(defS),
	)
	if b, _ := got.ToBytes(); string(b) != "abc" {
		t.Errorf("bytes default want 'abc'")
	}
}

func TestBindGetProtoFieldNullInput(t *testing.T) {
	got, err := BindGetProtoField(nil, value.IntValue(1), value.StringValue("int64"), value.StringValue(""))
	if err != nil {
		t.Fatalf("BindGetProtoField NULL: %v", err)
	}
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindGetProtoFieldArity(t *testing.T) {
	if _, err := BindGetProtoField(value.IntValue(1)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindGetProtoFieldEmptyMessage(t *testing.T) {
	got, err := BindGetProtoField(
		value.BytesValue([]byte{}),
		value.IntValue(1),
		value.StringValue("int64"),
		value.StringValue(""),
	)
	if err != nil {
		t.Fatalf("BindGetProtoField empty: %v", err)
	}
	if got != nil {
		t.Fatalf("empty proto with no default must be NULL")
	}
}

// --- BindGetProtoFieldRepeated ---

func TestBindGetProtoFieldRepeatedVarint(t *testing.T) {
	raw := append(buildVarintField(1, 10), buildVarintField(1, 20)...)
	got, err := BindGetProtoFieldRepeated(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("int64"),
	)
	if err != nil {
		t.Fatalf("BindGetProtoFieldRepeated: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 values, got %d", len(arr.Values))
	}
	n1, _ := arr.Values[0].ToInt64()
	n2, _ := arr.Values[1].ToInt64()
	if n1 != 10 || n2 != 20 {
		t.Fatalf("want [10,20], got [%d,%d]", n1, n2)
	}
}

func TestBindGetProtoFieldRepeatedString(t *testing.T) {
	raw := append(buildBytesField(2, []byte("a")), buildBytesField(2, []byte("b"))...)
	got, _ := BindGetProtoFieldRepeated(
		value.BytesValue(raw),
		value.IntValue(2),
		value.StringValue("string"),
	)
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2, got %d", len(arr.Values))
	}
	s, _ := arr.Values[0].ToString()
	if s != "a" {
		t.Fatalf("first = %q", s)
	}
}

func TestBindGetProtoFieldRepeatedFloat(t *testing.T) {
	raw := protowire.AppendTag(nil, 1, protowire.Fixed32Type)
	raw = protowire.AppendFixed32(raw, math.Float32bits(1.5))
	got, _ := BindGetProtoFieldRepeated(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("float"),
	)
	arr, _ := got.ToArray()
	if len(arr.Values) != 1 {
		t.Fatalf("want 1, got %d", len(arr.Values))
	}
	f, _ := arr.Values[0].ToFloat64()
	if f != 1.5 {
		t.Fatalf("want 1.5, got %v", f)
	}
}

func TestBindGetProtoFieldRepeatedDouble(t *testing.T) {
	raw := protowire.AppendTag(nil, 1, protowire.Fixed64Type)
	raw = protowire.AppendFixed64(raw, math.Float64bits(2.25))
	got, _ := BindGetProtoFieldRepeated(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("double"),
	)
	arr, _ := got.ToArray()
	f, _ := arr.Values[0].ToFloat64()
	if f != 2.25 {
		t.Fatalf("want 2.25, got %v", f)
	}
}

func TestBindGetProtoFieldRepeatedBool(t *testing.T) {
	raw := append(buildVarintField(1, 1), buildVarintField(1, 0)...)
	got, _ := BindGetProtoFieldRepeated(
		value.BytesValue(raw),
		value.IntValue(1),
		value.StringValue("bool"),
	)
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2, got %d", len(arr.Values))
	}
	b0, _ := arr.Values[0].ToBool()
	b1, _ := arr.Values[1].ToBool()
	if !b0 || b1 {
		t.Fatalf("want [true,false], got [%v,%v]", b0, b1)
	}
}

func TestBindGetProtoFieldRepeatedNullArgs(t *testing.T) {
	got, _ := BindGetProtoFieldRepeated(nil, value.IntValue(1), value.StringValue("int64"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindGetProtoFieldRepeated(value.BytesValue(nil), value.IntValue(1)); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindFromProto ---

func TestBindFromProtoInt32Wrapper(t *testing.T) {
	// Int32Value wrapper: field 1, varint
	raw := buildVarintField(1, 42)
	got, err := BindFromProto(value.BytesValue(raw), value.StringValue("int32"))
	if err != nil {
		t.Fatalf("BindFromProto: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 42 {
		t.Fatalf("want 42, got %d", n)
	}
}

func TestBindFromProtoStringWrapper(t *testing.T) {
	raw := buildBytesField(1, []byte("foo"))
	got, _ := BindFromProto(value.BytesValue(raw), value.StringValue("string"))
	s, _ := got.ToString()
	if s != "foo" {
		t.Fatalf("want 'foo', got %q", s)
	}
}

func TestBindFromProtoTimestamp(t *testing.T) {
	// google.protobuf.Timestamp { seconds = 100, nanos = 0 }
	raw := buildVarintField(1, 100)
	got, err := BindFromProto(value.BytesValue(raw), value.StringValue("timestamp"))
	if err != nil {
		t.Fatalf("BindFromProto: %v", err)
	}
	tv, ok := got.(value.TimestampValue)
	if !ok {
		t.Fatalf("want TimestampValue, got %T", got)
	}
	if gotime.Time(tv).Unix() != 100 {
		t.Fatalf("want unix 100, got %d", gotime.Time(tv).Unix())
	}
}

func TestBindFromProtoDate(t *testing.T) {
	// year=2024, month=1, day=15
	raw := append(buildVarintField(1, 2024), buildVarintField(2, 1)...)
	raw = append(raw, buildVarintField(3, 15)...)
	got, err := BindFromProto(value.BytesValue(raw), value.StringValue("date"))
	if err != nil {
		t.Fatalf("BindFromProto: %v", err)
	}
	dv, ok := got.(value.DateValue)
	if !ok {
		t.Fatalf("want DateValue, got %T", got)
	}
	tt := gotime.Time(dv)
	if tt.Year() != 2024 || tt.Month() != 1 || tt.Day() != 15 {
		t.Fatalf("want 2024-01-15, got %v", tt)
	}
}

// TestBindFromProtoIdentityPassthroughDate: when arg is already a
// DateValue (analyzer surfaced it without wrapping), FROM_PROTO must
// return it unchanged.
func TestBindFromProtoIdentityPassthroughDate(t *testing.T) {
	d := value.DateValue(gotime.Date(2020, 5, 1, 0, 0, 0, 0, gotime.UTC))
	got, err := BindFromProto(d, value.StringValue("date"))
	if err != nil {
		t.Fatalf("BindFromProto identity: %v", err)
	}
	if got != d {
		t.Fatalf("identity passthrough failed: got %v, want %v", got, d)
	}
}

func TestBindFromProtoIdentityPassthroughInt(t *testing.T) {
	got, _ := BindFromProto(value.IntValue(5), value.StringValue("int64"))
	n, _ := got.ToInt64()
	if n != 5 {
		t.Fatalf("identity int passthrough failed: %d", n)
	}
}

func TestBindFromProtoEmptyWrapperDefault(t *testing.T) {
	// Empty wrapper proto → zero value for the kind.
	cases := []struct {
		kind string
		want value.Value
	}{
		{"int64", value.IntValue(0)},
		{"int32", value.IntValue(0)},
		{"uint32", value.IntValue(0)},
		{"uint64", value.IntValue(0)},
		{"enum", value.IntValue(0)},
		{"bool", value.BoolValue(false)},
		{"float", value.FloatValue(0)},
		{"double", value.FloatValue(0)},
		{"string", value.StringValue("")},
		{"bytes", value.BytesValue(nil)},
		{"message", value.BytesValue(nil)},
	}
	for _, tc := range cases {
		got, _ := BindFromProto(value.BytesValue([]byte{}), value.StringValue(tc.kind))
		eq, _ := tc.want.EQ(got)
		if !eq {
			t.Errorf("empty wrapper (%s) = %v, want %v", tc.kind, got, tc.want)
		}
	}
}

func TestBindFromProtoNullAndArity(t *testing.T) {
	got, _ := BindFromProto(nil, value.StringValue("int64"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindFromProto(value.BytesValue(nil)); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindToProto ---

func TestBindToProtoInt64(t *testing.T) {
	got, err := BindToProto(value.IntValue(7), value.StringValue("int64"))
	if err != nil {
		t.Fatalf("BindToProto: %v", err)
	}
	b, _ := got.ToBytes()
	if len(b) == 0 {
		t.Fatalf("expected encoded wire bytes")
	}
	// Round-trip
	rt, _ := BindFromProto(value.BytesValue(b), value.StringValue("int64"))
	n, _ := rt.ToInt64()
	if n != 7 {
		t.Fatalf("round-trip want 7, got %d", n)
	}
}

func TestBindToProtoString(t *testing.T) {
	got, _ := BindToProto(value.StringValue("hello"), value.StringValue("string"))
	b, _ := got.ToBytes()
	if len(b) == 0 {
		t.Fatalf("expected encoded wire bytes")
	}
	rt, _ := BindFromProto(value.BytesValue(b), value.StringValue("string"))
	s, _ := rt.ToString()
	if s != "hello" {
		t.Fatalf("round-trip want 'hello', got %q", s)
	}
}

func TestBindToProtoTimestamp(t *testing.T) {
	now := gotime.Date(2024, 1, 1, 0, 0, 0, 5, gotime.UTC)
	got, err := BindToProto(value.TimestampValue(now), value.StringValue("timestamp"))
	if err != nil {
		t.Fatalf("BindToProto: %v", err)
	}
	rt, _ := BindFromProto(got, value.StringValue("timestamp"))
	tv, ok := rt.(value.TimestampValue)
	if !ok {
		t.Fatalf("want TimestampValue, got %T", rt)
	}
	if !gotime.Time(tv).Equal(now) {
		t.Fatalf("round-trip want %v, got %v", now, gotime.Time(tv))
	}
}

func TestBindToProtoDate(t *testing.T) {
	d := gotime.Date(2024, 3, 4, 0, 0, 0, 0, gotime.UTC)
	got, _ := BindToProto(value.DateValue(d), value.StringValue("date"))
	rt, _ := BindFromProto(got, value.StringValue("date"))
	dv, ok := rt.(value.DateValue)
	if !ok {
		t.Fatalf("want DateValue, got %T", rt)
	}
	tt := gotime.Time(dv)
	if tt.Year() != 2024 || tt.Month() != 3 || tt.Day() != 4 {
		t.Fatalf("round-trip want 2024-03-04, got %v", tt)
	}
}

// TestBindToProtoMessageIdentity: existing wire bytes passed in as
// BytesValue with kind=message must round-trip unchanged.
func TestBindToProtoMessageIdentity(t *testing.T) {
	raw := buildVarintField(1, 5)
	got, _ := BindToProto(value.BytesValue(raw), value.StringValue("message"))
	b, _ := got.ToBytes()
	if string(b) != string(raw) {
		t.Fatalf("identity passthrough failed")
	}
}

func TestBindToProtoNullAndArity(t *testing.T) {
	got, _ := BindToProto(nil, value.StringValue("int64"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindToProto(value.IntValue(1)); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindMakeProto ---

func TestBindMakeProtoSimple(t *testing.T) {
	got, err := BindMakeProto(
		value.IntValue(1), value.StringValue("int64"), value.IntValue(99),
		value.IntValue(2), value.StringValue("string"), value.StringValue("hi"),
	)
	if err != nil {
		t.Fatalf("BindMakeProto: %v", err)
	}
	b, _ := got.ToBytes()
	// Decode field 1 via GetProtoField.
	v1, _ := BindGetProtoField(value.BytesValue(b), value.IntValue(1), value.StringValue("int64"), value.StringValue(""))
	n, _ := v1.ToInt64()
	if n != 99 {
		t.Fatalf("field 1 = %d, want 99", n)
	}
	v2, _ := BindGetProtoField(value.BytesValue(b), value.IntValue(2), value.StringValue("string"), value.StringValue(""))
	s, _ := v2.ToString()
	if s != "hi" {
		t.Fatalf("field 2 = %q, want 'hi'", s)
	}
}

func TestBindMakeProtoEmpty(t *testing.T) {
	got, err := BindMakeProto()
	if err != nil {
		t.Fatalf("BindMakeProto empty: %v", err)
	}
	b, _ := got.ToBytes()
	if len(b) != 0 {
		t.Fatalf("want empty bytes, got %v", b)
	}
}

func TestBindMakeProtoErrors(t *testing.T) {
	if _, err := BindMakeProto(value.IntValue(1), value.StringValue("int64")); err == nil {
		t.Fatalf("non-triple should error")
	}
	if _, err := BindMakeProto(nil, value.StringValue("int64"), value.IntValue(1)); err == nil {
		t.Fatalf("NULL tag should error")
	}
	if _, err := BindMakeProto(value.IntValue(1), nil, value.IntValue(1)); err == nil {
		t.Fatalf("NULL kind should error")
	}
}

// --- BindProtoMapContainsKey ---

// TestBindProtoModifyMapDelete inserts then deletes the matching key.
// Driven via the existing entry (key="k1") + a NULL value op.
func TestBindProtoModifyMapDelete(t *testing.T) {
	entry := append(buildBytesField(1, []byte("k1")), buildBytesField(2, []byte("v1"))...)
	parent := buildBytesField(7, entry)
	got, err := BindProtoModifyMap(
		value.BytesValue(parent),
		value.IntValue(7),
		value.StringValue("string"),
		value.StringValue("string"),
		value.StringValue("k1"),
		nil, // delete
	)
	if err != nil {
		t.Fatalf("BindProtoModifyMap delete: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 0 {
		t.Fatalf("after delete want 0 entries, got %d", len(arr.Values))
	}
}

func TestBindProtoModifyMapDuplicateKey(t *testing.T) {
	_, err := BindProtoModifyMap(
		value.BytesValue(nil),
		value.IntValue(7),
		value.StringValue("string"),
		value.StringValue("string"),
		value.StringValue("k1"), value.StringValue("v1"),
		value.StringValue("k1"), value.StringValue("v2"),
	)
	if err == nil {
		t.Fatalf("duplicate key should error")
	}
}

func TestBindProtoModifyMapNullKey(t *testing.T) {
	_, err := BindProtoModifyMap(
		value.BytesValue(nil),
		value.IntValue(7),
		value.StringValue("string"),
		value.StringValue("string"),
		nil, value.StringValue("v1"),
	)
	if err == nil {
		t.Fatalf("NULL key should error")
	}
}

func TestBindProtoModifyMapNullParent(t *testing.T) {
	got, _ := BindProtoModifyMap(
		nil,
		value.IntValue(7),
		value.StringValue("string"),
		value.StringValue("string"),
		value.StringValue("k"), value.StringValue("v"),
	)
	if got != nil {
		t.Fatalf("NULL parent must produce NULL")
	}
}

func TestBindProtoModifyMapArity(t *testing.T) {
	if _, err := BindProtoModifyMap(value.BytesValue(nil), value.IntValue(1)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindReplaceFieldsNestedSynthesis(t *testing.T) {
	raw := []byte{}
	got, err := BindReplaceFields(
		value.BytesValue(raw),
		value.StringValue("3.2"),
		value.StringValue("int64"),
		value.IntValue(42),
	)
	if err != nil {
		t.Fatalf("BindReplaceFields nested synth: %v", err)
	}
	b, _ := got.ToBytes()
	v3, _ := BindGetProtoField(value.BytesValue(b), value.IntValue(3), value.StringValue("message"), value.StringValue(""))
	innerB, _ := v3.ToBytes()
	v2, _ := BindGetProtoField(value.BytesValue(innerB), value.IntValue(2), value.StringValue("int64"), value.StringValue(""))
	if n, _ := v2.ToInt64(); n != 42 {
		t.Fatalf("nested replaced value = %d, want 42", n)
	}
}

func TestBindProtoMapContainsKey(t *testing.T) {
	// map entry: { key (field 1) = "k1", value (field 2) = "v1" }
	entry := append(buildBytesField(1, []byte("k1")), buildBytesField(2, []byte("v1"))...)
	parent := buildBytesField(7, entry)

	got, err := BindProtoMapContainsKey(
		value.BytesValue(parent),
		value.IntValue(7),
		value.StringValue("string"),
		value.StringValue("k1"),
	)
	if err != nil {
		t.Fatalf("BindProtoMapContainsKey: %v", err)
	}
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want true for matching key")
	}

	got, _ = BindProtoMapContainsKey(
		value.BytesValue(parent),
		value.IntValue(7),
		value.StringValue("string"),
		value.StringValue("missing"),
	)
	b, _ = got.ToBool()
	if b {
		t.Fatalf("want false for non-matching key")
	}
}

func TestBindProtoMapContainsKeyNullAndArity(t *testing.T) {
	got, _ := BindProtoMapContainsKey(nil, value.IntValue(7), value.StringValue("string"), value.StringValue("k"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindProtoMapContainsKey(value.BytesValue(nil), value.IntValue(7)); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindEnumValueDescriptorProto ---

func TestBindEnumValueDescriptorProtoNumberOnly(t *testing.T) {
	got, err := BindEnumValueDescriptorProto(value.IntValue(3), value.StringValue("unknown.Enum"))
	if err != nil {
		t.Fatalf("BindEnumValueDescriptorProto: %v", err)
	}
	b, _ := got.ToBytes()
	if len(b) == 0 {
		t.Fatalf("expected encoded enum descriptor bytes")
	}
	// Walk the result: should contain a field 2 (number) with value 3.
	for len(b) > 0 {
		tag, _, n := protowire.ConsumeTag(b)
		b = b[n:]
		if tag == 2 {
			v, _ := protowire.ConsumeVarint(b)
			if v != 3 {
				t.Fatalf("want number 3, got %d", v)
			}
			return
		}
		n = protowire.ConsumeFieldValue(tag, protowire.VarintType, b)
		b = b[n:]
	}
	t.Fatalf("number field not found")
}

func TestBindEnumValueDescriptorProtoNullAndArity(t *testing.T) {
	got, _ := BindEnumValueDescriptorProto(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindEnumValueDescriptorProto(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// TestSetEnumValueNameLookup installs a custom lookup and verifies
// it is used by BindEnumValueDescriptorProto. Restores the default
// lookup afterwards so other tests are unaffected.
func TestSetEnumValueNameLookup(t *testing.T) {
	prev := lookupEnumValueName
	t.Cleanup(func() { lookupEnumValueName = prev })

	SetEnumValueNameLookup(func(enumFullName string, number int32) string {
		if enumFullName == "pkg.Color" && number == 1 {
			return "RED"
		}
		return ""
	})

	got, err := BindEnumValueDescriptorProto(value.IntValue(1), value.StringValue("pkg.Color"))
	if err != nil {
		t.Fatalf("BindEnumValueDescriptorProto: %v", err)
	}
	b, _ := got.ToBytes()
	// Expect field 1 (name) tag followed by 'RED'.
	tag1, _, n := protowire.ConsumeTag(b)
	if n < 0 || tag1 != 1 {
		t.Fatalf("expected name field tag 1, got %d", tag1)
	}
	payload, _ := protowire.ConsumeBytes(b[n:])
	if string(payload) != "RED" {
		t.Fatalf("want 'RED', got %q", string(payload))
	}

	// nil lookup must not displace the installed one.
	SetEnumValueNameLookup(nil)
}

// --- BindReplaceFields ---

func TestBindReplaceFieldsLeaf(t *testing.T) {
	raw := buildVarintField(2, 5)
	got, err := BindReplaceFields(
		value.BytesValue(raw),
		value.StringValue("2"),
		value.StringValue("int64"),
		value.IntValue(99),
	)
	if err != nil {
		t.Fatalf("BindReplaceFields: %v", err)
	}
	b, _ := got.ToBytes()
	v, _ := BindGetProtoField(value.BytesValue(b), value.IntValue(2), value.StringValue("int64"), value.StringValue(""))
	n, _ := v.ToInt64()
	if n != 99 {
		t.Fatalf("replaced value = %d, want 99", n)
	}
}

// TestBindReplaceFieldsRepeated covers the ArrayValue replacement
// branch (repeated field).
func TestBindReplaceFieldsRepeated(t *testing.T) {
	raw := buildBytesField(3, []byte("old"))
	arr := &value.ArrayValue{Values: []value.Value{value.StringValue("a"), value.StringValue("b")}}
	got, err := BindReplaceFields(
		value.BytesValue(raw),
		value.StringValue("3"),
		value.StringValue("array"),
		arr,
	)
	if err != nil {
		t.Fatalf("BindReplaceFields: %v", err)
	}
	b, _ := got.ToBytes()
	rep, _ := BindGetProtoFieldRepeated(value.BytesValue(b), value.IntValue(3), value.StringValue("string"))
	out, _ := rep.ToArray()
	if len(out.Values) != 2 {
		t.Fatalf("want 2 repeated entries, got %d", len(out.Values))
	}
}

// TestBindReplaceFieldsClearsNull: when val is nil the leaf is cleared.
func TestBindReplaceFieldsClearsNull(t *testing.T) {
	raw := buildVarintField(2, 5)
	got, _ := BindReplaceFields(
		value.BytesValue(raw),
		value.StringValue("2"),
		value.StringValue("int64"),
		nil,
	)
	b, _ := got.ToBytes()
	if len(b) != 0 {
		t.Fatalf("want cleared bytes, got %v", b)
	}
}

func TestBindReplaceFieldsInvalidPath(t *testing.T) {
	raw := buildVarintField(1, 1)
	// Empty path: function returns raw unchanged.
	got, _ := BindReplaceFields(
		value.BytesValue(raw),
		value.StringValue(""),
		value.StringValue("int64"),
		value.IntValue(2),
	)
	b, _ := got.ToBytes()
	if string(b) != string(raw) {
		t.Fatalf("empty path must leave proto unchanged")
	}
}

func TestBindReplaceFieldsNullAndArity(t *testing.T) {
	got, _ := BindReplaceFields(nil, value.StringValue("1"), value.StringValue("int64"), value.IntValue(1))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindReplaceFields(value.BytesValue(nil), value.StringValue("1")); err == nil {
		t.Fatalf("arity error expected")
	}
}

// TestBindFilterFieldsExcludeOnly covers the exclude-only branch
// (no '+' prefix) of BindFilterFields.
func TestBindFilterFieldsExcludeOnly(t *testing.T) {
	raw := append(buildVarintField(1, 1), buildVarintField(2, 2)...)
	got, err := BindFilterFields(
		value.BytesValue(raw),
		value.StringValue("-1"),
	)
	if err != nil {
		t.Fatalf("BindFilterFields: %v", err)
	}
	b, _ := got.ToBytes()
	// field 1 must be gone, field 2 retained.
	if _, err := BindGetProtoField(value.BytesValue(b), value.IntValue(2), value.StringValue("int64"), value.StringValue("")); err != nil {
		t.Fatalf("field 2 lookup: %v", err)
	}
	got2, _ := BindGetProtoField(value.BytesValue(b), value.IntValue(1), value.StringValue("int64"), value.StringValue(""))
	if got2 != nil {
		t.Fatalf("field 1 must be filtered out")
	}
}

// TestEncodePayloadRoundtripAllKinds round-trips the supported
// kinds through MakeProto + FromProto so the encoder/decoder branches
// for each scalar kind are exercised.
func TestEncodePayloadRoundtripAllKinds(t *testing.T) {
	cases := []struct {
		kind string
		in   value.Value
	}{
		{"bool", value.BoolValue(true)},
		{"int32", value.IntValue(123)},
		{"int64", value.IntValue(-7)},
		{"uint32", value.IntValue(1024)},
		{"uint64", value.IntValue(9999)},
		{"float", value.FloatValue(1.5)},
		{"double", value.FloatValue(3.14)},
		{"string", value.StringValue("hi")},
		{"bytes", value.BytesValue([]byte{9, 8, 7})},
		{"enum", value.IntValue(3)},
	}
	for _, tc := range cases {
		got, err := BindMakeProto(value.IntValue(1), value.StringValue(tc.kind), tc.in)
		if err != nil {
			t.Errorf("MAKE_PROTO(%s): %v", tc.kind, err)
			continue
		}
		b, _ := got.ToBytes()
		rt, err := BindFromProto(value.BytesValue(b), value.StringValue(tc.kind))
		if err != nil {
			t.Errorf("FROM_PROTO(%s): %v", tc.kind, err)
			continue
		}
		// Compare via EQ where supported.
		eq, _ := tc.in.EQ(rt)
		if !eq {
			t.Errorf("round-trip(%s) mismatch: in=%v out=%v", tc.kind, tc.in, rt)
		}
	}
}

// TestBindFilterFieldsIncludeWithNestedExcludes covers the
// applyFilterTreeIncludeWithExcludes branch.
func TestBindFilterFieldsIncludeWithNestedExcludes(t *testing.T) {
	inner := append(buildVarintField(1, 1), buildVarintField(2, 2)...)
	raw := buildBytesField(3, inner)
	got, err := BindFilterFields(
		value.BytesValue(raw),
		value.StringValue("+3,-3.2"),
	)
	if err != nil {
		t.Fatalf("BindFilterFields: %v", err)
	}
	b, _ := got.ToBytes()
	// field 3.1 must survive, field 3.2 must be excluded.
	msg, _ := BindGetProtoField(value.BytesValue(b), value.IntValue(3), value.StringValue("message"), value.StringValue(""))
	inB, _ := msg.ToBytes()
	v1, _ := BindGetProtoField(value.BytesValue(inB), value.IntValue(1), value.StringValue("int64"), value.StringValue(""))
	if n, _ := v1.ToInt64(); n != 1 {
		t.Fatalf("inner field 1 = %d, want 1", n)
	}
	v2, _ := BindGetProtoField(value.BytesValue(inB), value.IntValue(2), value.StringValue("int64"), value.StringValue(""))
	if v2 != nil {
		t.Fatalf("inner field 2 must be excluded, got %v", v2)
	}
}

// TestBindFilterFieldsResetRequired exercises ensureRequiredFieldDefaults
// via a stub RequiredFieldResolver.
func TestBindFilterFieldsResetRequired(t *testing.T) {
	prev := RequiredFieldResolver
	t.Cleanup(func() { RequiredFieldResolver = prev })

	RequiredFieldResolver = func(fullName string) []RequiredField {
		if fullName != "pkg.Msg" {
			return nil
		}
		return []RequiredField{{
			Number:  5,
			Wire:    uint8(protowire.VarintType),
			Payload: protowire.AppendVarint(nil, 7),
		}}
	}

	raw := buildVarintField(1, 1)
	got, err := BindFilterFields(
		value.BytesValue(raw),
		value.StringValue("+1"),
		value.BoolValue(true),
		value.StringValue("pkg.Msg"),
	)
	if err != nil {
		t.Fatalf("BindFilterFields: %v", err)
	}
	b, _ := got.ToBytes()
	// Required field 5 should have been appended.
	v5, _ := BindGetProtoField(value.BytesValue(b), value.IntValue(5), value.StringValue("int64"), value.StringValue(""))
	if n, _ := v5.ToInt64(); n != 7 {
		t.Fatalf("required default field 5 = %d, want 7", n)
	}
}

func TestBindFilterFieldsNullAndArity(t *testing.T) {
	got, _ := BindFilterFields(nil, value.StringValue("+1"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindFilterFields(value.BytesValue(nil)); err == nil {
		t.Fatalf("arity error expected")
	}
}
