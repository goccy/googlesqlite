package internal

import (
	"strings"
	"testing"
	gotime "time"

	"github.com/goccy/go-googlesql"

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

func TestCastTimestampToStringCanonicalForm(t *testing.T) {
	for _, tc := range []struct {
		in   gotime.Time
		want string
	}{
		{gotime.Date(2026, 8, 7, 22, 0, 0, 0, gotime.UTC), "2026-08-07 22:00:00+00"},
		{gotime.Date(2020, 1, 1, 0, 0, 0, 500000000, gotime.UTC), "2020-01-01 00:00:00.500+00"},
		{gotime.Date(2020, 1, 1, 0, 0, 0, 123400000, gotime.UTC), "2020-01-01 00:00:00.123400+00"},
		{gotime.Date(2020, 1, 1, 3, 0, 0, 0, gotime.FixedZone("", 3*3600)), "2020-01-01 00:00:00+00"},
	} {
		if got := castTimestampToString(tc.in); got != tc.want {
			t.Errorf("castTimestampToString(%v): got %q, want %q", tc.in, got, tc.want)
		}
		got, err := CastValue(m1(tf().MakeSimpleType(googlesql.TypeKindTypeString)), value.TimestampValue(tc.in))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.Value(value.StringValue(tc.want)) {
			t.Errorf("CastValue(STRING, TIMESTAMP %v): got %#v, want %q", tc.in, got, tc.want)
		}
	}
}
