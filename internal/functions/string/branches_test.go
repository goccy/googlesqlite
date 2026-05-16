package string_test

import (
	"strings"
	"testing"

	strfn "github.com/goccy/googlesqlite/internal/functions/string"
	"github.com/goccy/googlesqlite/internal/value"
)

// TestRegexpExtractAllBytes drives REGEXP_EXTRACT_ALL's BYTES branch.
// The upstream BigQuery docs Example uses BYTES inputs to demonstrate
// the bytewise variant; the resulting array elements are BYTES of the
// last capture group.
func TestRegexpExtractAllBytes(t *testing.T) {
	t.Parallel()

	got, err := strfn.REGEXP_EXTRACT_ALL(value.BytesValue("a1b2c3"), `\d`)
	if err != nil {
		t.Fatal(err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 3 {
		t.Fatalf("len: %d", len(arr.Values))
	}
	if b, _ := arr.Values[0].ToBytes(); string(b) != "1" {
		t.Fatalf("element 0: %q", b)
	}
}

// TestRegexpExtractAllNullSkip exercises the NULL-skip path of
// BindRegexpExtractAll: any NULL arg short-circuits to NULL output.
func TestRegexpExtractAllNullSkip(t *testing.T) {
	t.Parallel()

	got, err := strfn.BindRegexpExtractAll(nil, value.StringValue(`\d`))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for NULL input, got %v", got)
	}

	got, err = strfn.BindRegexpExtractAll(value.StringValue("a1b2"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for NULL regex, got %v", got)
	}
}

// TestRegexpExtractAllBadRegex drives the regexp.Compile error path.
func TestRegexpExtractAllBadRegex(t *testing.T) {
	t.Parallel()

	if _, err := strfn.REGEXP_EXTRACT_ALL(value.StringValue("a"), "("); err == nil {
		t.Fatal("expected error for bad regex")
	}
}

// TestRegexpExtractAllUnsupportedType drives the type-rejection branch.
func TestRegexpExtractAllUnsupportedType(t *testing.T) {
	t.Parallel()

	if _, err := strfn.REGEXP_EXTRACT_ALL(value.IntValue(1), `\d`); err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

// TestRegexpExtractAllEmptyResult covers the "no matches" path —
// the function returns an empty ArrayValue, not nil.
func TestRegexpExtractAllEmptyResult(t *testing.T) {
	t.Parallel()

	got, err := strfn.REGEXP_EXTRACT_ALL(value.StringValue("abc"), `\d`)
	if err != nil {
		t.Fatal(err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 0 {
		t.Fatalf("expected empty array, got %v", arr.Values)
	}
}

// TestFormatFlagsAllAcceptedShapes drives every flag character that
// parseFormatFlag recognises (minus, plus, space, sharp, zero, quote).
// The expected outputs follow the BigQuery FORMAT() function
// reference: each flag affects the integer / float specifier as
// documented.
func TestFormatFlagsAcceptedShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		fmt  string
		args []value.Value
		want string
	}{
		// `0` flag pads with zeros for integers.
		{"zero_pad", "%05d", []value.Value{value.IntValue(7)}, "00007"},
		// `+` flag forces a leading sign for positive integers.
		{"plus_int", "%+d", []value.Value{value.IntValue(7)}, "+7"},
		// space flag adds a leading space for integers.
		{"space_int", "% d", []value.Value{value.IntValue(7)}, " 7"},
		// `'` flag inserts comma thousands-separator for integers.
		{"quote_int", "%'d", []value.Value{value.IntValue(1234567)}, "1,234,567"},
		{"quote_int_mod3", "%'d", []value.Value{value.IntValue(123)}, "123"},
		{"quote_int_4", "%'d", []value.Value{value.IntValue(1234)}, "1,234"},
		// `+` flag forces a leading sign on positive floats.
		{"plus_float", "%+f", []value.Value{value.FloatValue(1.5)}, "+1.500000"},
		// space flag on float.
		{"space_float", "% f", []value.Value{value.FloatValue(1.5)}, " 1.500000"},
		// width-from-arg with `*`.
		{"width_star", "%*d", []value.Value{value.IntValue(5), value.IntValue(7)}, "    7"},
		// width=0 (degenerate) round-trip.
		{"width_zero", "%0d", []value.Value{value.IntValue(7)}, "7"},
		// `i` is an alias for `d`.
		{"i_alias_d", "%i", []value.Value{value.IntValue(7)}, "7"},
		// `F` is an alias for `f`.
		{"F_alias_f", "%F", []value.Value{value.FloatValue(1.5)}, "1.500000"},
		// Width >0 with non-zero flag pads with spaces (float).
		{"width_float_space", "%6f", []value.Value{value.FloatValue(1.5)}, "1.500000"},
		// Width >0 with zero flag for float.
		{"width_float_zero", "%010.2f", []value.Value{value.FloatValue(1.5)}, "0000001.50"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			args := append([]value.Value{value.StringValue(c.fmt)}, c.args...)
			got, err := strfn.BindFormat(args...)
			if err != nil {
				t.Fatalf("BindFormat(%q): %v", c.fmt, err)
			}
			s, _ := got.ToString()
			if s != c.want {
				t.Fatalf("BindFormat(%q) = %q, want %q", c.fmt, s, c.want)
			}
		})
	}
}

// TestFormatFlagsRejectedCombinations covers the explicit
// "currently doesn't support" branches.
func TestFormatFlagsRejectedCombinations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		fmt  string
		args []value.Value
		want string // substring of expected error
	}{
		// `'` flag not supported for non-`d`/`i` specifier.
		{"quote_on_octal", "%'o", []value.Value{value.IntValue(7)}, "doesn't support"},
		// `-` flag not supported for integer.
		{"minus_on_int", "%-d", []value.Value{value.IntValue(7)}, "doesn't support"},
		// `#` flag not supported for integer.
		{"sharp_on_int", "%#d", []value.Value{value.IntValue(7)}, "doesn't support"},
		// `-` flag not supported for float.
		{"minus_on_float", "%-f", []value.Value{value.FloatValue(1)}, "doesn't support"},
		// `#` flag not supported for float.
		{"sharp_on_float", "%#f", []value.Value{value.FloatValue(1)}, "doesn't support"},
		// `'` flag not supported for float.
		{"quote_on_float", "%'f", []value.Value{value.FloatValue(1)}, "doesn't support"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			args := append([]value.Value{value.StringValue(c.fmt)}, c.args...)
			_, err := strfn.BindFormat(args...)
			if err == nil {
				t.Fatalf("expected error for %q", c.fmt)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err %q does not contain %q", err.Error(), c.want)
			}
		})
	}
}

