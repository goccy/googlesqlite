package longtail

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// --- debug ---

func TestBindErrorRaises(t *testing.T) {
	if _, err := BindError(value.StringValue("boom")); err == nil || err.Error() != "boom" {
		t.Fatalf("BindError want 'boom', got %v", err)
	}
}

func TestBindErrorNullArg(t *testing.T) {
	_, err := BindError(nil)
	if err == nil {
		t.Fatalf("BindError should still error on NULL arg")
	}
	if err.Error() != "user-raised error" {
		t.Fatalf("BindError default message, got %v", err)
	}
}

func TestBindErrorWrongArity(t *testing.T) {
	if _, err := BindError(); err == nil {
		t.Fatalf("BindError(): expected arity error")
	}
}

func TestBindIsErrorEager(t *testing.T) {
	got, err := BindIsError(value.IntValue(1))
	if err != nil {
		t.Fatalf("BindIsError: %v", err)
	}
	b, _ := got.ToBool()
	if b {
		t.Fatalf("eager runtime should report false")
	}
	if _, err := BindIsError(); err == nil {
		t.Fatalf("BindIsError(): expected arity error")
	}
}

func TestBindIfErrorReturnsFirstNonNull(t *testing.T) {
	got, err := BindIfError(value.IntValue(7), value.IntValue(0))
	if err != nil {
		t.Fatalf("BindIfError: %v", err)
	}
	v, _ := got.ToInt64()
	if v != 7 {
		t.Fatalf("want 7, got %d", v)
	}

	got, err = BindIfError(nil, value.IntValue(99))
	if err != nil {
		t.Fatalf("BindIfError: %v", err)
	}
	v, _ = got.ToInt64()
	if v != 99 {
		t.Fatalf("want 99 (fallback), got %d", v)
	}

	if _, err := BindIfError(value.IntValue(1)); err == nil {
		t.Fatalf("BindIfError arity error expected")
	}
}

func TestBindNullIfErrorPassthrough(t *testing.T) {
	got, err := BindNullIfError(value.IntValue(11))
	if err != nil {
		t.Fatalf("BindNullIfError: %v", err)
	}
	v, _ := got.ToInt64()
	if v != 11 {
		t.Fatalf("want 11, got %d", v)
	}
	if _, err := BindNullIfError(); err == nil {
		t.Fatalf("BindNullIfError arity error expected")
	}
}

// --- string ---

func TestBindCollatePassthrough(t *testing.T) {
	got, err := BindCollate(value.StringValue("abc"), value.StringValue("und:ci"))
	if err != nil {
		t.Fatalf("BindCollate: %v", err)
	}
	s, _ := got.ToString()
	if s != "abc" {
		t.Fatalf("want 'abc', got %q", s)
	}
	got, err = BindCollate(value.StringValue("solo"))
	if err != nil {
		t.Fatalf("BindCollate single arg: %v", err)
	}
	s, _ = got.ToString()
	if s != "solo" {
		t.Fatalf("want 'solo', got %q", s)
	}
	if _, err := BindCollate(); err == nil {
		t.Fatalf("BindCollate arity error expected")
	}
	if _, err := BindCollate(value.IntValue(1), value.IntValue(2), value.IntValue(3)); err == nil {
		t.Fatalf("BindCollate 3-arg should fail")
	}
}

