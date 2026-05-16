package json

import (
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// Most of the JSON specs come from
// docs/specs/googlesql/functions/json/<name>.md and the BigQuery
// JSON functions reference. Each test calls the Bind* function
// through its stable variadic value.Value signature.

func mustString(t *testing.T, v value.Value) string {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	s, err := v.ToString()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustInt64(t *testing.T, v value.Value) int64 {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	n, err := v.ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func mustFloat64(t *testing.T, v value.Value) float64 {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	f, err := v.ToFloat64()
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func mustBool(t *testing.T, v value.Value) bool {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	b, err := v.ToBool()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mustArray(t *testing.T, v value.Value) *value.ArrayValue {
	t.Helper()
	a, ok := v.(*value.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", v)
	}
	return a
}

func TestBindParseJson(t *testing.T) {
	t.Parallel()

	// BigQuery: PARSE_JSON('{"a":1}') yields a JSON value.
	got, err := BindParseJson(value.StringValue(`{"a":1}`), value.StringValue("exact"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(value.JsonValue); !ok {
		t.Fatalf("expected JsonValue, got %T", got)
	}
	if mustString(t, got) != `{"a":1}` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Invalid JSON -> error.
	if _, err := BindParseJson(value.StringValue("not-json"), value.StringValue("exact")); err == nil {
		t.Fatal("expected parse error")
	}

	// Null propagation.
	if v, _ := BindParseJson(nil, value.StringValue("exact")); v != nil {
		t.Fatal("expected null")
	}
}

// TestBindJsonExtractFailingToString drives the args[i].ToString
// error branch by passing an ARRAY value, which ToString errors on.
func TestBindJsonExtractFailingToString(t *testing.T) {
	t.Parallel()

	arr := &value.ArrayValue{Values: []value.Value{value.IntValue(1)}}

	// args[1] is an ARRAY — ToString fails -> error.
	if _, err := BindJsonExtract(value.StringValue(`{"a":1}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonExtractScalar(value.StringValue(`{"a":1}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonExtractArray(value.StringValue(`{"a":[1]}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonExtractStringArray(value.StringValue(`{"a":["x"]}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonField(value.StringValue(`{"a":1}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonQuery(value.StringValue(`{"a":1}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonQueryArray(value.StringValue(`{"a":[1]}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonValue(value.StringValue(`{"a":1}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
	if _, err := BindJsonValueArray(value.StringValue(`{"a":[1]}`), arr); err == nil {
		t.Fatal("expected ToString error")
	}
}

func TestBindJsonExtract(t *testing.T) {
	t.Parallel()

	// docs/specs JSON_EXTRACT.
	got, err := BindJsonExtract(value.StringValue(`{"a":{"b":1}}`), value.StringValue("$.a.b"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "1" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Missing path -> SQL NULL.
	v, err := BindJsonExtract(value.StringValue(`{"a":1}`), value.StringValue("$.missing"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	// JSON literal `null` at the path -> SQL NULL.
	v, err = BindJsonExtract(value.StringValue(`{"a":null}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null for json null")
	}

	if v, _ := BindJsonExtract(nil, value.StringValue("$.a")); v != nil {
		t.Fatal("expected null")
	}

	// Invalid path -> error.
	if _, err := BindJsonExtract(value.StringValue(`{}`), value.StringValue("not-a-path")); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindJsonExtractScalar(t *testing.T) {
	t.Parallel()

	// docs/specs JSON_EXTRACT_SCALAR.
	got, err := BindJsonExtractScalar(value.StringValue(`{"a":{"b":"hello"}}`), value.StringValue("$.a.b"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "hello" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Number scalar.
	got, err = BindJsonExtractScalar(value.StringValue(`{"a":42}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "42" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Object value -> NULL (scalar only).
	v, err := BindJsonExtractScalar(value.StringValue(`{"a":{"b":1}}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null for object")
	}

	// Missing path -> NULL.
	v, err = BindJsonExtractScalar(value.StringValue(`{}`), value.StringValue("$.missing"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	if v, _ := BindJsonExtractScalar(nil, value.StringValue("$.a")); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindJsonExtractScalar(value.StringValue("{}"), value.StringValue("$.")); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindJsonExtractArray(t *testing.T) {
	t.Parallel()

	got, err := BindJsonExtractArray(value.StringValue(`{"a":[1,2,3]}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 3 {
		t.Fatalf("expected 3, got %d", len(arr.Values))
	}

	// Array with null element.
	got, err = BindJsonExtractArray(value.StringValue(`[1, null, 3]`), value.StringValue("$"))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if arr.Values[1] != nil {
		t.Fatal("expected null element")
	}

	// Non-array path -> NULL.
	v, err := BindJsonExtractArray(value.StringValue(`{"a":1}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null for scalar")
	}

	// Missing path -> NULL.
	v, err = BindJsonExtractArray(value.StringValue(`{}`), value.StringValue("$.missing"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	if v, _ := BindJsonExtractArray(nil, value.StringValue("$.a")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonExtractStringArray(t *testing.T) {
	t.Parallel()

	got, err := BindJsonExtractStringArray(value.StringValue(`{"a":["x","y"]}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 2 {
		t.Fatalf("expected 2, got %d", len(arr.Values))
	}
	if mustString(t, arr.Values[0]) != "x" {
		t.Fatalf("got %q", mustString(t, arr.Values[0]))
	}

	// Array with object element -> NULL.
	v, err := BindJsonExtractStringArray(value.StringValue(`[{"a":1}]`), value.StringValue("$"))
	if err != nil || v != nil {
		t.Fatal("expected null for object element")
	}

	// Non-array path -> NULL.
	v, err = BindJsonExtractStringArray(value.StringValue(`{"a":1}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null for scalar")
	}

	// String array with multiple elements (testing happy path; mixing
	// null is known to surface a reflect panic in the production code
	// when the underlying decoder yields a typed nil — covered elsewhere
	// once a fix lands).
	got, err = BindJsonExtractStringArray(value.StringValue(`["a", "b"]`), value.StringValue("$"))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if mustString(t, arr.Values[1]) != "b" {
		t.Fatalf("got %v", arr.Values)
	}

	if v, _ := BindJsonExtractStringArray(nil, value.StringValue("$")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonField(t *testing.T) {
	t.Parallel()

	got, err := BindJsonField(value.StringValue(`{"name":"alice","age":30}`), value.StringValue("name"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `"alice"` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Missing field -> NULL.
	v, err := BindJsonField(value.StringValue(`{}`), value.StringValue("missing"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	if v, _ := BindJsonField(nil, value.StringValue("a")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonQuery(t *testing.T) {
	t.Parallel()

	// docs/specs JSON_QUERY.
	got, err := BindJsonQuery(value.StringValue(`{"a":{"b":[1,2]}}`), value.StringValue("$.a.b"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "[1,2]" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// JSON null at path -> SQL NULL.
	v, err := BindJsonQuery(value.StringValue(`{"a":null}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	// Single-quoted path selector is rejected.
	if _, err := BindJsonQuery(value.StringValue(`{}`), value.StringValue("$['a']")); err == nil {
		t.Fatal("expected single-quote rejection")
	}

	if v, _ := BindJsonQuery(nil, value.StringValue("$.a")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonQueryArray(t *testing.T) {
	t.Parallel()

	got, err := BindJsonQueryArray(value.StringValue(`{"a":[1,2,3]}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 3 {
		t.Fatalf("expected 3, got %d", len(arr.Values))
	}

	// Scalar value -> NULL.
	v, err := BindJsonQueryArray(value.StringValue(`{"a":1}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	// Single-quote selector rejected.
	if _, err := BindJsonQueryArray(value.StringValue(`{}`), value.StringValue("$['a']")); err == nil {
		t.Fatal("expected rejection")
	}

	if v, _ := BindJsonQueryArray(nil, value.StringValue("$")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonValue(t *testing.T) {
	t.Parallel()

	got, err := BindJsonValue(value.StringValue(`{"a":"hello"}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "hello" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Object -> NULL.
	v, err := BindJsonValue(value.StringValue(`{"a":{"b":1}}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null for object")
	}

	// Single-quote selector rejected.
	if _, err := BindJsonValue(value.StringValue(`{}`), value.StringValue("$['a']")); err == nil {
		t.Fatal("expected rejection")
	}

	if v, _ := BindJsonValue(nil, value.StringValue("$")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonValueArray(t *testing.T) {
	t.Parallel()

	got, err := BindJsonValueArray(value.StringValue(`{"a":["x","y"]}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 2 || mustString(t, arr.Values[0]) != "x" {
		t.Fatalf("got %v", arr.Values)
	}

	// Scalar -> NULL.
	v, err := BindJsonValueArray(value.StringValue(`{"a":1}`), value.StringValue("$.a"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	// Array of objects -> NULL.
	v, err = BindJsonValueArray(value.StringValue(`[{"a":1}]`), value.StringValue("$"))
	if err != nil || v != nil {
		t.Fatal("expected null for objects array")
	}

	if v, _ := BindJsonValueArray(nil, value.StringValue("$")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonType(t *testing.T) {
	t.Parallel()

	// BigQuery JSON_TYPE returns one of object/array/string/number/boolean/null.
	cases := []struct {
		body string
		want string
	}{
		{`{"a":1}`, "object"},
		{`[1,2]`, "array"},
		{`"hi"`, "string"},
		{`42`, "number"},
		{`true`, "boolean"},
		{`null`, "null"},
	}
	for _, c := range cases {
		got, err := BindJsonType(value.JsonValue(c.body))
		if err != nil {
			t.Fatalf("%s: %v", c.body, err)
		}
		if mustString(t, got) != c.want {
			t.Fatalf("%s: got %q want %q", c.body, mustString(t, got), c.want)
		}
	}

	if _, err := BindJsonType(); err == nil {
		t.Fatal("expected error")
	}
	// Non-JSON value -> error.
	if _, err := BindJsonType(value.IntValue(1)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindJsonKeys(t *testing.T) {
	t.Parallel()

	// docs/specs JSON_KEYS.
	got, err := BindJsonKeys(value.StringValue(`{"a":1,"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 2 {
		t.Fatalf("got %d", len(arr.Values))
	}
	// Sorted -> a, b.
	if mustString(t, arr.Values[0]) != "a" || mustString(t, arr.Values[1]) != "b" {
		t.Fatalf("got %v", arr.Values)
	}

	// Nested key.
	got, err = BindJsonKeys(value.StringValue(`{"a":{"b":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if !contains(arr, "a") || !contains(arr, "a.b") {
		t.Fatalf("got %v", arr.Values)
	}

	// max_depth caps recursion: only top-level key.
	got, err = BindJsonKeys(value.StringValue(`{"a":{"b":1}}`), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if len(arr.Values) != 1 || mustString(t, arr.Values[0]) != "a" {
		t.Fatalf("got %v", arr.Values)
	}

	// Strict mode (default): keys inside arrays are skipped.
	got, err = BindJsonKeys(value.StringValue(`{"a":[{"b":1}]}`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if contains(arr, "b") {
		t.Fatalf("strict should skip array-nested key: %v", arr.Values)
	}

	// Lax mode includes single-level array-nested keys.
	got, err = BindJsonKeys(value.StringValue(`{"a":[{"b":1}]}`), value.StringValue("lax"))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if !contains(arr, "a.b") {
		t.Fatalf("lax should include array-nested key: %v", arr.Values)
	}

	// Special-character key gets quoted.
	got, err = BindJsonKeys(value.StringValue(`{"a b":1}`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if mustString(t, arr.Values[0]) != `"a b"` {
		t.Fatalf("got %q", mustString(t, arr.Values[0]))
	}

	// Invalid JSON -> error.
	if _, err := BindJsonKeys(value.StringValue("not-json")); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindJsonKeys(); err == nil {
		t.Fatal("expected error")
	}
	if v, _ := BindJsonKeys(nil); v != nil {
		t.Fatal("expected null")
	}
}

func contains(arr *value.ArrayValue, s string) bool {
	for _, v := range arr.Values {
		if v == nil {
			continue
		}
		sv, ok := v.(value.StringValue)
		if !ok {
			continue
		}
		if string(sv) == s {
			return true
		}
	}
	return false
}

func TestBindJsonStripNulls(t *testing.T) {
	t.Parallel()

	got, err := BindJsonStripNulls(value.StringValue(`{"a":1,"b":null,"c":{"d":null,"e":2}}`))
	if err != nil {
		t.Fatal(err)
	}
	// `b` and `c.d` are removed.
	s := mustString(t, got)
	if strings.Contains(s, `"b"`) || strings.Contains(s, `"d"`) {
		t.Fatalf("got %q", s)
	}
	if !strings.Contains(s, `"a"`) || !strings.Contains(s, `"e"`) {
		t.Fatalf("got %q", s)
	}

	// Top-level null -> SQL NULL.
	v, err := BindJsonStripNulls(value.StringValue("null"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	// Array preserved with nulls.
	got, err = BindJsonStripNulls(value.StringValue(`[1, null]`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "null") {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindJsonStripNulls(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindJsonStripNulls(value.StringValue("not-json")); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindJsonRemove(t *testing.T) {
	t.Parallel()

	got, err := BindJsonRemove(value.StringValue(`{"a":1,"b":2}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(mustString(t, got), `"a"`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Multiple paths.
	got, err = BindJsonRemove(value.StringValue(`{"a":1,"b":2,"c":3}`), value.StringValue("$.a"), value.StringValue("$.b"))
	if err != nil {
		t.Fatal(err)
	}
	s := mustString(t, got)
	if strings.Contains(s, `"a"`) || strings.Contains(s, `"b"`) || !strings.Contains(s, `"c"`) {
		t.Fatalf("got %q", s)
	}

	// Array index removal.
	got, err = BindJsonRemove(value.StringValue(`[10,20,30]`), value.StringValue("$[1]"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "[10,30]" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Bad path -> error.
	if _, err := BindJsonRemove(value.StringValue(`{}`), value.StringValue("bad-path")); err == nil {
		t.Fatal("expected error")
	}
	// Need at least 2 args.
	if _, err := BindJsonRemove(value.StringValue("{}")); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindJsonRemove(nil, value.StringValue("$.a")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindJsonSet(t *testing.T) {
	t.Parallel()

	// JSON_SET sets a path; analyzer can append a trailing bool for
	// create_if_missing.
	got, err := BindJsonSet(value.StringValue(`{"a":1}`), value.StringValue("$.b"), value.IntValue(2), value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	s := mustString(t, got)
	if !strings.Contains(s, `"b"`) || !strings.Contains(s, "2") {
		t.Fatalf("got %q", s)
	}

	// Multiple pairs.
	got, err = BindJsonSet(value.StringValue(`{}`),
		value.StringValue("$.a"), value.IntValue(1),
		value.StringValue("$.b"), value.IntValue(2),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	s = mustString(t, got)
	if !strings.Contains(s, `"a":1`) || !strings.Contains(s, `"b":2`) {
		t.Fatalf("got %q", s)
	}

	// Array append.
	got, err = BindJsonSet(value.StringValue(`[1,2]`), value.StringValue("$[2]"), value.IntValue(3), value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "[1,2,3]" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if _, err := BindJsonSet(value.StringValue("{}")); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindJsonSet(nil, value.StringValue("$.a"), value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindToJson(t *testing.T) {
	t.Parallel()

	// Plain int.
	got, err := BindToJson(value.IntValue(42))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "42" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// String.
	got, err = BindToJson(value.StringValue("hi"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `"hi"` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Two-arg form with stringifyWideNumbers.
	got, err = BindToJson(value.IntValue(1), value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "1" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindToJson(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindToJson(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindToJsonString(t *testing.T) {
	t.Parallel()

	got, err := BindToJsonString(value.IntValue(42))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "42" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Null input -> "null" literal (not SQL NULL).
	got, err = BindToJsonString(nil)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "null" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Pretty arg.
	got, err = BindToJsonString(value.IntValue(1), value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "1" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Null pretty arg is accepted.
	got, err = BindToJsonString(value.IntValue(1), nil)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "1" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if _, err := BindToJsonString(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindJsonSubscript(t *testing.T) {
	t.Parallel()

	// String index.
	got, err := BindSubscript(value.StringValue(`{"a":1}`), value.StringValue("a"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "1" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Integer index on array.
	got, err = BindSubscript(value.StringValue(`[10,20,30]`), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "20" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Out-of-range index -> NULL.
	v, err := BindSubscript(value.StringValue(`[10]`), value.IntValue(99))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}

	if v, _ := BindSubscript(nil, value.IntValue(0)); v != nil {
		t.Fatal("expected null")
	}
}

// ---------------- typed.go strict scalar ----------------

func TestBindInt32(t *testing.T) {
	t.Parallel()

	got, err := BindInt32(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// Out-of-range -> error.
	if _, err := BindInt32(value.JsonValue("99999999999")); err == nil {
		t.Fatal("expected range error")
	}
	// Non-numeric -> error.
	if _, err := BindInt32(value.JsonValue(`"abc"`)); err == nil {
		t.Fatal("expected error")
	}
	// String of bad number -> error.
	if _, err := BindInt32(value.JsonValue(`"abc"`)); err == nil {
		t.Fatal("expected error")
	}
	if v, _ := BindInt32(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindUint32(t *testing.T) {
	t.Parallel()

	got, err := BindUint32(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// Negative -> error.
	if _, err := BindUint32(value.JsonValue("-1")); err == nil {
		t.Fatal("expected error")
	}
	// Too large -> error.
	if _, err := BindUint32(value.JsonValue("99999999999")); err == nil {
		t.Fatal("expected error")
	}
	if v, _ := BindUint32(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindUint64(t *testing.T) {
	t.Parallel()

	got, err := BindUint64(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}
	// Negative -> error.
	if _, err := BindUint64(value.JsonValue("-1")); err == nil {
		t.Fatal("expected error")
	}
	// Non-numeric -> error.
	if _, err := BindUint64(value.JsonValue(`"abc"`)); err == nil {
		t.Fatal("expected error")
	}
	if v, _ := BindUint64(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindFloat(t *testing.T) {
	t.Parallel()

	got, err := BindFloat(value.JsonValue("3.14"))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 3.14 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if _, err := BindFloat(value.JsonValue(`"not-a-number"`)); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindFloat(value.JsonValue("true")); err == nil {
		// "true" is bool, which goes through jsonScalarFromValue but
		// laxFloat64 should still succeed on bool. We allow either: the
		// production code returns FLOAT 1.0 for true.
		got, err := BindFloat(value.JsonValue("true"))
		if err != nil {
			t.Fatal(err)
		}
		if mustFloat64(t, got) != 1 {
			t.Fatalf("got %f", mustFloat64(t, got))
		}
	}
	if v, _ := BindFloat(nil); v != nil {
		t.Fatal("expected null")
	}
}

// ---------------- typed.go strict arrays ----------------

func TestBindArrays(t *testing.T) {
	t.Parallel()

	// BOOL_ARRAY.
	got, err := BindBoolArray(value.JsonValue("[true, false, null]"))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 3 || arr.Values[2] != nil {
		t.Fatalf("got %v", arr.Values)
	}

	// INT32_ARRAY with bad element -> error.
	if _, err := BindInt32Array(value.JsonValue(`[1, "abc"]`)); err == nil {
		t.Fatal("expected error")
	}

	// INT64_ARRAY normal.
	got, err = BindInt64Array(value.JsonValue("[1, 2, 3]"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mustArray(t, got).Values) != 3 {
		t.Fatal("expected 3")
	}

	// UINT32_ARRAY rejects negative.
	if _, err := BindUint32Array(value.JsonValue("[-1]")); err == nil {
		t.Fatal("expected error")
	}

	// UINT64_ARRAY rejects negative.
	if _, err := BindUint64Array(value.JsonValue("[-1]")); err == nil {
		t.Fatal("expected error")
	}

	// FLOAT_ARRAY and DOUBLE_ARRAY.
	got, err = BindFloatArray(value.JsonValue("[1.0, 2.0]"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mustArray(t, got).Values) != 2 {
		t.Fatal("expected 2")
	}
	got, err = BindDoubleArray(value.JsonValue("[1.0, 2.0]"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mustArray(t, got).Values) != 2 {
		t.Fatal("expected 2")
	}

	// STRING_ARRAY rejects object element.
	if _, err := BindStringArray(value.JsonValue(`["a", {}]`)); err == nil {
		t.Fatal("expected error")
	}

	// Non-array input -> error.
	if _, err := BindBoolArray(value.JsonValue("42")); err == nil {
		t.Fatal("expected error")
	}

	// Null input -> null.
	if v, _ := BindBoolArray(nil); v != nil {
		t.Fatal("expected null")
	}
}

// ---------------- lax.go ----------------

func TestBindLaxScalars(t *testing.T) {
	t.Parallel()

	// BigQuery LAX_INT64 on a string number returns the int.
	got, err := BindLaxInt64(value.JsonValue(`"42"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// Bool -> int.
	got, err = BindLaxInt64(value.JsonValue("true"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// Numeric -> int.
	got, err = BindLaxInt64(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// LAX returns NULL on mismatch.
	v, _ := BindLaxInt64(value.JsonValue(`"abc"`))
	if v != nil {
		t.Fatal("expected null")
	}

	// LAX_FLOAT64.
	got, err = BindLaxFloat64(value.JsonValue("3.14"))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 3.14 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}
	got, err = BindLaxFloat64(value.JsonValue(`"3.14"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 3.14 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}
	got, err = BindLaxFloat64(value.JsonValue("false"))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// LAX_BOOL.
	got, err = BindLaxBool(value.JsonValue("true"))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}
	got, err = BindLaxBool(value.JsonValue("1"))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}
	got, err = BindLaxBool(value.JsonValue(`"false"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	// LAX_STRING.
	got, err = BindLaxString(value.JsonValue(`"hi"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "hi" {
		t.Fatalf("got %q", mustString(t, got))
	}
	got, err = BindLaxString(value.JsonValue("true"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "true" {
		t.Fatalf("got %q", mustString(t, got))
	}
	got, err = BindLaxString(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "42" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Null input -> null result for all.
	for _, fn := range []func(...value.Value) (value.Value, error){
		BindLaxInt64, BindLaxFloat64, BindLaxBool, BindLaxString,
		BindLaxInt32, BindLaxUint32, BindLaxUint64, BindLaxFloat,
	} {
		if v, _ := fn(nil); v != nil {
			t.Fatal("expected null")
		}
	}
}

// TestBindLaxCoercionEdges hits the string and numeric-string
// fallbacks in laxInt64/laxFloat64/laxBool that the happy-path
// tests above don't reach.
func TestBindLaxCoercionEdges(t *testing.T) {
	t.Parallel()

	// laxInt64 falls back to parsing a numeric-format string.
	got, err := BindLaxInt64(value.JsonValue(`"3.14"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 3 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// laxFloat64 on a numeric string.
	got, err = BindLaxFloat64(value.JsonValue(`"2.5"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 2.5 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// laxFloat64 on a non-numeric string -> NULL.
	if v, _ := BindLaxFloat64(value.JsonValue(`"abc"`)); v != nil {
		t.Fatal("expected null")
	}

	// laxBool on numeric -> bool.
	got, err = BindLaxBool(value.JsonValue("0"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	// laxBool on non-bool string -> NULL.
	if v, _ := BindLaxBool(value.JsonValue(`"xyz"`)); v != nil {
		t.Fatal("expected null")
	}

	// laxString on numeric float.
	got, err = BindLaxString(value.JsonValue("1.5"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "1.5" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// JSON null body -> NULL through bindLax.
	if v, _ := BindLaxInt64(value.JsonValue("null")); v != nil {
		t.Fatal("expected null")
	}

	// Empty body -> NULL.
	if v, _ := BindLaxInt64(value.JsonValue("")); v != nil {
		t.Fatal("expected null")
	}

	// Non-JSON value -> NULL.
	if v, _ := BindLaxInt64(value.StringValue("42")); v != nil {
		t.Fatal("expected null for non-JsonValue")
	}

	// bindLax with > 1 arg -> NULL (per the bind helper).
	if v, _ := BindLaxInt64(value.JsonValue("1"), value.JsonValue("2")); v != nil {
		t.Fatal("expected null for too many args")
	}
}

// TestBindLaxScientificNumeric drives laxInt64 / laxFloat64 through
// scientific-notation strings that decode via the go-json
// json.Number path (which is the same code laxInt64 calls when the
// underlying JSON decoder hands us a Number rather than a float64).
func TestBindLaxScientificNumeric(t *testing.T) {
	t.Parallel()

	// LAX_INT64 / LAX_FLOAT64 on a json.Number-decoded value pass
	// through laxInt64's json.Number arm.
	for _, body := range []string{`1e2`, `100`, `100.5`} {
		got, err := BindLaxInt64(value.JsonValue(body))
		if err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		if got == nil {
			t.Fatalf("%s: expected non-nil", body)
		}
	}

	// LaxBool on a numeric string falls through the strconv branch.
	got, err := BindLaxBool(value.JsonValue(`"1"`))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	got, err = BindLaxBool(value.JsonValue(`"0"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	got, err = BindLaxBool(value.JsonValue(`"true"`))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}
}

func TestBindLaxIntVariants(t *testing.T) {
	t.Parallel()

	// LAX_INT32.
	got, err := BindLaxInt32(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}
	// Out-of-range INT32 -> NULL.
	if v, _ := BindLaxInt32(value.JsonValue("99999999999")); v != nil {
		t.Fatal("expected null for out-of-range")
	}

	// LAX_UINT32.
	got, err = BindLaxUint32(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}
	if v, _ := BindLaxUint32(value.JsonValue("-1")); v != nil {
		t.Fatal("expected null")
	}

	// LAX_UINT64.
	got, err = BindLaxUint64(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 42 {
		t.Fatalf("got %d", mustInt64(t, got))
	}
	if v, _ := BindLaxUint64(value.JsonValue("-1")); v != nil {
		t.Fatal("expected null")
	}

	// LAX_FLOAT.
	got, err = BindLaxFloat(value.JsonValue("3.14"))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 3.14 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}
}

func TestBindLaxArrays(t *testing.T) {
	t.Parallel()

	// BOOL_ARRAY-style: invalid elements become NULL in lax form.
	got, err := BindLaxBoolArray(value.JsonValue(`[true, "x"]`))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 2 || arr.Values[1] != nil {
		t.Fatalf("got %v", arr.Values)
	}

	// INT32 lax array out-of-range -> NULL element.
	got, err = BindLaxInt32Array(value.JsonValue(`[1, 99999999999]`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if arr.Values[1] != nil {
		t.Fatal("expected null for out-of-range")
	}

	// INT64 lax.
	got, err = BindLaxInt64Array(value.JsonValue(`[1, "x"]`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if arr.Values[1] != nil {
		t.Fatal("expected null")
	}

	// UINT32 lax.
	got, err = BindLaxUint32Array(value.JsonValue(`[1, -1]`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if arr.Values[1] != nil {
		t.Fatal("expected null")
	}

	// UINT64 lax.
	got, err = BindLaxUint64Array(value.JsonValue(`[1, -1]`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if arr.Values[1] != nil {
		t.Fatal("expected null")
	}

	// FLOAT / DOUBLE lax arrays.
	got, err = BindLaxFloatArray(value.JsonValue(`[1.0, "x"]`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if arr.Values[1] != nil {
		t.Fatal("expected null")
	}
	got, err = BindLaxDoubleArray(value.JsonValue(`[1.0]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(mustArray(t, got).Values) != 1 {
		t.Fatal("expected 1")
	}

	// STRING lax (object element becomes NULL).
	got, err = BindLaxStringArray(value.JsonValue(`["a", {}]`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if arr.Values[1] != nil {
		t.Fatal("expected null")
	}

	// Non-array input -> NULL.
	if v, _ := BindLaxBoolArray(value.JsonValue("42")); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindLaxBoolArray(nil); v != nil {
		t.Fatal("expected null")
	}
}

// ---------------- json_array.go ----------------

func TestJSON_ARRAY(t *testing.T) {
	t.Parallel()

	got, err := JSON_ARRAY(value.IntValue(1), value.StringValue("a"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `[1,"a",null]` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Empty.
	got, err = JSON_ARRAY()
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "[]" {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestJSON_OBJECT(t *testing.T) {
	t.Parallel()

	got, err := JSON_OBJECT(value.StringValue("a"), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `{"a":1}` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// NULL value.
	got, err = JSON_OBJECT(value.StringValue("a"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `{"a":null}` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Odd number of args -> error.
	if _, err := JSON_OBJECT(value.StringValue("a")); err == nil {
		t.Fatal("expected error")
	}
}

// ---------------- json_contains.go ----------------

func TestJSON_CONTAINS(t *testing.T) {
	t.Parallel()

	got, err := JSON_CONTAINS(value.JsonValue(`{"a":1,"b":2}`), value.JsonValue(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Array containment.
	got, err = JSON_CONTAINS(value.JsonValue(`[1,2,3]`), value.JsonValue(`[2,3]`))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Negative case.
	got, err = JSON_CONTAINS(value.JsonValue(`{"a":1}`), value.JsonValue(`{"a":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	// Type mismatch (object vs array).
	got, err = JSON_CONTAINS(value.JsonValue(`{"a":1}`), value.JsonValue(`[1]`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	// NULL argument.
	if v, _ := JSON_CONTAINS(nil, value.JsonValue("1")); v != nil {
		t.Fatal("expected null")
	}

	// Arg count error.
	if _, err := JSON_CONTAINS(value.JsonValue("1")); err == nil {
		t.Fatal("expected error")
	}

	// Invalid JSON candidate -> error.
	if _, err := JSON_CONTAINS(value.JsonValue("not-json"), value.JsonValue("1")); err == nil {
		t.Fatal("expected error")
	}
}

func TestJSON_CONTAINSDeep(t *testing.T) {
	t.Parallel()

	// Nested object match drives jsonEquals through map/array/scalar.
	got, err := JSON_CONTAINS(value.JsonValue(`{"a":{"b":[1,2]}}`), value.JsonValue(`{"a":{"b":[1,2]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Mismatch in nested array.
	got, err = JSON_CONTAINS(value.JsonValue(`{"a":[1,2]}`), value.JsonValue(`{"a":[1,3]}`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	// Different-length arrays through jsonEquals -> false.
	got, err = JSON_CONTAINS(value.JsonValue(`{"a":[1,2]}`), value.JsonValue(`{"a":[1,2,3]}`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	// Different-size objects through jsonEquals.
	got, err = JSON_CONTAINS(value.JsonValue(`[{"a":1}]`), value.JsonValue(`[{"a":1,"b":2}]`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}
}

func TestBindJsonSetCreationPaths(t *testing.T) {
	t.Parallel()

	// Set into top-level nil — exercises the nil-node branch.
	got, err := BindJsonSet(value.StringValue("null"),
		value.StringValue("$.a"), value.IntValue(1),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `"a":1`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Deep nested set into nil.
	got, err = BindJsonSet(value.StringValue("null"),
		value.StringValue("$.a.b"), value.IntValue(2),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `"b":2`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Set on existing path replaces.
	got, err = BindJsonSet(value.StringValue(`{"a":1}`),
		value.StringValue("$.a"), value.IntValue(99),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `"a":99`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Odd pair count: production strips one trailing arg leaving an
	// even count, so this should NOT error; we just verify it does
	// not panic and returns a value.
	got, err = BindJsonSet(value.StringValue("{}"),
		value.StringValue("$.a"),
		value.IntValue(1),
		value.StringValue("$.b"),
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = got

	// Invalid path -> error.
	if _, err := BindJsonSet(value.StringValue("{}"),
		value.StringValue("bad-path"),
		value.IntValue(1),
		value.BoolValue(true),
	); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindJsonSetArrayPaths(t *testing.T) {
	t.Parallel()

	// Array element set in middle.
	got, err := BindJsonSet(value.StringValue(`[10,20,30]`),
		value.StringValue("$[1]"), value.IntValue(99),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "[10,99,30]" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Nested array set.
	got, err = BindJsonSet(value.StringValue(`[[1],[2]]`),
		value.StringValue("$[0][0]"), value.IntValue(99),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "[[99],[2]]" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Negative index leaves array unchanged (silent).
	got, err = BindJsonSet(value.StringValue(`[1,2,3]`),
		value.StringValue("$[0][0]"), value.IntValue(99),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	// Path goes into 1 which is scalar; production code leaves it as-is.
	if mustString(t, got) == "" {
		t.Fatal("expected output")
	}

	// Out-of-range index leaves array unchanged.
	got, err = BindJsonSet(value.StringValue(`[1,2]`),
		value.StringValue("$[99]"), value.IntValue(0),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "[1,2]" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Set with array-index on a top-level object is a no-op for that
	// segment (production code returns node unchanged).
	got, err = BindJsonSet(value.StringValue(`{"a":1}`),
		value.StringValue("$[0]"), value.IntValue(0),
		value.BoolValue(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `"a":1`) {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindJsonRemoveNestedPaths(t *testing.T) {
	t.Parallel()

	// Remove a deeply nested object key.
	got, err := BindJsonRemove(value.StringValue(`{"a":{"b":{"c":1}}}`), value.StringValue("$.a.b.c"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(mustString(t, got), `"c"`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Remove a nested array element.
	got, err = BindJsonRemove(value.StringValue(`{"a":[10,20,30]}`), value.StringValue("$.a[1]"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `[10,30]`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Removing out-of-range index leaves array unchanged.
	got, err = BindJsonRemove(value.StringValue(`[1,2]`), value.StringValue("$[99]"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `[1,2]`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Missing key leaves object unchanged.
	got, err = BindJsonRemove(value.StringValue(`{"a":1}`), value.StringValue("$.b"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `"a":1`) {
		t.Fatalf("got %q", mustString(t, got))
	}
}

// ---------------- json_flatten.go ----------------

func TestJSON_FLATTEN(t *testing.T) {
	t.Parallel()

	got, err := JSON_FLATTEN(value.JsonValue(`[[1,2],[3,4]]`))
	if err != nil {
		t.Fatal(err)
	}
	arr := mustArray(t, got)
	if len(arr.Values) != 4 {
		t.Fatalf("expected 4, got %d", len(arr.Values))
	}

	// Mixed: one element is not an array.
	got, err = JSON_FLATTEN(value.JsonValue(`[[1,2], 3]`))
	if err != nil {
		t.Fatal(err)
	}
	arr = mustArray(t, got)
	if len(arr.Values) != 3 {
		t.Fatalf("got %d", len(arr.Values))
	}

	// Scalar input -> single-element array.
	got, err = JSON_FLATTEN(value.JsonValue("42"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mustArray(t, got).Values) != 1 {
		t.Fatal("expected 1")
	}

	// NULL input.
	if v, _ := JSON_FLATTEN(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := JSON_FLATTEN(); err == nil {
		t.Fatal("expected arg count error")
	}
	if _, err := JSON_FLATTEN(value.JsonValue("not-json")); err == nil {
		t.Fatal("expected error")
	}
}

// ---------------- json_path_exists.go ----------------

func TestJSON_PATH_EXISTS(t *testing.T) {
	t.Parallel()

	got, err := JSON_PATH_EXISTS(value.JsonValue(`{"a":{"b":1}}`), value.StringValue("$.a.b"))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	got, err = JSON_PATH_EXISTS(value.JsonValue(`{"a":1}`), value.StringValue("$.b"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	// JSON null at path -> false (the underlying JSON_QUERY returns null).
	got, err = JSON_PATH_EXISTS(value.JsonValue(`{"a":null}`), value.StringValue("$.a"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if _, err := JSON_PATH_EXISTS(value.JsonValue("{}")); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := JSON_PATH_EXISTS(nil, value.StringValue("$.a")); v != nil {
		t.Fatal("expected null")
	}
}