// TestFormatWidthAndPrecisionFromArgs exercises the width / precision
// from-arg type-validation branches.
func TestFormatWidthAndPrecisionFromArgsTypeChecks(t *testing.T) {
	t.Parallel()

	// width via *: arg must be IntValue.
	if _, err := strfn.BindFormat(value.StringValue("%*d"), value.StringValue("nope"), value.IntValue(7)); err == nil {
		t.Fatal("expected width-type error")
	}
}

// TestFormatHexAndOctal drives both numeric bases.
func TestFormatHexAndOctal(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fmt  string
		v    int64
		want string
	}{
		{"%o", 8, "10"},
		{"%x", 255, "ff"},
		{"%X", 255, "FF"},
	}
	for _, c := range cases {
		got, err := strfn.BindFormat(value.StringValue(c.fmt), value.IntValue(c.v))
		if err != nil {
			t.Fatalf("BindFormat(%q): %v", c.fmt, err)
		}
		s, _ := got.ToString()
		if s != c.want {
			t.Fatalf("BindFormat(%q) = %q, want %q", c.fmt, s, c.want)
		}
	}
}

// TestFormatWidthAtMaxParseRejects feeds parseFormatWidth a value
// that overflows int64.
func TestFormatWidthOverflowRejected(t *testing.T) {
	t.Parallel()

	if _, err := strfn.BindFormat(value.StringValue("%99999999999999999999d"), value.IntValue(1)); err == nil {
		t.Fatal("expected overflow error from width parser")
	}
	if _, err := strfn.BindFormat(value.StringValue("%.99999999999999999999f"), value.FloatValue(1)); err == nil {
		t.Fatal("expected overflow error from precision parser")
	}
}

// TestFormatTSpecifier drives both %t and %T specifiers, which use
// parsePrintableString to call .Format on the value.
func TestFormatTSpecifier(t *testing.T) {
	t.Parallel()

	got, err := strfn.BindFormat(value.StringValue("%t"), value.IntValue(42))
	if err != nil {
		t.Fatal(err)
	}
	s, _ := got.ToString()
	if s != "42" {
		t.Fatalf("%%t int: %q", s)
	}
	got, err = strfn.BindFormat(value.StringValue("%T"), value.IntValue(42))
	if err != nil {
		t.Fatal(err)
	}
	s, _ = got.ToString()
	if s != "42" {
		t.Fatalf("%%T int: %q", s)
	}
}