func TestBindRegexpMatch(t *testing.T) {
	got, err := BindRegexpMatch(value.StringValue("hello world"), value.StringValue("w.r"))
	if err != nil {
		t.Fatalf("BindRegexpMatch: %v", err)
	}
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("expected match for 'w.r' in 'hello world'")
	}

	got, _ = BindRegexpMatch(value.StringValue("abc"), value.StringValue("xyz"))
	b, _ = got.ToBool()
	if b {
		t.Fatalf("expected no match")
	}

	if got, _ := BindRegexpMatch(nil, value.StringValue("x")); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}

	if _, err := BindRegexpMatch(value.StringValue("a"), value.StringValue("[")); err == nil {
		t.Fatalf("invalid regex should error")
	}

	if _, err := BindRegexpMatch(value.StringValue("a")); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindRegexpExtractGroupsPositional(t *testing.T) {
	got, err := BindRegexpExtractGroups(
		value.StringValue("abc-123"),
		value.StringValue(`([a-z]+)-([0-9]+)`),
	)
	if err != nil {
		t.Fatalf("BindRegexpExtractGroups: %v", err)
	}
	st, err := got.ToStruct()
	if err != nil {
		t.Fatalf("ToStruct: %v", err)
	}
	if len(st.Keys) != 2 {
		t.Fatalf("want 2 fields, got %v", st.Keys)
	}
	s, _ := st.Values[0].ToString()
	if s != "abc" {
		t.Fatalf("first group = %q, want 'abc'", s)
	}
	s, _ = st.Values[1].ToString()
	if s != "123" {
		t.Fatalf("second group = %q, want '123'", s)
	}
}

func TestBindRegexpExtractGroupsNamedAndCoerced(t *testing.T) {
	got, err := BindRegexpExtractGroups(
		value.StringValue("count=42"),
		value.StringValue(`count=(?P<n__INT64>\d+)`),
	)
	if err != nil {
		t.Fatalf("BindRegexpExtractGroups: %v", err)
	}
	st, _ := got.ToStruct()
	if len(st.Keys) != 1 || st.Keys[0] != "n" {
		t.Fatalf("want field 'n', got %v", st.Keys)
	}
	n, err := st.Values[0].ToInt64()
	if err != nil || n != 42 {
		t.Fatalf("coerced int = %d/%v, want 42", n, err)
	}
}

func TestBindRegexpExtractGroupsNoMatch(t *testing.T) {
	got, err := BindRegexpExtractGroups(value.StringValue("foo"), value.StringValue("bar"))
	if err != nil {
		t.Fatalf("BindRegexpExtractGroups: %v", err)
	}
	if got != nil {
		t.Fatalf("want NULL for no match, got %v", got)
	}
}

func TestBindRegexpExtractGroupsFloatBool(t *testing.T) {
	got, err := BindRegexpExtractGroups(
		value.StringValue("rate=3.14 ok=true"),
		value.StringValue(`rate=(?P<r__FLOAT64>[0-9.]+) ok=(?P<ok__BOOL>true|false)`),
	)
	if err != nil {
		t.Fatalf("BindRegexpExtractGroups: %v", err)
	}
	st, _ := got.ToStruct()
	if len(st.Keys) != 2 {
		t.Fatalf("want 2 fields, got %v", st.Keys)
	}
	f, err := st.Values[0].ToFloat64()
	if err != nil || f != 3.14 {
		t.Fatalf("float field = %v / %v, want 3.14", f, err)
	}
	b, err := st.Values[1].ToBool()
	if err != nil || !b {
		t.Fatalf("bool field = %v / %v, want true", b, err)
	}
}

// TestBindRegexpExtractGroupsHex covers the parseInt64Flexible
// hex-prefix path through the __INT64 coercion.
func TestBindRegexpExtractGroupsHex(t *testing.T) {
	got, err := BindRegexpExtractGroups(
		value.StringValue("x=0x1A"),
		value.StringValue(`x=(?P<v__INT64>0x[0-9A-F]+)`),
	)
	if err != nil {
		t.Fatalf("BindRegexpExtractGroups: %v", err)
	}
	st, _ := got.ToStruct()
	n, err := st.Values[0].ToInt64()
	if err != nil || n != 26 {
		t.Fatalf("hex int field = %d / %v, want 26", n, err)
	}
}

