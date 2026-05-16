// Unit tests for PARSE_NUMERIC / PARSE_BIGNUMERIC.
// Expected outputs follow the GoogleSQL numeric functions reference
// (docs/third_party/googlesql-docs/numbering_functions.md and
// conversion_functions.md examples for PARSE_NUMERIC).
package numeric_test

import (
	"math/big"
	"testing"

	numfn "github.com/goccy/googlesqlite/internal/functions/numeric"
	"github.com/goccy/googlesqlite/internal/value"
)

func TestParseNumeric(t *testing.T) {
	t.Parallel()

	got, err := numfn.BindParseNumeric(value.StringValue("123.45"))
	if err != nil {
		t.Fatalf("BindParseNumeric: %v", err)
	}
	nv, ok := got.(*value.NumericValue)
	if !ok {
		t.Fatalf("got %T; want *NumericValue", got)
	}
	if nv.IsBigNumeric {
		t.Errorf("PARSE_NUMERIC must produce NUMERIC (not BIGNUMERIC)")
	}
	if nv.Cmp(big.NewRat(2469, 20)) != 0 {
		t.Errorf("Rat = %s; want 123.45 (12345/100 = 2469/20)", nv.RatString())
	}

	// Negative value.
	got, err = numfn.BindParseNumeric(value.StringValue("-0.5"))
	if err != nil {
		t.Fatalf("neg: %v", err)
	}
	nv = got.(*value.NumericValue)
	if nv.Cmp(big.NewRat(-1, 2)) != 0 {
		t.Errorf("neg Rat = %s; want -1/2", nv.RatString())
	}

	// Invalid -> error.
	if _, err := numfn.BindParseNumeric(value.StringValue("not-a-number")); err == nil {
		t.Errorf("expected error for invalid input")
	}
}

func TestParseBigNumeric(t *testing.T) {
	t.Parallel()

	got, err := numfn.BindParseBigNumeric(value.StringValue("1e10"))
	if err != nil {
		t.Fatalf("BindParseBigNumeric: %v", err)
	}
	nv, ok := got.(*value.NumericValue)
	if !ok {
		t.Fatalf("got %T; want *NumericValue", got)
	}
	if !nv.IsBigNumeric {
		t.Errorf("PARSE_BIGNUMERIC must set IsBigNumeric=true")
	}
	want := new(big.Rat)
	want.SetString("1e10")
	if nv.Cmp(want) != 0 {
		t.Errorf("Rat = %s; want %s", nv.RatString(), want.RatString())
	}

	// Invalid -> error.
	if _, err := numfn.BindParseBigNumeric(value.StringValue("xxx")); err == nil {
		t.Errorf("expected error for invalid input")
	}
}
