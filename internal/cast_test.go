package internal

import (
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// CAST() and bindCast() carry defensive error paths that the analyzer's
// upstream signature checks prevent from firing through SQL. Drive them
// directly here so a future refactor that re-routes those paths still
// has a regression signal.

// TestCAST_NilFromType exercises the "ToGoogleSQLType failed" branch on
// the source type by passing a nil *Type.
func TestCAST_NilFromType(t *testing.T) {
	if _, err := CAST(value.IntValue(1), nil, &Type{Name: "INT64"}, false); err == nil {
		t.Fatalf("expected error from CAST with nil fromType")
	}
}

// TestCAST_NilToType exercises the symmetric "ToGoogleSQLType failed"
// branch on the destination type.
func TestCAST_NilToType(t *testing.T) {
	if _, err := CAST(value.IntValue(1), &Type{Name: "INT64"}, nil, false); err == nil {
		t.Fatalf("expected error from CAST with nil toType")
	}
}

// TestBindCast_ArityErrors covers the up-front argument-count check in
// bindCast.
func TestBindCast_ArityErrors(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 5} {
		args := make([]value.Value, n)
		for i := range args {
			args[i] = value.StringValue("")
		}
		if _, err := bindCast(args...); err == nil {
			t.Fatalf("expected error for %d args", n)
		}
	}
}

// TestBindCast_BadFromTypeJSON covers the json.Unmarshal failure path
// for the fromType slot.
func TestBindCast_BadFromTypeJSON(t *testing.T) {
	args := []value.Value{
		value.IntValue(1),
		value.StringValue("{ malformed json"),
		value.StringValue(`{"Name":"INT64"}`),
		value.BoolValue(false),
	}
	_, err := bindCast(args...)
	if err == nil {
		t.Fatalf("expected JSON parse error for fromType")
	}
}

// TestBindCast_BadToTypeJSON covers the json.Unmarshal failure path
// for the toType slot.
func TestBindCast_BadToTypeJSON(t *testing.T) {
	args := []value.Value{
		value.IntValue(1),
		value.StringValue(`{"Name":"INT64"}`),
		value.StringValue("not-json"),
		value.BoolValue(false),
	}
	_, err := bindCast(args...)
	if err == nil {
		t.Fatalf("expected JSON parse error for toType")
	}
}

// TestBindCast_HappyPath confirms a well-formed bindCast call succeeds
// — guards against the JSON / ToString paths erroring on valid input.
func TestBindCast_HappyPath(t *testing.T) {
	args := []value.Value{
		value.IntValue(42),
		value.StringValue(`{"Name":"INT64"}`),
		value.StringValue(`{"Name":"INT64"}`),
		value.BoolValue(false),
	}
	got, err := bindCast(args...)
	if err != nil {
		t.Fatalf("bindCast: %v", err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "42") {
		t.Fatalf("expected result containing 42, got %q", s)
	}
}