// TestBindRegexpExtractGroupsOptionalNonMatch hits the start < 0
// branch (optional group did not participate).
func TestBindRegexpExtractGroupsOptionalNonMatch(t *testing.T) {
	got, err := BindRegexpExtractGroups(
		value.StringValue("foo"),
		value.StringValue(`(foo)(?:(bar))?`),
	)
	if err != nil {
		t.Fatalf("BindRegexpExtractGroups: %v", err)
	}
	st, _ := got.ToStruct()
	if len(st.Values) != 2 {
		t.Fatalf("want 2 fields, got %v", st.Values)
	}
	if st.Values[1] != nil {
		t.Fatalf("optional non-match field = %v, want nil", st.Values[1])
	}
}

// TestBindRegexpExtractGroupsCoerceError: invalid int coercion text
// surfaces an error.
func TestBindRegexpExtractGroupsCoerceError(t *testing.T) {
	_, err := BindRegexpExtractGroups(
		value.StringValue("count=abc"),
		value.StringValue(`count=(?P<n__INT64>\S+)`),
	)
	if err == nil {
		t.Fatalf("expected coercion error")
	}
}

func TestBindRegexpExtractGroupsErrors(t *testing.T) {
	if _, err := BindRegexpExtractGroups(value.StringValue("a")); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindRegexpExtractGroups(value.StringValue("a"), value.StringValue("[")); err == nil {
		t.Fatalf("invalid regex should error")
	}
	if got, _ := BindRegexpExtractGroups(nil, value.StringValue(".")); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindSplitSubstr(t *testing.T) {
	// Positive position (1-based).
	got, err := BindSplitSubstr(
		value.StringValue("a,b,c,d"),
		value.StringValue(","),
		value.IntValue(2),
	)
	if err != nil {
		t.Fatalf("BindSplitSubstr: %v", err)
	}
	s, _ := got.ToString()
	if s != "b" {
		t.Fatalf("position 2 = %q, want 'b'", s)
	}

	// Negative position counts from the end.
	got, _ = BindSplitSubstr(
		value.StringValue("a,b,c,d"),
		value.StringValue(","),
		value.IntValue(-1),
	)
	s, _ = got.ToString()
	if s != "d" {
		t.Fatalf("position -1 = %q, want 'd'", s)
	}

	// Position 0 returns empty.
	got, _ = BindSplitSubstr(
		value.StringValue("a,b,c"),
		value.StringValue(","),
		value.IntValue(0),
	)
	s, _ = got.ToString()
	if s != "" {
		t.Fatalf("position 0 = %q, want empty", s)
	}

	// Count argument joins multiple parts.
	got, _ = BindSplitSubstr(
		value.StringValue("a,b,c,d"),
		value.StringValue(","),
		value.IntValue(2),
		value.IntValue(2),
	)
	s, _ = got.ToString()
	if s != "b,c" {
		t.Fatalf("count=2 = %q, want 'b,c'", s)
	}

	// Index past the end → empty.
	got, _ = BindSplitSubstr(
		value.StringValue("a,b"),
		value.StringValue(","),
		value.IntValue(5),
	)
	s, _ = got.ToString()
	if s != "" {
		t.Fatalf("out-of-range = %q, want empty", s)
	}

	if got, _ := BindSplitSubstr(nil, value.StringValue(","), value.IntValue(1)); got != nil {
		t.Fatalf("NULL input must produce NULL")
	}
	if _, err := BindSplitSubstr(value.StringValue("a"), value.StringValue(",")); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- json ---

func TestBindJsonArrayAppendTopLevel(t *testing.T) {
	got, err := BindJsonArrayAppend(
		value.JsonValue(`[1,2]`),
		value.StringValue("$"),
		value.IntValue(3),
	)
	if err != nil {
		t.Fatalf("BindJsonArrayAppend: %v", err)
	}
	s, _ := got.ToString()
	var arr []any
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("want 3 entries, got %v", arr)
	}
}

func TestBindJsonArrayAppendNestedField(t *testing.T) {
	got, err := BindJsonArrayAppend(
		value.JsonValue(`{"a":[1]}`),
		value.StringValue("$.a"),
		value.IntValue(2),
	)
	if err != nil {
		t.Fatalf("BindJsonArrayAppend: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "[1,2]") {
		t.Fatalf("expected appended array, got %s", s)
	}
}

func TestBindJsonArrayAppendNullDoc(t *testing.T) {
	got, err := BindJsonArrayAppend(nil, value.StringValue("$"), value.IntValue(1))
	if err != nil {
		t.Fatalf("BindJsonArrayAppend NULL: %v", err)
	}
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindJsonArrayAppendInvalidJson(t *testing.T) {
	_, err := BindJsonArrayAppend(value.StringValue("not json"), value.StringValue("$"), value.IntValue(1))
	if err == nil {
		t.Fatalf("invalid JSON should error")
	}
}

func TestBindJsonArrayAppendArity(t *testing.T) {
	if _, err := BindJsonArrayAppend(value.JsonValue(`[]`)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindJsonArrayInsertIndexed(t *testing.T) {
	got, err := BindJsonArrayInsert(
		value.JsonValue(`[1,2,3]`),
		value.StringValue("$[1]"),
		value.IntValue(99),
	)
	if err != nil {
		t.Fatalf("BindJsonArrayInsert: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "1,99,2,3") && !strings.Contains(s, "[1,99,2,3]") {
		t.Fatalf("unexpected: %s", s)
	}
}

func TestBindJsonArrayInsertNullDoc(t *testing.T) {
	got, err := BindJsonArrayInsert(nil, value.StringValue("$[0]"), value.IntValue(1))
	if err != nil {
		t.Fatalf("BindJsonArrayInsert NULL: %v", err)
	}
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

// TestBindJsonArrayInsertInsertEachFalse exercises the
// insert_each_element=FALSE trailing-bool branch: the inserted ARRAY
// must be preserved as a single nested element.
func TestBindJsonArrayInsertInsertEachFalse(t *testing.T) {
	inner := &value.ArrayValue{Values: []value.Value{value.IntValue(9), value.IntValue(8)}}
	got, err := BindJsonArrayInsert(
		value.JsonValue(`[1,2,3]`),
		value.StringValue("$[1]"),
		inner,
		value.BoolValue(false),
	)
	if err != nil {
		t.Fatalf("BindJsonArrayInsert: %v", err)
	}
	s, _ := got.ToString()
	// expect the inner array to remain a nested array literal
	if !strings.Contains(s, "[9,8]") {
		t.Fatalf("expected nested array preserved, got %s", s)
	}
}

// TestBindJsonArrayAppendStructValue covers the struct branch of
// jsonAnyOf via JSON_ARRAY_APPEND.
func TestBindJsonArrayAppendStructValue(t *testing.T) {
	st := &value.StructValue{
		Keys:   []string{"k"},
		Values: []value.Value{value.IntValue(1)},
		M:      map[string]value.Value{"k": value.IntValue(1)},
	}
	got, err := BindJsonArrayAppend(
		value.JsonValue(`[]`),
		value.StringValue("$"),
		st,
	)
	if err != nil {
		t.Fatalf("BindJsonArrayAppend struct: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, `"k":1`) {
		t.Fatalf("want struct serialized, got %s", s)
	}
}

// TestBindJsonArrayAppendBoolTrailing covers the bool-trailing branch
// of splitJsonModifyArgs (create_if_missing flag).
func TestBindJsonArrayAppendBoolTrailing(t *testing.T) {
	got, err := BindJsonArrayAppend(
		value.JsonValue(`[1]`),
		value.StringValue("$"),
		value.IntValue(2),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatalf("BindJsonArrayAppend bool trailing: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "1,2") {
		t.Fatalf("want '[1,2]', got %s", s)
	}
}

// TestBindJsonArrayAppendIntTrailing covers the IntValue trailing-arg
// branch.
func TestBindJsonArrayAppendIntTrailing(t *testing.T) {
	got, err := BindJsonArrayAppend(
		value.JsonValue(`[1]`),
		value.StringValue("$"),
		value.IntValue(2),
		value.IntValue(1),
	)
	if err != nil {
		t.Fatalf("BindJsonArrayAppend int trailing: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "1,2") {
		t.Fatalf("want '[1,2]', got %s", s)
	}
}

func TestBindJsonArrayInsertInvalid(t *testing.T) {
	if _, err := BindJsonArrayInsert(value.StringValue("garbage"), value.StringValue("$[0]"), value.IntValue(1)); err == nil {
		t.Fatalf("invalid JSON should error")
	}
	if _, err := BindJsonArrayInsert(value.JsonValue(`[]`)); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- json helper ---

func TestBindSafeToJson(t *testing.T) {
	got, err := BindSafeToJson(value.IntValue(42))
	if err != nil {
		t.Fatalf("BindSafeToJson int: %v", err)
	}
	s, _ := got.ToString()
	if strings.TrimSpace(s) != "42" {
		t.Fatalf("want '42', got %q", s)
	}

	arr := &value.ArrayValue{Values: []value.Value{value.IntValue(1), value.IntValue(2)}}
	got, _ = BindSafeToJson(arr)
	s, _ = got.ToString()
	if s != "[1,2]" {
		t.Fatalf("want '[1,2]', got %q", s)
	}

	if got, _ := BindSafeToJson(nil); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindSafeToJson(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- hash ---

func TestBindHighwayFingerprint128(t *testing.T) {
	got, err := BindHighwayFingerprint128(value.StringValue("hello"))
	if err != nil {
		t.Fatalf("BindHighwayFingerprint128: %v", err)
	}
	b, _ := got.ToBytes()
	if len(b) != 16 {
		t.Fatalf("want 16-byte fingerprint, got %d", len(b))
	}

	// Bytes input goes through the BytesValue branch.
	got, err = BindHighwayFingerprint128(value.BytesValue([]byte{1, 2, 3}))
	if err != nil {
		t.Fatalf("BindHighwayFingerprint128 bytes: %v", err)
	}
	b2, _ := got.ToBytes()
	if len(b2) != 16 {
		t.Fatalf("want 16-byte fingerprint, got %d", len(b2))
	}

	if got, _ := BindHighwayFingerprint128(nil); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindHighwayFingerprint128(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- hll ---

func TestBindHLLCountMergeDistinct(t *testing.T) {
	agg := BindHLLCountMerge()()
	if err := agg.Step("a"); err != nil {
		t.Fatalf("step: %v", err)
	}
	if err := agg.Step("a"); err != nil {
		t.Fatalf("step dup: %v", err)
	}
	if err := agg.Step("b"); err != nil {
		t.Fatalf("step: %v", err)
	}
	if err := agg.Step(nil); err != nil {
		t.Fatalf("step nil: %v", err)
	}
	got, err := agg.Done()
	if err != nil {
		t.Fatalf("done: %v", err)
	}
	// EncodeValue returns int64 for IntValue.
	n, ok := got.(int64)
	if !ok {
		t.Fatalf("Done = %T, want int64", got)
	}
	if n != 2 {
		t.Fatalf("want 2 distinct, got %d", n)
	}
}

func TestBindHLLCountMergePartial(t *testing.T) {
	agg := BindHLLCountMergePartial()()
	if err := agg.Step("x"); err != nil {
		t.Fatalf("step: %v", err)
	}
	if err := agg.Step("y"); err != nil {
		t.Fatalf("step: %v", err)
	}
	got, err := agg.Done()
	if err != nil {
		t.Fatalf("done: %v", err)
	}
	if got == nil {
		t.Fatalf("Done returned nil")
	}
	// Decode the bytes payload via value.DecodeValue to recover a
	// BytesValue, then JSON-unmarshal to the key list.
	dv, err := value.DecodeValue(got)
	if err != nil {
		t.Fatalf("DecodeValue: %v", err)
	}
	b, err := dv.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes: %v", err)
	}
	var keys []string
	if err := json.Unmarshal(b, &keys); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %v", keys)
	}
}

// --- array ---

func TestBindFlatten(t *testing.T) {
	in := &value.ArrayValue{Values: []value.Value{
		&value.ArrayValue{Values: []value.Value{value.IntValue(1), value.IntValue(2)}},
		&value.ArrayValue{Values: []value.Value{value.IntValue(3)}},
	}}
	got, err := BindFlatten(in)
	if err != nil {
		t.Fatalf("BindFlatten: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 3 {
		t.Fatalf("want 3 elems, got %d", len(arr.Values))
	}

	// NULL element passes through.
	in2 := &value.ArrayValue{Values: []value.Value{
		nil,
		&value.ArrayValue{Values: []value.Value{value.IntValue(9)}},
	}}
	got, _ = BindFlatten(in2)
	arr, _ = got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 elems (nil + 1), got %d", len(arr.Values))
	}

	// NULL input.
	if got, _ := BindFlatten(nil); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}

	// Non-array argument.
	if _, err := BindFlatten(value.IntValue(1)); err == nil {
		t.Fatalf("non-array input should error")
	}

	// Arity.
	if _, err := BindFlatten(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindArrayZip(t *testing.T) {
	a := &value.ArrayValue{Values: []value.Value{value.IntValue(1), value.IntValue(2)}}
	b := &value.ArrayValue{Values: []value.Value{value.StringValue("x"), value.StringValue("y"), value.StringValue("z")}}
	got, err := BindArrayZip(a, b)
	if err != nil {
		t.Fatalf("BindArrayZip: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 zipped pairs (min length), got %d", len(arr.Values))
	}
	st, _ := arr.Values[0].ToStruct()
	if len(st.Keys) != 2 || st.Keys[0] != "f0" || st.Keys[1] != "f1" {
		t.Fatalf("want f0/f1, got %v", st.Keys)
	}

	// NULL argument returns NULL.
	if got, _ := BindArrayZip(nil, b); got != nil {
		t.Fatalf("NULL arg must produce NULL output")
	}

	// Non-array argument.
	if _, err := BindArrayZip(value.IntValue(1), value.IntValue(2)); err == nil {
		t.Fatalf("non-array input should error")
	}

	// Arity.
	if _, err := BindArrayZip(a); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- aggregate ---

func TestBindGrouping(t *testing.T) {
	got, err := BindGrouping(value.IntValue(1))
	if err != nil {
		t.Fatalf("BindGrouping: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 0 {
		t.Fatalf("want 0 (column is grouped), got %d", n)
	}
	if _, err := BindGrouping(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- net ---

func TestBindNetIpInNet(t *testing.T) {
	got, err := BindNetIpInNet(value.StringValue("10.0.0.1"), value.StringValue("10.0.0.0/8"))
	if err != nil {
		t.Fatalf("BindNetIpInNet: %v", err)
	}
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("10.0.0.1 must be in 10.0.0.0/8")
	}

	got, _ = BindNetIpInNet(value.StringValue("11.0.0.1"), value.StringValue("10.0.0.0/8"))
	b, _ = got.ToBool()
	if b {
		t.Fatalf("11.0.0.1 must not be in 10.0.0.0/8")
	}

	if _, err := BindNetIpInNet(value.StringValue("bad"), value.StringValue("10.0.0.0/8")); err == nil {
		t.Fatalf("invalid IP should error")
	}
	if _, err := BindNetIpInNet(value.StringValue("10.0.0.1"), value.StringValue("bad")); err == nil {
		t.Fatalf("invalid CIDR should error")
	}
	if got, _ := BindNetIpInNet(nil, value.StringValue("10.0.0.0/8")); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindNetIpInNet(value.StringValue("10.0.0.1")); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindNetParseIpAndFormat(t *testing.T) {
	got, err := BindNetParseIp(value.StringValue("10.0.0.1"))
	if err != nil {
		t.Fatalf("BindNetParseIp: %v", err)
	}
	n, _ := got.ToInt64()
	want := int64(10)<<24 | int64(1)
	if n != want {
		t.Fatalf("parse 10.0.0.1 = %d, want %d", n, want)
	}
	// Round-trip
	got, err = BindNetFormatIp(value.IntValue(n))
	if err != nil {
		t.Fatalf("BindNetFormatIp: %v", err)
	}
	s, _ := got.ToString()
	if s != "10.0.0.1" {
		t.Fatalf("format = %q, want '10.0.0.1'", s)
	}

	// IPv6
	if _, err := BindNetParseIp(value.StringValue("::1")); err != nil {
		t.Fatalf("BindNetParseIp ipv6: %v", err)
	}

	if _, err := BindNetParseIp(value.StringValue("not an ip")); err == nil {
		t.Fatalf("invalid IP should error")
	}
	if got, _ := BindNetParseIp(nil); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindNetParseIp(); err == nil {
		t.Fatalf("arity error expected")
	}

	if _, err := BindNetFormatIp(value.IntValue(-1)); err == nil {
		t.Fatalf("out-of-range IPv4 packed must error")
	}
	if got, _ := BindNetFormatIp(nil); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindNetFormatIp(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindNetParsePackedIp(t *testing.T) {
	got, err := BindNetParsePackedIp(value.StringValue("192.168.1.1"))
	if err != nil {
		t.Fatalf("BindNetParsePackedIp: %v", err)
	}
	b, _ := got.ToBytes()
	if len(b) != 4 || b[0] != 192 || b[1] != 168 || b[2] != 1 || b[3] != 1 {
		t.Fatalf("packed v4 = %v, want [192 168 1 1]", b)
	}

	got, _ = BindNetParsePackedIp(value.StringValue("::1"))
	b, _ = got.ToBytes()
	if len(b) != 16 {
		t.Fatalf("packed v6 must be 16 bytes, got %d", len(b))
	}

	// Round-trip format.
	got, err = BindNetFormatPackedIp(value.BytesValue([]byte{8, 8, 8, 8}))
	if err != nil {
		t.Fatalf("BindNetFormatPackedIp: %v", err)
	}
	s, _ := got.ToString()
	if s != "8.8.8.8" {
		t.Fatalf("format = %q, want '8.8.8.8'", s)
	}

	if _, err := BindNetParsePackedIp(value.StringValue("not")); err == nil {
		t.Fatalf("invalid IP should error")
	}
	if got, _ := BindNetParsePackedIp(nil); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindNetParsePackedIp(); err == nil {
		t.Fatalf("arity error expected")
	}

	if _, err := BindNetFormatPackedIp(value.BytesValue([]byte{1, 2, 3})); err == nil {
		t.Fatalf("wrong length should error")
	}
	if got, _ := BindNetFormatPackedIp(nil); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindNetFormatPackedIp(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindNetMakeNet(t *testing.T) {
	got, err := BindNetMakeNet(value.StringValue("10.1.2.3"), value.IntValue(16))
	if err != nil {
		t.Fatalf("BindNetMakeNet: %v", err)
	}
	s, _ := got.ToString()
	if s != "10.1.0.0/16" {
		t.Fatalf("want '10.1.0.0/16', got %q", s)
	}

	// IPv6
	got, err = BindNetMakeNet(value.StringValue("2001:db8::1"), value.IntValue(32))
	if err != nil {
		t.Fatalf("BindNetMakeNet v6: %v", err)
	}
	s, _ = got.ToString()
	if !strings.Contains(s, "/32") {
		t.Fatalf("want /32 prefix, got %q", s)
	}

	if _, err := BindNetMakeNet(value.StringValue("10.1.2.3"), value.IntValue(99)); err == nil {
		t.Fatalf("out-of-range prefix should error")
	}
	if _, err := BindNetMakeNet(value.StringValue("bad"), value.IntValue(8)); err == nil {
		t.Fatalf("invalid IP should error")
	}
	if got, _ := BindNetMakeNet(nil, value.IntValue(8)); got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindNetMakeNet(value.StringValue("10.0.0.1")); err == nil {
		t.Fatalf("arity error expected")
	}
}