// TestFormatScientificE drives %e / %E.
func TestFormatScientificNotation(t *testing.T) {
	t.Parallel()

	got, err := strfn.BindFormat(value.StringValue("%e"), value.FloatValue(1500.0))
	if err != nil {
		t.Fatal(err)
	}
	s, _ := got.ToString()
	if !strings.Contains(s, "e+03") && !strings.Contains(s, "E+03") {
		t.Fatalf("%%e: %q", s)
	}
	got, err = strfn.BindFormat(value.StringValue("%g"), value.FloatValue(1500.0))
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s == "" {
		t.Fatal("g specifier produced empty output")
	}
}

// TestLowerUpperBytesAndNil exercises the BYTES branch and the
// unsupported-type error path of LOWER/UPPER. The BigQuery docs
// describe LOWER on BYTES as a byte-wise lowercase conversion.
func TestLowerUpperBytesAndUnsupportedType(t *testing.T) {
	t.Parallel()

	// BYTES branch.
	got, err := strfn.LOWER(value.BytesValue("ABC"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBytes(); string(b) != "abc" {
		t.Fatalf("LOWER bytes: %q", b)
	}
	got, err = strfn.UPPER(value.BytesValue("abc"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBytes(); string(b) != "ABC" {
		t.Fatalf("UPPER bytes: %q", b)
	}

	// Nil input on LOWER returns nil (LOWER short-circuits nil).
	got, err = strfn.LOWER(nil)
	if err != nil || got != nil {
		t.Fatalf("LOWER(nil): %v / err=%v", got, err)
	}
	// BindUpper short-circuits nil at the bind layer.
	got, err = strfn.BindUpper(nil)
	if err != nil || got != nil {
		t.Fatalf("BindUpper(nil): %v / err=%v", got, err)
	}

	// Unsupported type.
	if _, err := strfn.LOWER(value.IntValue(1)); err == nil {
		t.Fatal("LOWER(int) should fail")
	}
	if _, err := strfn.UPPER(value.IntValue(1)); err == nil {
		t.Fatal("UPPER(int) should fail")
	}
}

// TestLPADBytesAndUnsupported drives the BYTES + unsupported-type
// branches of LPAD.
func TestLPADBytesAndUnsupported(t *testing.T) {
	t.Parallel()

	// BYTES branch with default pad (space).
	got, err := strfn.LPAD(value.BytesValue("abc"), 5, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBytes(); string(b) != "  abc" {
		t.Fatalf("LPAD bytes default: %q", b)
	}

	// BYTES branch with explicit pattern that needs repetition.
	got, err = strfn.LPAD(value.BytesValue("a"), 7, value.BytesValue("xy"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBytes(); string(b) != "xyxyxya" {
		t.Fatalf("LPAD bytes pattern: %q", b)
	}

	// BYTES branch: returnLength shorter than input truncates.
	got, err = strfn.LPAD(value.BytesValue("abcdef"), 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBytes(); string(b) != "abc" {
		t.Fatalf("LPAD bytes truncate: %q", b)
	}

	// String branch with pattern that requires repetition.
	got, err = strfn.LPAD(value.StringValue("a"), 7, value.StringValue("xy"))
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s != "xyxyxya" {
		t.Fatalf("LPAD string pattern: %q", s)
	}

	// Unsupported type.
	if _, err := strfn.LPAD(value.IntValue(1), 4, nil); err == nil {
		t.Fatal("LPAD int should fail")
	}

	// Bind with wrong arity.
	if _, err := strfn.BindLpad(value.StringValue("a")); err == nil {
		t.Fatal("BindLpad arity should fail")
	}
	// Bind with length not an int.
	if _, err := strfn.BindLpad(value.StringValue("a"), value.StringValue("x")); err == nil {
		t.Fatal("BindLpad length type should fail")
	}
	// Bind NULL short-circuits.
	got, err = strfn.BindLpad(nil, value.IntValue(3))
	if err != nil || got != nil {
		t.Fatalf("BindLpad nil: %v / err=%v", got, err)
	}
}

// TestRegexpReplaceBadRegexAndUnsupported drives the regexp.Compile
// error and the unsupported-type error of REGEXP_REPLACE for both
// the STRING and BYTES dispatch arms.
func TestRegexpReplaceBranches(t *testing.T) {
	t.Parallel()

	// Bad regex on STRING dispatch.
	if _, err := strfn.REGEXP_REPLACE(value.StringValue("a"), value.StringValue("("), value.StringValue("x")); err == nil {
		t.Fatal("expected bad regex error")
	}
	// Bad regex on BYTES dispatch.
	if _, err := strfn.REGEXP_REPLACE(value.BytesValue("a"), value.BytesValue("("), value.BytesValue("x")); err == nil {
		t.Fatal("expected bad regex error (bytes)")
	}
	// Unsupported type.
	if _, err := strfn.REGEXP_REPLACE(value.IntValue(1), value.StringValue("a"), value.StringValue("b")); err == nil {
		t.Fatal("expected unsupported-type error")
	}
}

// TestReplaceBranches drives the REPLACE BYTES branch + unsupported.
func TestReplaceBranches(t *testing.T) {
	t.Parallel()
	// BYTES happy path.
	got, err := strfn.REPLACE(value.BytesValue("abc"), value.BytesValue("b"), value.BytesValue("X"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBytes(); string(b) != "aXc" {
		t.Fatalf("got %q", b)
	}
	// Unsupported type.
	if _, err := strfn.REPLACE(value.IntValue(1), value.StringValue("a"), value.StringValue("b")); err == nil {
		t.Fatal("expected unsupported-type error")
	}
}

// TestRegexpInstrBranches drives REGEXP_INSTR position/occurrence
// validation + BYTES branch + bad regex.
func TestRegexpInstrBranches(t *testing.T) {
	t.Parallel()

	// occurrence=0 invalid.
	if _, err := strfn.REGEXP_INSTR(value.StringValue("foo"), value.StringValue("o"), 1, 0, 0); err == nil {
		t.Fatal("expected occurrence=0 error")
	}
	// position > input length: returns 0.
	got, err := strfn.REGEXP_INSTR(value.StringValue("foo"), value.StringValue("o"), 10, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 0 {
		t.Fatalf("expected 0, got %d", i)
	}
	// not enough occurrences: returns 0.
	got, err = strfn.REGEXP_INSTR(value.StringValue("foo"), value.StringValue("o"), 1, 99, 0)
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 0 {
		t.Fatalf("expected 0, got %d", i)
	}
	// BYTES happy path.
	got, err = strfn.REGEXP_INSTR(value.BytesValue("foobar"), value.BytesValue("o+"), 1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 2 {
		t.Fatalf("expected 2, got %d", i)
	}
	// BYTES bad regex.
	if _, err := strfn.REGEXP_INSTR(value.BytesValue("a"), value.BytesValue("("), 1, 1, 0); err == nil {
		t.Fatal("expected bad-regex error (bytes)")
	}
	// Unsupported type.
	if _, err := strfn.REGEXP_INSTR(value.IntValue(1), value.StringValue("a"), 1, 1, 0); err == nil {
		t.Fatal("expected unsupported-type error")
	}
}

// TestStartsEndsBytes drives the BYTES dispatch of STARTS_WITH /
// ENDS_WITH plus the LIKE error path.
func TestStartsEndsBytes(t *testing.T) {
	t.Parallel()

	got, err := strfn.STARTS_WITH(value.BytesValue("abc"), value.BytesValue("ab"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBool(); !b {
		t.Fatal("STARTS_WITH bytes should be true")
	}
	got, err = strfn.ENDS_WITH(value.BytesValue("abc"), value.BytesValue("bc"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBool(); !b {
		t.Fatal("ENDS_WITH bytes should be true")
	}
	// Unsupported type.
	if _, err := strfn.STARTS_WITH(value.IntValue(1), value.StringValue("x")); err == nil {
		t.Fatal("STARTS_WITH int should fail")
	}
	if _, err := strfn.ENDS_WITH(value.IntValue(1), value.StringValue("x")); err == nil {
		t.Fatal("ENDS_WITH int should fail")
	}
}

// TestFormatTrailingPercentRejected covers the trailing-% error path.
func TestFormatTrailingPercentRejected(t *testing.T) {
	t.Parallel()
	if _, err := strfn.BindFormat(value.StringValue("hello%")); err == nil {
		t.Fatal("expected trailing-% error")
	}
}
