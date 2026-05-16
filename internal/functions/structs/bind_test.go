// Unit tests for the Bind* surface of the structs package.
// Expected outputs follow the upstream GoogleSQL / BigQuery struct
// reference (docs/third_party/googlesql-docs/data_types.md and
// statistical_aggregate_functions.md examples).
package structs_test

import (
	"testing"

	structs "github.com/goccy/googlesqlite/internal/functions/structs"
	"github.com/goccy/googlesqlite/internal/value"
)

func TestMakeStructAndStructField(t *testing.T) {
	t.Parallel()

	// MAKE_STRUCT("a", 1, "b", "two") -> STRUCT(a:1, b:"two").
	sv, err := structs.BindMakeStruct(
		value.StringValue("a"), value.IntValue(1),
		value.StringValue("b"), value.StringValue("two"),
	)
	if err != nil {
		t.Fatalf("BindMakeStruct: %v", err)
	}
	s, ok := sv.(*value.StructValue)
	if !ok {
		t.Fatalf("MAKE_STRUCT result is %T; want *StructValue", sv)
	}
	if len(s.Keys) != 2 || s.Keys[0] != "a" || s.Keys[1] != "b" {
		t.Fatalf("Keys = %v; want [a b]", s.Keys)
	}
	if v, _ := s.Values[0].ToInt64(); v != 1 {
		t.Errorf("Values[0] = %v; want 1", s.Values[0])
	}
	if s.Values[1] != value.StringValue("two") {
		t.Errorf("Values[1] = %v; want two", s.Values[1])
	}
	if s.M["a"] != s.Values[0] || s.M["b"] != s.Values[1] {
		t.Errorf("M map and Values slice should match")
	}

	// MAKE_STRUCT with odd args -> error.
	if _, err := structs.BindMakeStruct(value.StringValue("a")); err == nil {
		t.Errorf("expected error on odd argument count")
	}

	// STRUCT_FIELD: index access.
	got, err := structs.BindStructField(s, value.IntValue(0))
	if err != nil {
		t.Fatalf("BindStructField[0]: %v", err)
	}
	if i, _ := got.ToInt64(); i != 1 {
		t.Errorf("[0] = %v; want 1", got)
	}
	got, err = structs.BindStructField(s, value.IntValue(1))
	if err != nil {
		t.Fatalf("BindStructField[1]: %v", err)
	}
	if got != value.StringValue("two") {
		t.Errorf("[1] = %v; want two", got)
	}
	// Out-of-range index -> nil (per BQ "missing struct field returns NULL").
	got, err = structs.BindStructField(s, value.IntValue(99))
	if err != nil {
		t.Fatalf("BindStructField[99]: %v", err)
	}
	if got != nil {
		t.Errorf("oob = %v; want nil", got)
	}
	// NULL input -> NULL.
	got, err = structs.BindStructField(nil, value.IntValue(0))
	if err != nil {
		t.Fatalf("BindStructField NULL: %v", err)
	}
	if got != nil {
		t.Errorf("NULL = %v; want nil", got)
	}
}

func TestStructWithFieldSet(t *testing.T) {
	t.Parallel()

	// Build STRUCT(a:1, b:"old", c:3).
	sv, err := structs.BindMakeStruct(
		value.StringValue("a"), value.IntValue(1),
		value.StringValue("b"), value.StringValue("old"),
		value.StringValue("c"), value.IntValue(3),
	)
	if err != nil {
		t.Fatalf("BindMakeStruct: %v", err)
	}

	// Replace field index 1 with "new".
	got, err := structs.BindStructWithFieldSet(sv, value.IntValue(1), value.StringValue("new"))
	if err != nil {
		t.Fatalf("BindStructWithFieldSet: %v", err)
	}
	s, ok := got.(*value.StructValue)
	if !ok {
		t.Fatalf("got %T; want *StructValue", got)
	}
	if len(s.Keys) != 3 || s.Keys[0] != "a" || s.Keys[1] != "b" || s.Keys[2] != "c" {
		t.Errorf("Keys = %v; want [a b c]", s.Keys)
	}
	if s.Values[1] != value.StringValue("new") {
		t.Errorf("Values[1] = %v; want new", s.Values[1])
	}
	// Verify map is in sync.
	if s.M["b"] != value.StringValue("new") {
		t.Errorf("M[b] = %v; want new", s.M["b"])
	}
	// Original struct must be untouched (defensive copy).
	if sv.(*value.StructValue).Values[1] != value.StringValue("old") {
		t.Errorf("original struct mutated: %v", sv.(*value.StructValue).Values[1])
	}

	// Out-of-range index returns the input unchanged.
	got, err = structs.BindStructWithFieldSet(sv, value.IntValue(99), value.StringValue("ignored"))
	if err != nil {
		t.Fatalf("oob: %v", err)
	}
	if got != sv.(*value.StructValue) && got.(*value.StructValue).Values[1] != value.StringValue("old") {
		t.Errorf("oob should leave struct unchanged")
	}

	// NULL input -> NULL.
	got, err = structs.BindStructWithFieldSet(nil, value.IntValue(0), value.StringValue("x"))
	if err != nil {
		t.Fatalf("nil input: %v", err)
	}
	if got != nil {
		t.Errorf("nil input = %v; want nil", got)
	}

	// Arity error.
	if _, err := structs.BindStructWithFieldSet(sv, value.IntValue(0)); err == nil {
		t.Errorf("expected arity error on <3 args")
	}
}
