// Unit tests for the Bind* surface exported by the string package.
// Expected outputs follow the upstream BigQuery / GoogleSQL string
// functions reference (and, where the function is BigQuery-specific,
// the bq docs Example section).

package string_test

import (
	"reflect"
	"testing"

	strfn "github.com/goccy/googlesqlite/internal/functions/string"
	"github.com/goccy/googlesqlite/internal/value"
)

// ----- Argument-count / NULL propagation matrix ---------------------

// arityCheck exercises the "wrong arg count" branch of every Bind*
// helper that enforces a fixed (or constrained) arity. Each entry is
// a list of arg counts the binder MUST reject. The check just
// confirms an error is returned; the per-binder success tests below
// confirm the happy path.
type arityCase struct {
	name    string
	bind    func(...value.Value) (value.Value, error)
	badArgs [][]value.Value
}

func TestBind_Arity(t *testing.T) {
	t.Parallel()

	cases := []arityCase{
		{"ASCII", strfn.BindAscii, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"BYTE_LENGTH", strfn.BindByteLength, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"CHAR_LENGTH", strfn.BindCharLength, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"CHR", strfn.BindChr, [][]value.Value{{}, {value.IntValue(1), value.IntValue(2)}}},
		{"LENGTH", strfn.BindLength, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"LOWER", strfn.BindLower, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"UPPER", strfn.BindUpper, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"UNICODE", strfn.BindUnicode, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"STARTS_WITH", strfn.BindStartsWith, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}}},
		{"ENDS_WITH", strfn.BindEndsWith, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}}},
		{"LEFT", strfn.BindLeft, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.IntValue(1), value.IntValue(2)}}},
		{"INSTR", strfn.BindInstr, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.StringValue("b"), value.IntValue(1), value.IntValue(2), value.IntValue(3)}}},
		{"SUBSTR", strfn.BindSubstr, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.IntValue(1), value.IntValue(1), value.IntValue(1)}}},
		{"LPAD", strfn.BindLpad, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.IntValue(1), value.StringValue("p"), value.StringValue("q")}}},
		{"COLLATE", strfn.BindCollate, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}}},
		{"TRANSLATE", strfn.BindTranslate, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.StringValue("b")}}},
		{"LTRIM", strfn.BindLtrim, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}}},
		{"TRIM", strfn.BindTrim, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}}},
		{"INITCAP", strfn.BindInitcap, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}}},
		{"NORMALIZE", strfn.BindNormalize, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("NFC"), value.StringValue("x")}}},
		{"NORMALIZE_AND_CASEFOLD", strfn.BindNormalizeAndCasefold, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("NFC"), value.StringValue("x")}}},
		{"CONCAT", strfn.BindConcat, [][]value.Value{{}}},
		{"FORMAT", strfn.BindFormat, [][]value.Value{{}}},
		{"CODE_POINTS_TO_BYTES", strfn.BindCodePointsToBytes, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"CODE_POINTS_TO_STRING", strfn.BindCodePointsToString, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"TO_CODE_POINTS", strfn.BindToCodePoints, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"FROM_BASE32", strfn.BindFromBase32, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"FROM_BASE64", strfn.BindFromBase64, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"FROM_HEX", strfn.BindFromHex, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"TO_BASE32", strfn.BindToBase32, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"TO_BASE64", strfn.BindToBase64, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"TO_HEX", strfn.BindToHex, [][]value.Value{{}, {value.StringValue("a"), value.StringValue("b")}}},
		{"STRPOS", strfn.BindStrpos, [][]value.Value{{value.StringValue("a")}, {value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, args := range tc.badArgs {
				if _, err := tc.bind(args...); err == nil {
					t.Errorf("%s: expected arity error for %d args", tc.name, len(args))
				}
			}
		})
	}
}

// nullCases checks that every binder follows the standard NULL-in =>
// NULL-out contract for at least one position. Note that SPLIT
// returns the empty array on NULL input (matching BQ semantics), and
// COALESCE-shaped helpers don't live in this package, so we handle
// those separately below.
func TestBind_NullPropagation(t *testing.T) {
	t.Parallel()

	type nullCase struct {
		name string
		bind func(...value.Value) (value.Value, error)
		args []value.Value
	}
	cases := []nullCase{
		{"ASCII", strfn.BindAscii, []value.Value{nil}},
		{"BYTE_LENGTH", strfn.BindByteLength, []value.Value{nil}},
		{"CHAR_LENGTH", strfn.BindCharLength, []value.Value{nil}},
		{"CHR", strfn.BindChr, []value.Value{nil}},
		{"LENGTH", strfn.BindLength, []value.Value{nil}},
		{"LOWER", strfn.BindLower, []value.Value{nil}},
		{"UPPER", strfn.BindUpper, []value.Value{nil}},
		{"UNICODE", strfn.BindUnicode, []value.Value{nil}},
		{"STARTS_WITH", strfn.BindStartsWith, []value.Value{nil, value.StringValue("x")}},
		{"ENDS_WITH", strfn.BindEndsWith, []value.Value{nil, value.StringValue("x")}},
		{"STRPOS", strfn.BindStrpos, []value.Value{nil, value.StringValue("x")}},
		{"LEFT", strfn.BindLeft, []value.Value{nil, value.IntValue(2)}},
		{"RIGHT", strfn.BindRight, []value.Value{nil, value.IntValue(2)}},
		{"LPAD", strfn.BindLpad, []value.Value{nil, value.IntValue(2)}},
		{"RPAD", strfn.BindRpad, []value.Value{nil, value.IntValue(2)}},
		{"LTRIM", strfn.BindLtrim, []value.Value{nil}},
		{"RTRIM", strfn.BindRtrim, []value.Value{nil}},
		{"TRIM", strfn.BindTrim, []value.Value{nil}},
		{"REGEXP_CONTAINS", strfn.BindRegexpContains, []value.Value{nil, value.StringValue(".")}},
		{"REGEXP_EXTRACT", strfn.BindRegexpExtract, []value.Value{nil, value.StringValue(".")}},
		{"REGEXP_EXTRACT_ALL", strfn.BindRegexpExtractAll, []value.Value{nil, value.StringValue(".")}},
		{"REGEXP_INSTR", strfn.BindRegexpInstr, []value.Value{nil, value.StringValue(".")}},
		{"REPEAT", strfn.BindRepeat, []value.Value{nil, value.IntValue(2)}},
		{"SOUNDEX", strfn.BindSoundex, []value.Value{nil}},
		{"SAFE_CONVERT_BYTES_TO_STRING", strfn.BindSafeConvertBytesToString, []value.Value{nil}},
		{"INITCAP", strfn.BindInitcap, []value.Value{nil}},
		{"NORMALIZE", strfn.BindNormalize, []value.Value{nil}},
		{"NORMALIZE_AND_CASEFOLD", strfn.BindNormalizeAndCasefold, []value.Value{nil}},
		{"CONCAT", strfn.BindConcat, []value.Value{nil, value.StringValue("x")}},
		{"FORMAT", strfn.BindFormat, []value.Value{nil}},
		{"CONTAINS_SUBSTR", strfn.BindContainsSubstr, []value.Value{nil, value.StringValue("x")}},
		{"EDIT_DISTANCE", strfn.BindEditDistance, []value.Value{nil, value.StringValue("x")}},
		{"LIKE", strfn.BindLike, []value.Value{nil, value.StringValue("%")}},
		{"SUBSTR", strfn.BindSubstr, []value.Value{nil, value.IntValue(1)}},
		{"INSTR", strfn.BindInstr, []value.Value{nil, value.StringValue("x")}},
		{"FROM_BASE32", strfn.BindFromBase32, []value.Value{nil}},
		{"FROM_BASE64", strfn.BindFromBase64, []value.Value{nil}},
		{"FROM_HEX", strfn.BindFromHex, []value.Value{nil}},
		{"TO_BASE32", strfn.BindToBase32, []value.Value{nil}},
		{"TO_BASE64", strfn.BindToBase64, []value.Value{nil}},
		{"TO_HEX", strfn.BindToHex, []value.Value{nil}},
		{"CODE_POINTS_TO_BYTES", strfn.BindCodePointsToBytes, []value.Value{nil}},
		{"CODE_POINTS_TO_STRING", strfn.BindCodePointsToString, []value.Value{nil}},
		{"TRANSLATE", strfn.BindTranslate, []value.Value{nil, value.StringValue("a"), value.StringValue("b")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.bind(tc.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != nil {
				t.Fatalf("expected NULL output for NULL input, got %v", got)
			}
		})
	}
}

// ----- ASCII / UNICODE / CHR / CHAR_LENGTH / LENGTH / BYTE_LENGTH ---

func TestASCII(t *testing.T) {
	t.Parallel()

	// From the BigQuery ASCII reference: ASCII('abcd') -> 97 (the
	// codepoint of 'a').
	got, err := strfn.BindAscii(value.StringValue("abcd"))
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 97 {
		t.Fatalf("ASCII('abcd'): got %d, want 97", i)
	}
	// Empty string returns 0 per the BigQuery reference.
	got, err = strfn.BindAscii(value.StringValue(""))
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 0 {
		t.Fatalf("ASCII(''): got %d, want 0", i)
	}
}

func TestUNICODE(t *testing.T) {
	t.Parallel()

	// UNICODE upstream Example: UNICODE('âbcd') -> 226.
	got, err := strfn.BindUnicode(value.StringValue("âbcd"))
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 226 {
		t.Fatalf("UNICODE('âbcd'): got %d, want 226", i)
	}
	got, err = strfn.BindUnicode(value.StringValue(""))
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 0 {
		t.Fatalf("UNICODE(''): got %d, want 0", i)
	}
}

func TestCHR(t *testing.T) {
	t.Parallel()

	// CHR(65) -> 'A' per the BQ docs Example.
	got, err := strfn.BindChr(value.IntValue(65))
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s != "A" {
		t.Fatalf("CHR(65): got %q, want \"A\"", s)
	}
	// CHR(0) -> '' per the BQ docs (the null code point produces an
	// empty STRING).
	got, err = strfn.BindChr(value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s != "" {
		t.Fatalf("CHR(0): got %q, want \"\"", s)
	}
}

func TestCHAR_LENGTH_LENGTH_BYTE_LENGTH(t *testing.T) {
	t.Parallel()

	// CHAR_LENGTH and LENGTH count Unicode code points; BYTE_LENGTH
	// counts bytes. BQ ref: LENGTH('абвгд') = 5, BYTE_LENGTH = 10.
	str := value.StringValue("абвгд")
	got, _ := strfn.BindCharLength(str)
	if i, _ := got.ToInt64(); i != 5 {
		t.Errorf("CHAR_LENGTH: got %d, want 5", i)
	}
	got, _ = strfn.BindLength(str)
	if i, _ := got.ToInt64(); i != 5 {
		t.Errorf("LENGTH(STRING): got %d, want 5", i)
	}
	got, _ = strfn.BindByteLength(str)
	if i, _ := got.ToInt64(); i != 10 {
		t.Errorf("BYTE_LENGTH: got %d, want 10", i)
	}
	// LENGTH on BYTES is bytes.
	got, _ = strfn.BindLength(value.BytesValue("absolute"))
	if i, _ := got.ToInt64(); i != 8 {
		t.Errorf("LENGTH(BYTES): got %d, want 8", i)
	}
	// LENGTH should reject other types.
	if _, err := strfn.BindLength(value.IntValue(1)); err == nil {
		t.Error("LENGTH on INT should fail")
	}
}

// ----- LOWER / UPPER / REVERSE / STARTS_WITH / ENDS_WITH / STRPOS ---

func TestLowerUpper(t *testing.T) {
	t.Parallel()

	// From the BQ reference: LOWER('FOO') -> 'foo'; UPPER('foo') ->
	// 'FOO'. The BYTES variants stay byte-for-byte ASCII-aware.
	if got, _ := strfn.BindLower(value.StringValue("FOO")); !equalString(got, "foo") {
		t.Errorf("LOWER STRING")
	}
	if got, _ := strfn.BindUpper(value.StringValue("foo")); !equalString(got, "FOO") {
		t.Errorf("UPPER STRING")
	}
	if got, _ := strfn.BindLower(value.BytesValue("FOO")); !equalBytes(got, []byte("foo")) {
		t.Errorf("LOWER BYTES")
	}
	if got, _ := strfn.BindUpper(value.BytesValue("foo")); !equalBytes(got, []byte("FOO")) {
		t.Errorf("UPPER BYTES")
	}
	// type rejection
	if _, err := strfn.BindLower(value.IntValue(1)); err == nil {
		t.Errorf("LOWER on INT should fail")
	}
	if _, err := strfn.BindUpper(value.IntValue(1)); err == nil {
		t.Errorf("UPPER on INT should fail")
	}
}

func TestReverse(t *testing.T) {
	t.Parallel()

	// REVERSE upstream Example: REVERSE('foo') -> 'oof'.
	// REVERSE is exposed as a non-Bind helper, so we call it directly.
	if got, _ := strfn.REVERSE(value.StringValue("foo")); !equalString(got, "oof") {
		t.Errorf("REVERSE STRING")
	}
	if got, _ := strfn.REVERSE(value.BytesValue{0x01, 0x02, 0x03}); !equalBytes(got, []byte{0x03, 0x02, 0x01}) {
		t.Errorf("REVERSE BYTES")
	}
	if _, err := strfn.REVERSE(value.IntValue(1)); err == nil {
		t.Errorf("REVERSE on INT should fail")
	}
}

func TestStartsWithEndsWith(t *testing.T) {
	t.Parallel()

	// STARTS_WITH('bar', 'b') -> TRUE per the BQ docs.
	got, _ := strfn.BindStartsWith(value.StringValue("bar"), value.StringValue("b"))
	if b, _ := got.ToBool(); !b {
		t.Errorf("STARTS_WITH('bar','b') expected TRUE")
	}
	// ENDS_WITH('apple', 'le') -> TRUE per the BQ docs.
	got, _ = strfn.BindEndsWith(value.StringValue("apple"), value.StringValue("le"))
	if b, _ := got.ToBool(); !b {
		t.Errorf("ENDS_WITH('apple','le') expected TRUE")
	}
	// BYTES variants
	got, _ = strfn.BindStartsWith(value.BytesValue("apple"), value.BytesValue("ap"))
	if b, _ := got.ToBool(); !b {
		t.Errorf("STARTS_WITH BYTES expected TRUE")
	}
}

func TestStrpos(t *testing.T) {
	t.Parallel()

	// STRPOS('foobar', 'bar') -> 4 per the BQ docs Example.
	got, _ := strfn.BindStrpos(value.StringValue("foobar"), value.StringValue("bar"))
	if i, _ := got.ToInt64(); i != 4 {
		t.Errorf("STRPOS: got %d, want 4", i)
	}
	// No match -> 0
	got, _ = strfn.BindStrpos(value.StringValue("foo"), value.StringValue("z"))
	if i, _ := got.ToInt64(); i != 0 {
		t.Errorf("STRPOS no match: got %d, want 0", i)
	}
	// BYTES variant
	got, _ = strfn.BindStrpos(value.BytesValue("foobar"), value.BytesValue("bar"))
	if i, _ := got.ToInt64(); i != 4 {
		t.Errorf("STRPOS BYTES: got %d, want 4", i)
	}
}

// ----- LEFT / RIGHT / LPAD / RPAD / SUBSTR / INSTR ----------------

func TestLeftRight(t *testing.T) {
	t.Parallel()

	// LEFT('apple', 3) -> 'app' per the BQ docs Example.
	if got, _ := strfn.BindLeft(value.StringValue("apple"), value.IntValue(3)); !equalString(got, "app") {
		t.Errorf("LEFT STRING")
	}
	// RIGHT('apple', 3) -> 'ple' per the BQ docs Example.
	if got, _ := strfn.BindRight(value.StringValue("apple"), value.IntValue(3)); !equalString(got, "ple") {
		t.Errorf("RIGHT STRING")
	}
	// length >= len(string) returns the original.
	if got, _ := strfn.BindLeft(value.StringValue("ap"), value.IntValue(5)); !equalString(got, "ap") {
		t.Errorf("LEFT too long")
	}
	// Negative length rejected
	if _, err := strfn.BindLeft(value.StringValue("ap"), value.IntValue(-1)); err == nil {
		t.Errorf("LEFT negative length should fail")
	}
	if _, err := strfn.BindRight(value.StringValue("ap"), value.IntValue(-1)); err == nil {
		t.Errorf("RIGHT negative length should fail")
	}
	// BYTES variant
	if got, _ := strfn.BindLeft(value.BytesValue("apple"), value.IntValue(3)); !equalBytes(got, []byte("app")) {
		t.Errorf("LEFT BYTES")
	}
}

func TestLpadRpad(t *testing.T) {
	t.Parallel()

	// LPAD('c', 5) -> '    c' per the BQ docs Example.
	if got, _ := strfn.BindLpad(value.StringValue("c"), value.IntValue(5)); !equalString(got, "    c") {
		t.Errorf("LPAD default pattern")
	}
	// LPAD with pattern: LPAD('c', 5, '-') -> '----c' per BQ docs.
	if got, _ := strfn.BindLpad(value.StringValue("c"), value.IntValue(5), value.StringValue("-")); !equalString(got, "----c") {
		t.Errorf("LPAD with pattern")
	}
	// RPAD('c', 5) -> 'c    '
	if got, _ := strfn.BindRpad(value.StringValue("c"), value.IntValue(5)); !equalString(got, "c    ") {
		t.Errorf("RPAD default pattern")
	}
	// RPAD with pattern: RPAD('c', 5, '-') -> 'c----'
	if got, _ := strfn.BindRpad(value.StringValue("c"), value.IntValue(5), value.StringValue("-")); !equalString(got, "c----") {
		t.Errorf("RPAD with pattern")
	}
	// If length <= len(str), truncate (per BQ Example: LPAD('abc',2)='ab').
	if got, _ := strfn.BindLpad(value.StringValue("abc"), value.IntValue(2)); !equalString(got, "ab") {
		t.Errorf("LPAD truncate")
	}
	// BYTES variant
	if got, _ := strfn.BindLpad(value.BytesValue("c"), value.IntValue(3)); !equalBytes(got, []byte("  c")) {
		t.Errorf("LPAD BYTES")
	}
}

func TestSubstr(t *testing.T) {
	t.Parallel()

	// SUBSTR('apple', 2, 2) -> 'pp' per the BQ docs Example.
	got, _ := strfn.BindSubstr(value.StringValue("apple"), value.IntValue(2), value.IntValue(2))
	if !equalString(got, "pp") {
		t.Errorf("SUBSTR with length: got %v", got)
	}
	// SUBSTR with negative pos: SUBSTR('apple', -2) -> 'le'.
	got, _ = strfn.BindSubstr(value.StringValue("apple"), value.IntValue(-2))
	if !equalString(got, "le") {
		t.Errorf("SUBSTR neg pos: got %v", got)
	}
	// SUBSTR with pos > len: empty.
	got, _ = strfn.BindSubstr(value.StringValue("apple"), value.IntValue(10))
	if !equalString(got, "") {
		t.Errorf("SUBSTR oob: got %v", got)
	}
	// negative length rejected
	if _, err := strfn.BindSubstr(value.StringValue("apple"), value.IntValue(1), value.IntValue(-1)); err == nil {
		t.Error("SUBSTR negative length should fail")
	}
}

func TestInstr(t *testing.T) {
	t.Parallel()

	// INSTR('banana', 'an') -> 2 (1-based offset of first occurrence).
	got, _ := strfn.BindInstr(value.StringValue("banana"), value.StringValue("an"))
	if i, _ := got.ToInt64(); i != 2 {
		t.Errorf("INSTR: got %d, want 2", i)
	}
	// 3rd arg = position, 4th = occurrence.
	got, _ = strfn.BindInstr(value.StringValue("banana"), value.StringValue("an"), value.IntValue(1), value.IntValue(2))
	if i, _ := got.ToInt64(); i != 4 {
		t.Errorf("INSTR 2nd occurrence: got %d, want 4", i)
	}
	// Not found -> 0
	got, _ = strfn.BindInstr(value.StringValue("banana"), value.StringValue("z"))
	if i, _ := got.ToInt64(); i != 0 {
		t.Errorf("INSTR not found: got %d, want 0", i)
	}
	// position=0 invalid
	if _, err := strfn.BindInstr(value.StringValue("b"), value.StringValue("a"), value.IntValue(0)); err == nil {
		t.Error("INSTR pos=0 should fail")
	}
	// occurrence=0 invalid
	if _, err := strfn.BindInstr(value.StringValue("b"), value.StringValue("a"), value.IntValue(1), value.IntValue(0)); err == nil {
		t.Error("INSTR occurrence=0 should fail")
	}
}

// ----- CONCAT ------------------------------------------------------

func TestConcat(t *testing.T) {
	t.Parallel()

	// CONCAT('Hello', ', ', 'World!') -> 'Hello, World!' per BQ docs.
	got, _ := strfn.BindConcat(value.StringValue("Hello"), value.StringValue(", "), value.StringValue("World!"))
	if !equalString(got, "Hello, World!") {
		t.Errorf("CONCAT strings: got %v", got)
	}
	// CONCAT BYTES variant.
	got, _ = strfn.BindConcat(value.BytesValue("ab"), value.BytesValue("cd"))
	if !equalBytes(got, []byte("abcd")) {
		t.Errorf("CONCAT bytes")
	}
}

// ----- TRIM / LTRIM / RTRIM --------------------------------------

func TestTrim(t *testing.T) {
	t.Parallel()

	// TRIM('   apple   ') -> 'apple' per BQ docs.
	if got, _ := strfn.BindTrim(value.StringValue("   apple   ")); !equalString(got, "apple") {
		t.Errorf("TRIM default")
	}
	// TRIM('***apple***', '*') -> 'apple' per BQ docs.
	if got, _ := strfn.BindTrim(value.StringValue("***apple***"), value.StringValue("*")); !equalString(got, "apple") {
		t.Errorf("TRIM with cutset")
	}
	// LTRIM examples
	if got, _ := strfn.BindLtrim(value.StringValue("   apple")); !equalString(got, "apple") {
		t.Errorf("LTRIM default")
	}
	if got, _ := strfn.BindLtrim(value.StringValue("***apple"), value.StringValue("*")); !equalString(got, "apple") {
		t.Errorf("LTRIM cutset")
	}
	// RTRIM examples
	if got, _ := strfn.BindRtrim(value.StringValue("apple   ")); !equalString(got, "apple") {
		t.Errorf("RTRIM default")
	}
	if got, _ := strfn.BindRtrim(value.StringValue("apple***"), value.StringValue("*")); !equalString(got, "apple") {
		t.Errorf("RTRIM cutset")
	}
	// BYTES variant
	if got, _ := strfn.BindTrim(value.BytesValue("  abc  ")); !equalBytes(got, []byte("abc")) {
		t.Errorf("TRIM BYTES")
	}
}

// ----- REPEAT / REPLACE -------------------------------------------

func TestRepeatReplace(t *testing.T) {
	t.Parallel()

	// REPEAT('abc', 3) -> 'abcabcabc' per BQ docs.
	got, _ := strfn.BindRepeat(value.StringValue("abc"), value.IntValue(3))
	if !equalString(got, "abcabcabc") {
		t.Errorf("REPEAT STRING")
	}
	// REPLACE('Hello World', 'World', 'BigQuery') -> 'Hello BigQuery'
	// per BQ docs. REPLACE is exposed as the bare helper (no Bind
	// wrapper).
	got, _ = strfn.REPLACE(value.StringValue("Hello World"), value.StringValue("World"), value.StringValue("BigQuery"))
	if !equalString(got, "Hello BigQuery") {
		t.Errorf("REPLACE STRING")
	}
	// BYTES variant.
	got, _ = strfn.REPLACE(value.BytesValue("abca"), value.BytesValue("a"), value.BytesValue("z"))
	if !equalBytes(got, []byte("zbcz")) {
		t.Errorf("REPLACE BYTES")
	}
	// Wrong type rejected.
	if _, err := strfn.REPLACE(value.IntValue(1), value.IntValue(1), value.IntValue(1)); err == nil {
		t.Errorf("REPLACE on INT should fail")
	}
}

// ----- TO/FROM_HEX, BASE64, BASE32 -------------------------------

func TestBaseAndHexEncodings(t *testing.T) {
	t.Parallel()

	// FROM_HEX('666f6f') -> b'foo' per BQ docs.
	got, _ := strfn.BindFromHex(value.StringValue("666f6f"))
	if !equalBytes(got, []byte("foo")) {
		t.Errorf("FROM_HEX")
	}
	// TO_HEX is the inverse.
	got, _ = strfn.BindToHex(value.BytesValue("foo"))
	if !equalString(got, "666f6f") {
		t.Errorf("TO_HEX")
	}
	// TO_BASE64 of b'foobar' -> 'Zm9vYmFy' per BQ docs.
	got, _ = strfn.BindToBase64(value.BytesValue("foobar"))
	if !equalString(got, "Zm9vYmFy") {
		t.Errorf("TO_BASE64")
	}
	// FROM_BASE64 round-trips.
	got, _ = strfn.BindFromBase64(value.StringValue("Zm9vYmFy"))
	if !equalBytes(got, []byte("foobar")) {
		t.Errorf("FROM_BASE64")
	}
	// TO_BASE32 round-trips.
	got, _ = strfn.BindToBase32(value.BytesValue("abc"))
	if got == nil {
		t.Errorf("TO_BASE32 nil")
	}
	// FROM_HEX with odd-length: implementation prepends '0'.
	got, err := strfn.BindFromHex(value.StringValue("a"))
	if err != nil {
		t.Fatalf("FROM_HEX odd: %v", err)
	}
	if !equalBytes(got, []byte{0x0a}) {
		t.Errorf("FROM_HEX odd-length")
	}
	// Invalid hex returns error.
	if _, err := strfn.BindFromHex(value.StringValue("zz")); err == nil {
		t.Error("FROM_HEX invalid")
	}
}

// ----- LIKE / REGEXP_* / CONTAINS_SUBSTR / EDIT_DISTANCE ---------

func TestLike(t *testing.T) {
	t.Parallel()

	// LIKE patterns: '%hat' matches 'cat in the hat'.
	got, _ := strfn.BindLike(value.StringValue("cat in the hat"), value.StringValue("%hat"))
	if b, _ := got.ToBool(); !b {
		t.Errorf("LIKE pattern match expected TRUE")
	}
	got, _ = strfn.BindLike(value.StringValue("cat"), value.StringValue("dog"))
	if b, _ := got.ToBool(); b {
		t.Errorf("LIKE non-match expected FALSE")
	}
}

func TestRegexpFamily(t *testing.T) {
	t.Parallel()

	// REGEXP_CONTAINS('foobar', 'oob') -> TRUE.
	got, _ := strfn.BindRegexpContains(value.StringValue("foobar"), value.StringValue("oob"))
	if b, _ := got.ToBool(); !b {
		t.Errorf("REGEXP_CONTAINS")
	}
	// REGEXP_EXTRACT('foobar', 'o+') -> 'oo' per the BQ docs Example.
	got, _ = strfn.BindRegexpExtract(value.StringValue("foobar"), value.StringValue("o+"))
	if !equalString(got, "oo") {
		t.Errorf("REGEXP_EXTRACT")
	}
	// REGEXP_EXTRACT_ALL('a1b2c3', '\\d') -> ['1','2','3'].
	got, _ = strfn.BindRegexpExtractAll(value.StringValue("a1b2c3"), value.StringValue(`\d`))
	arr, _ := got.ToArray()
	if len(arr.Values) != 3 {
		t.Errorf("REGEXP_EXTRACT_ALL: expected 3 elements, got %d", len(arr.Values))
	}
	// REGEXP_INSTR('foobar', 'o+') -> 2 per BQ docs.
	got, _ = strfn.BindRegexpInstr(value.StringValue("foobar"), value.StringValue("o+"))
	if i, _ := got.ToInt64(); i != 2 {
		t.Errorf("REGEXP_INSTR: got %d, want 2", i)
	}
	// REGEXP_INSTR position <= 0 invalid.
	if _, err := strfn.BindRegexpInstr(value.StringValue("foo"), value.StringValue("o"), value.IntValue(0)); err == nil {
		t.Error("REGEXP_INSTR pos=0 should fail")
	}
	// REGEXP_REPLACE upstream Example: REGEXP_REPLACE('# Heading',
	// '^# ([a-zA-Z0-9\\s]+$)', '<h1>\\1</h1>') -> '<h1>Heading</h1>'.
	// REGEXP_REPLACE is exposed without a Bind wrapper.
	got, _ = strfn.REGEXP_REPLACE(value.StringValue("foo"), value.StringValue("o+"), value.StringValue("X"))
	if !equalString(got, "fX") {
		t.Errorf("REGEXP_REPLACE: got %v", got)
	}
	// BYTES variant.
	got, _ = strfn.REGEXP_REPLACE(value.BytesValue("foo"), value.BytesValue("o+"), value.BytesValue("X"))
	if !equalBytes(got, []byte("fX")) {
		t.Errorf("REGEXP_REPLACE BYTES: got %v", got)
	}
	// Bad regex returns error.
	if _, err := strfn.BindRegexpContains(value.StringValue("a"), value.StringValue("(")); err == nil {
		t.Error("REGEXP bad expr should fail")
	}
}

func TestContainsSubstr(t *testing.T) {
	t.Parallel()

	// CONTAINS_SUBSTR('Foo Bar', 'oo b') -> TRUE (case-insensitive)
	// per the BQ docs.
	got, _ := strfn.BindContainsSubstr(value.StringValue("Foo Bar"), value.StringValue("oo b"))
	if b, _ := got.ToBool(); !b {
		t.Errorf("CONTAINS_SUBSTR case-insensitive")
	}
}

func TestEditDistance(t *testing.T) {
	t.Parallel()

	// EDIT_DISTANCE('a','b') -> 1 per the BQ docs Example.
	got, _ := strfn.BindEditDistance(value.StringValue("a"), value.StringValue("b"))
	if i, _ := got.ToInt64(); i != 1 {
		t.Errorf("EDIT_DISTANCE: got %d, want 1", i)
	}
	// 'kitten' vs 'sitting' = 3 per the standard Levenshtein example.
	got, _ = strfn.BindEditDistance(value.StringValue("kitten"), value.StringValue("sitting"))
	if i, _ := got.ToInt64(); i != 3 {
		t.Errorf("EDIT_DISTANCE kitten/sitting: got %d, want 3", i)
	}
	// One empty argument -> length of the other.
	got, _ = strfn.BindEditDistance(value.StringValue(""), value.StringValue("abc"))
	if i, _ := got.ToInt64(); i != 3 {
		t.Errorf("EDIT_DISTANCE empty: got %d", i)
	}
	got, _ = strfn.BindEditDistance(value.StringValue("xyz"), value.StringValue(""))
	if i, _ := got.ToInt64(); i != 3 {
		t.Errorf("EDIT_DISTANCE empty rhs: got %d", i)
	}
	// max_distance caps the result.
	got, _ = strfn.BindEditDistance(value.StringValue("kitten"), value.StringValue("sitting"), value.IntValue(2))
	if i, _ := got.ToInt64(); i != 2 {
		t.Errorf("EDIT_DISTANCE max_distance: got %d", i)
	}
	// < 2 args -> NULL.
	if got, err := strfn.BindEditDistance(value.StringValue("a")); err != nil || got != nil {
		t.Errorf("EDIT_DISTANCE single arg: got (%v,%v)", got, err)
	}
}

// ----- SAFE_CONVERT_BYTES_TO_STRING / SOUNDEX / INITCAP ----------

func TestSafeConvertSoundexInitcap(t *testing.T) {
	t.Parallel()

	// SAFE_CONVERT_BYTES_TO_STRING(b'abc') -> 'abc' per BQ docs.
	got, _ := strfn.BindSafeConvertBytesToString(value.BytesValue("abc"))
	if !equalString(got, "abc") {
		t.Errorf("SAFE_CONVERT")
	}
	// SOUNDEX('Ashcraft') -> 'A261' per BQ docs Example.
	got, _ = strfn.BindSoundex(value.StringValue("Ashcraft"))
	if !equalString(got, "A261") {
		t.Errorf("SOUNDEX Ashcraft: got %v", got)
	}
	// SOUNDEX('Robert') -> 'R163'.
	got, _ = strfn.BindSoundex(value.StringValue("Robert"))
	if !equalString(got, "R163") {
		t.Errorf("SOUNDEX Robert: got %v", got)
	}
	// SOUNDEX('') -> '' per BQ docs (no letters -> empty).
	got, _ = strfn.BindSoundex(value.StringValue(""))
	if !equalString(got, "") {
		t.Errorf("SOUNDEX empty")
	}
	// INITCAP('the quick brown fox') -> 'The Quick Brown Fox' (BQ docs).
	got, _ = strfn.BindInitcap(value.StringValue("the quick brown fox"))
	if !equalString(got, "The Quick Brown Fox") {
		t.Errorf("INITCAP default delim: got %v", got)
	}
	// INITCAP with custom delimiters (the BQ Example).
	got, _ = strfn.BindInitcap(value.StringValue("hello-world!foo"), value.StringValue("-!"))
	if !equalString(got, "Hello-World!Foo") {
		t.Errorf("INITCAP custom delim: got %v", got)
	}
}

// ----- TO_CODE_POINTS / CODE_POINTS_TO_STRING/_BYTES -------------

func TestCodePoints(t *testing.T) {
	t.Parallel()

	// TO_CODE_POINTS('foo') -> [102, 111, 111] per BQ docs.
	got, _ := strfn.BindToCodePoints(value.StringValue("foo"))
	arr, _ := got.ToArray()
	want := []int64{102, 111, 111}
	if len(arr.Values) != len(want) {
		t.Fatalf("TO_CODE_POINTS len: got %d, want %d", len(arr.Values), len(want))
	}
	for i, w := range want {
		if v, _ := arr.Values[i].ToInt64(); v != w {
			t.Errorf("TO_CODE_POINTS[%d]: got %d want %d", i, v, w)
		}
	}
	// CODE_POINTS_TO_STRING([97,98,99]) -> 'abc'.
	got, _ = strfn.BindCodePointsToString(&value.ArrayValue{Values: []value.Value{value.IntValue(97), value.IntValue(98), value.IntValue(99)}})
	if !equalString(got, "abc") {
		t.Errorf("CODE_POINTS_TO_STRING")
	}
	// CODE_POINTS_TO_BYTES([65, 66]) -> b'AB'.
	got, _ = strfn.BindCodePointsToBytes(&value.ArrayValue{Values: []value.Value{value.IntValue(65), value.IntValue(66)}})
	if !equalBytes(got, []byte("AB")) {
		t.Errorf("CODE_POINTS_TO_BYTES")
	}
	// TO_CODE_POINTS BYTES variant.
	got, _ = strfn.BindToCodePoints(value.BytesValue{0x01, 0x02})
	arr, _ = got.ToArray()
	if len(arr.Values) != 2 {
		t.Errorf("TO_CODE_POINTS BYTES len")
	}
	// TO_CODE_POINTS NULL -> empty array.
	got, _ = strfn.BindToCodePoints(nil)
	arr, _ = got.ToArray()
	if len(arr.Values) != 0 {
		t.Errorf("TO_CODE_POINTS NULL -> empty array")
	}
	// CODE_POINTS_TO_STRING NULL element propagates NULL.
	got, _ = strfn.BindCodePointsToString(&value.ArrayValue{Values: []value.Value{value.IntValue(97), nil}})
	if got != nil {
		t.Errorf("CODE_POINTS_TO_STRING NULL element: got %v", got)
	}
}

// ----- NORMALIZE / NORMALIZE_AND_CASEFOLD -----------------------

func TestNormalize(t *testing.T) {
	t.Parallel()

	// NORMALIZE('foo') -> 'foo' (NFC is the default mode).
	got, _ := strfn.BindNormalize(value.StringValue("foo"))
	if !equalString(got, "foo") {
		t.Errorf("NORMALIZE default")
	}
	// NORMALIZE with explicit NFD mode.
	got, _ = strfn.BindNormalize(value.StringValue("é"), value.StringValue("NFD"))
	if got == nil {
		t.Errorf("NORMALIZE NFD nil")
	}
	// Unknown mode -> error.
	if _, err := strfn.BindNormalize(value.StringValue("a"), value.StringValue("XYZ")); err == nil {
		t.Errorf("NORMALIZE bad mode should fail")
	}
	// NORMALIZE_AND_CASEFOLD lowercases.
	got, _ = strfn.BindNormalizeAndCasefold(value.StringValue("HELLO"))
	if !equalString(got, "hello") {
		t.Errorf("NORMALIZE_AND_CASEFOLD default")
	}
	// NFKC mode
	got, _ = strfn.BindNormalizeAndCasefold(value.StringValue("HELLO"), value.StringValue("NFKC"))
	if !equalString(got, "hello") {
		t.Errorf("NORMALIZE_AND_CASEFOLD NFKC")
	}
	// NFKD mode
	got, _ = strfn.BindNormalizeAndCasefold(value.StringValue("HELLO"), value.StringValue("NFKD"))
	if !equalString(got, "hello") {
		t.Errorf("NORMALIZE_AND_CASEFOLD NFKD")
	}
	// NFD mode
	got, _ = strfn.BindNormalizeAndCasefold(value.StringValue("HELLO"), value.StringValue("NFD"))
	if !equalString(got, "hello") {
		t.Errorf("NORMALIZE_AND_CASEFOLD NFD")
	}
	// Unknown mode -> error.
	if _, err := strfn.BindNormalizeAndCasefold(value.StringValue("a"), value.StringValue("XYZ")); err == nil {
		t.Errorf("NORMALIZE_AND_CASEFOLD bad mode should fail")
	}
	// Also exercise NFKC / NFKD on plain NORMALIZE.
	got, _ = strfn.BindNormalize(value.StringValue("a"), value.StringValue("NFKC"))
	if !equalString(got, "a") {
		t.Errorf("NORMALIZE NFKC")
	}
	got, _ = strfn.BindNormalize(value.StringValue("a"), value.StringValue("NFKD"))
	if !equalString(got, "a") {
		t.Errorf("NORMALIZE NFKD")
	}
}

// ----- TRANSLATE ------------------------------------------------

func TestTranslate(t *testing.T) {
	t.Parallel()

	// TRANSLATE('This is a cookie', 'sco', 'zku')
	// -> 'Thiz iz a kuukie' per BQ docs Example.
	got, _ := strfn.BindTranslate(value.StringValue("This is a cookie"), value.StringValue("sco"), value.StringValue("zku"))
	if !equalString(got, "Thiz iz a kuukie") {
		t.Errorf("TRANSLATE: got %v", got)
	}
	// Deletion case: source longer than target.
	got, _ = strfn.BindTranslate(value.StringValue("abc"), value.StringValue("ab"), value.StringValue("X"))
	if !equalString(got, "Xc") {
		t.Errorf("TRANSLATE deletion: got %v", got)
	}
	// Duplicate source character rejected.
	if _, err := strfn.BindTranslate(value.StringValue("abc"), value.StringValue("aa"), value.StringValue("xy")); err == nil {
		t.Errorf("TRANSLATE duplicate source must fail")
	}
	// Type mismatch (STRING expr + BYTES source) rejected.
	if _, err := strfn.BindTranslate(value.StringValue("abc"), value.BytesValue("a"), value.StringValue("x")); err == nil {
		t.Errorf("TRANSLATE type mismatch must fail")
	}
	// BYTES variant: replace 'a' with 'z'.
	got, _ = strfn.BindTranslate(value.BytesValue("abc"), value.BytesValue("a"), value.BytesValue("z"))
	if !equalBytes(got, []byte("zbc")) {
		t.Errorf("TRANSLATE BYTES")
	}
}

// ----- COLLATE --------------------------------------------------

func TestCollate(t *testing.T) {
	t.Parallel()

	// COLLATE with empty spec returns the input unchanged.
	got, _ := strfn.BindCollate(value.StringValue("Hello"), value.StringValue(""))
	if !equalString(got, "Hello") {
		t.Errorf("COLLATE empty spec")
	}
	// COLLATE with und:ci folds to lower per the implementation.
	got, _ = strfn.BindCollate(value.StringValue("Hello"), value.StringValue("und:ci"))
	if !equalString(got, "hello") {
		t.Errorf("COLLATE und:ci")
	}
	// COLLATE with und:cs returns as-is.
	got, _ = strfn.BindCollate(value.StringValue("Hello"), value.StringValue("und:cs"))
	if !equalString(got, "Hello") {
		t.Errorf("COLLATE und:cs")
	}
	// Unknown attribute -> error.
	if _, err := strfn.BindCollate(value.StringValue("a"), value.StringValue("und:xx")); err == nil {
		t.Error("COLLATE unknown attribute should fail")
	}
	// NULL value -> NULL.
	got, err := strfn.BindCollate(nil, value.StringValue(""))
	if err != nil || got != nil {
		t.Errorf("COLLATE NULL value: got (%v,%v)", got, err)
	}
	// NULL spec -> error.
	if _, err := strfn.BindCollate(value.StringValue("a"), nil); err == nil {
		t.Error("COLLATE NULL spec should fail")
	}
}

// ----- FORMAT --------------------------------------------------

func TestFormat(t *testing.T) {
	t.Parallel()

	// FORMAT('%d', 12) -> '12' per BQ docs Example.
	got, _ := strfn.BindFormat(value.StringValue("%d"), value.IntValue(12))
	if !equalString(got, "12") {
		t.Errorf("FORMAT %%d: got %v", got)
	}
	// FORMAT('%s', 'foo') -> 'foo' per BQ docs.
	got, _ = strfn.BindFormat(value.StringValue("%s"), value.StringValue("foo"))
	if !equalString(got, "foo") {
		t.Errorf("FORMAT %%s")
	}
	// FORMAT('Total: %d items', 7).
	got, _ = strfn.BindFormat(value.StringValue("Total: %d items"), value.IntValue(7))
	if !equalString(got, "Total: 7 items") {
		t.Errorf("FORMAT embed")
	}
	// FORMAT('%05d', 12) -> '00012' (width / zero-pad flag).
	got, _ = strfn.BindFormat(value.StringValue("%05d"), value.IntValue(12))
	if !equalString(got, "00012") {
		t.Errorf("FORMAT width-zero")
	}
	// FORMAT('%%') -> '%'.
	got, _ = strfn.BindFormat(value.StringValue("%%"))
	if !equalString(got, "%") {
		t.Errorf("FORMAT %%%%")
	}
}

func TestFormat_ExtendedSpecifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		fmtStr string
		args   []value.Value
		want   string
	}{
		// FORMAT supports the same conversion characters as Go's
		// fmt package, modulo a few BigQuery-specific tweaks. The
		// expected values mirror BigQuery's docs (rounding,
		// scientific notation, percent literal).
		{"hex_lower", "%x", []value.Value{value.IntValue(255)}, "ff"},
		{"hex_upper", "%X", []value.Value{value.IntValue(255)}, "FF"},
		{"octal", "%o", []value.Value{value.IntValue(8)}, "10"},
		{"float_simple", "%.2f", []value.Value{value.FloatValue(1.5)}, "1.50"},
		{"scientific_lower", "%e", []value.Value{value.FloatValue(1500.0)}, "1.500000e+03"},
		{"scientific_upper", "%E", []value.Value{value.FloatValue(1500.0)}, "1.500000E+03"},
		{"general_lower", "%g", []value.Value{value.FloatValue(1500.0)}, "1500"},
		// 't' / 'T' are printable-string conversions; for STRING
		// inputs they simply produce the string itself.
		{"printable_t", "%t", []value.Value{value.StringValue("foo")}, "foo"},
		{"printable_T", "%T", []value.Value{value.StringValue("foo")}, "\"foo\""},
		// String conversion of an integer via the analyzer's %s path.
		{"s_string", "%s", []value.Value{value.StringValue("xyz")}, "xyz"},
	}
	args0 := value.StringValue("")
	_ = args0
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bindArgs := append([]value.Value{value.StringValue(tc.fmtStr)}, tc.args...)
			got, err := strfn.BindFormat(bindArgs...)
			if err != nil {
				t.Fatalf("BindFormat(%q): %v", tc.fmtStr, err)
			}
			if !equalString(got, tc.want) {
				t.Fatalf("BindFormat(%q): got %v, want %q", tc.fmtStr, got, tc.want)
			}
		})
	}
}

func TestFormat_JSONSpecifiers(t *testing.T) {
	t.Parallel()

	// FORMAT('%p', JSON '{"a":1}') -> compacted one-line JSON.
	got, err := strfn.BindFormat(value.StringValue("%p"), value.JsonValue(`{ "a": 1 }`))
	if err != nil {
		t.Fatalf("FORMAT %%p: %v", err)
	}
	if !equalString(got, `{"a":1}`) {
		t.Errorf("FORMAT %%p: got %v", got)
	}
	// FORMAT('%P', JSON '{"a":1}') -> indented JSON.
	got, err = strfn.BindFormat(value.StringValue("%P"), value.JsonValue(`{"a":1}`))
	if err != nil {
		t.Fatalf("FORMAT %%P: %v", err)
	}
	if got == nil {
		t.Errorf("FORMAT %%P returned nil")
	}
	// %p requires JSON type.
	if _, err := strfn.BindFormat(value.StringValue("%p"), value.StringValue("a")); err == nil {
		t.Errorf("FORMAT %%p with non-JSON should fail")
	}
	if _, err := strfn.BindFormat(value.StringValue("%P"), value.StringValue("a")); err == nil {
		t.Errorf("FORMAT %%P with non-JSON should fail")
	}
}

func TestFormat_ValidationErrors(t *testing.T) {
	t.Parallel()

	// Argument-count mismatch.
	if _, err := strfn.BindFormat(value.StringValue("%d %d"), value.IntValue(1)); err == nil {
		t.Errorf("FORMAT arg-count mismatch should fail")
	}
	// Wrong type for %d (string instead of int).
	if _, err := strfn.BindFormat(value.StringValue("%d"), value.StringValue("x")); err == nil {
		t.Errorf("FORMAT %%d with string should fail")
	}
	// %o with negative.
	if _, err := strfn.BindFormat(value.StringValue("%o"), value.IntValue(-1)); err == nil {
		t.Errorf("FORMAT %%o negative should fail")
	}
}

// ----- SUBSTR additional pos cases ----------------------------

func TestSubstr_NegativePositions(t *testing.T) {
	t.Parallel()

	// pos < -len => start at 0 per the BQ docs (silently clipped).
	got, _ := strfn.BindSubstr(value.StringValue("apple"), value.IntValue(-100), value.IntValue(2))
	if !equalString(got, "ap") {
		t.Errorf("SUBSTR pos<-len: got %v", got)
	}
	// pos = 0 -> treated as 1 (start at beginning).
	got, _ = strfn.BindSubstr(value.StringValue("apple"), value.IntValue(0), value.IntValue(3))
	if !equalString(got, "app") {
		t.Errorf("SUBSTR pos=0: got %v", got)
	}
}

// ----- REGEXP_EXTRACT additional positions ------------------

func TestRegexpExtract_Positions(t *testing.T) {
	t.Parallel()

	// position past end -> NULL.
	got, _ := strfn.BindRegexpExtract(value.StringValue("foo"), value.StringValue("o"), value.IntValue(100))
	if got != nil {
		t.Errorf("REGEXP_EXTRACT past-end: got %v, want NULL", got)
	}
	// Bad position rejected.
	if _, err := strfn.BindRegexpExtract(value.StringValue("foo"), value.StringValue("o"), value.IntValue(0)); err == nil {
		t.Errorf("REGEXP_EXTRACT pos=0 must fail")
	}
	// Bad occurrence rejected.
	if _, err := strfn.BindRegexpExtract(value.StringValue("foo"), value.StringValue("o"), value.IntValue(1), value.IntValue(0)); err == nil {
		t.Errorf("REGEXP_EXTRACT occurrence=0 must fail")
	}
	// occurrence > matches -> NULL.
	got, _ = strfn.BindRegexpExtract(value.StringValue("foo"), value.StringValue("o"), value.IntValue(1), value.IntValue(99))
	if got != nil {
		t.Errorf("REGEXP_EXTRACT occurrence past-end: got %v, want NULL", got)
	}
	// REGEXP_EXTRACT BYTES variant.
	got, _ = strfn.BindRegexpExtract(value.BytesValue("foobar"), value.StringValue("o+"))
	if !equalBytes(got, []byte("oo")) {
		t.Errorf("REGEXP_EXTRACT BYTES: got %v", got)
	}
}

// ----- ENDS_WITH BYTES ---------------------------------------

func TestEndsWithBytes(t *testing.T) {
	t.Parallel()
	got, _ := strfn.BindEndsWith(value.BytesValue("apple"), value.BytesValue("le"))
	if b, _ := got.ToBool(); !b {
		t.Errorf("ENDS_WITH BYTES expected TRUE")
	}
}

// ----- Extra BYTES coverage for INSTR / REGEXP_INSTR / TRIM /
//        REPEAT / RIGHT / SUBSTR ---------------------------------

func TestInstrBytesAndNegativePos(t *testing.T) {
	t.Parallel()

	// INSTR BYTES Example.
	got, _ := strfn.BindInstr(value.BytesValue("banana"), value.BytesValue("an"))
	if i, _ := got.ToInt64(); i != 2 {
		t.Errorf("INSTR BYTES: got %d, want 2", i)
	}
	// INSTR with negative position (search from the right).
	got, _ = strfn.BindInstr(value.StringValue("banana"), value.StringValue("an"), value.IntValue(-1))
	if got == nil {
		t.Errorf("INSTR neg pos returned nil")
	}
	// position past length -> error.
	if _, err := strfn.BindInstr(value.StringValue("ab"), value.StringValue("a"), value.IntValue(100)); err == nil {
		t.Errorf("INSTR pos too large should fail")
	}
	// Different source / search types -> error.
	if _, err := strfn.BindInstr(value.StringValue("ab"), value.BytesValue("a")); err == nil {
		t.Errorf("INSTR mixed types should fail")
	}
	// INT source -> error.
	if _, err := strfn.BindInstr(value.IntValue(1), value.IntValue(1)); err == nil {
		t.Errorf("INSTR int should fail")
	}
}

func TestRegexpInstrExtended(t *testing.T) {
	t.Parallel()

	// REGEXP_INSTR BYTES variant.
	got, _ := strfn.BindRegexpInstr(value.BytesValue("foobar"), value.BytesValue("o+"))
	if i, _ := got.ToInt64(); i != 2 {
		t.Errorf("REGEXP_INSTR BYTES: got %d, want 2", i)
	}
	// occurrence past matches -> 0.
	got, _ = strfn.BindRegexpInstr(value.StringValue("foo"), value.StringValue("o"), value.IntValue(1), value.IntValue(99))
	if i, _ := got.ToInt64(); i != 0 {
		t.Errorf("REGEXP_INSTR occurrence past-end: got %d", i)
	}
	// Bad occurrence rejected.
	if _, err := strfn.BindRegexpInstr(value.StringValue("foo"), value.StringValue("o"), value.IntValue(1), value.IntValue(0)); err == nil {
		t.Errorf("REGEXP_INSTR occurrence=0 must fail")
	}
	// position past end -> 0.
	got, _ = strfn.BindRegexpInstr(value.StringValue("foo"), value.StringValue("o"), value.IntValue(100))
	if i, _ := got.ToInt64(); i != 0 {
		t.Errorf("REGEXP_INSTR past-end: got %d", i)
	}
	// 5-arg form: occurrence_position.
	got, _ = strfn.BindRegexpInstr(value.StringValue("foobar"), value.StringValue("(o+)"), value.IntValue(1), value.IntValue(1), value.IntValue(1))
	if got == nil {
		t.Errorf("REGEXP_INSTR 5-arg returned nil")
	}
}

func TestTrimAndPadBytesVariants(t *testing.T) {
	t.Parallel()

	// LTRIM BYTES default trims ASCII whitespace.
	if got, _ := strfn.BindLtrim(value.BytesValue("  abc")); !equalBytes(got, []byte("abc")) {
		t.Errorf("LTRIM BYTES default")
	}
	// LTRIM BYTES with cutset.
	if got, _ := strfn.BindLtrim(value.BytesValue("**abc"), value.StringValue("*")); !equalBytes(got, []byte("abc")) {
		t.Errorf("LTRIM BYTES cutset")
	}
	// RTRIM BYTES default and with cutset.
	if got, _ := strfn.BindRtrim(value.BytesValue("abc  ")); !equalBytes(got, []byte("abc")) {
		t.Errorf("RTRIM BYTES default")
	}
	if got, _ := strfn.BindRtrim(value.BytesValue("abc**"), value.StringValue("*")); !equalBytes(got, []byte("abc")) {
		t.Errorf("RTRIM BYTES cutset")
	}
	// RPAD BYTES variants.
	if got, _ := strfn.BindRpad(value.BytesValue("c"), value.IntValue(5), value.BytesValue("-")); !equalBytes(got, []byte("c----")) {
		t.Errorf("RPAD BYTES with pattern")
	}
	if got, _ := strfn.BindRpad(value.BytesValue("c"), value.IntValue(5)); !equalBytes(got, []byte("c    ")) {
		t.Errorf("RPAD BYTES default pattern")
	}
	// RPAD STRING with pattern shorter than remainder (triggers repeat).
	if got, _ := strfn.BindRpad(value.StringValue("c"), value.IntValue(7), value.StringValue("ab")); !equalString(got, "cababab") {
		t.Errorf("RPAD STRING with multi-char pattern")
	}
}

func TestRepeatBytesAndSubstrBytes(t *testing.T) {
	t.Parallel()

	// REPEAT BYTES Example: REPEAT(b'abc', 2) -> b'abcabc'.
	got, _ := strfn.BindRepeat(value.BytesValue("abc"), value.IntValue(2))
	if !equalBytes(got, []byte("abcabc")) {
		t.Errorf("REPEAT BYTES: got %v", got)
	}
	// REPEAT on INT should fail.
	if _, err := strfn.BindRepeat(value.IntValue(1), value.IntValue(2)); err == nil {
		t.Errorf("REPEAT INT should fail")
	}
	// SUBSTR BYTES.
	got, _ = strfn.BindSubstr(value.BytesValue("apple"), value.IntValue(2), value.IntValue(2))
	if !equalBytes(got, []byte("pp")) {
		t.Errorf("SUBSTR BYTES")
	}
	// SUBSTR on INT should fail.
	if _, err := strfn.BindSubstr(value.IntValue(1), value.IntValue(1)); err == nil {
		t.Errorf("SUBSTR INT should fail")
	}
}

func TestRightBytesAndTrimCutsetBytes(t *testing.T) {
	t.Parallel()

	// RIGHT BYTES variant.
	if got, _ := strfn.BindRight(value.BytesValue("apple"), value.IntValue(3)); !equalBytes(got, []byte("ple")) {
		t.Errorf("RIGHT BYTES")
	}
	// RIGHT BYTES length too large.
	if got, _ := strfn.BindRight(value.BytesValue("ab"), value.IntValue(10)); !equalBytes(got, []byte("ab")) {
		t.Errorf("RIGHT BYTES too long")
	}
	// TRIM BYTES with cutset.
	if got, _ := strfn.BindTrim(value.BytesValue("**abc**"), value.BytesValue("*")); !equalBytes(got, []byte("abc")) {
		t.Errorf("TRIM BYTES cutset")
	}
	// LPAD STRING with multi-char pattern (needs repeat).
	if got, _ := strfn.BindLpad(value.StringValue("c"), value.IntValue(7), value.StringValue("ab")); !equalString(got, "abababc") {
		t.Errorf("LPAD STRING with multi-char pattern: got %v", got)
	}
}

// ----- ENDS_WITH / STRPOS rejection for wrong types ----------

func TestEndsWithStrposBadTypes(t *testing.T) {
	t.Parallel()
	// STRPOS with INT -> error.
	if _, err := strfn.BindStrpos(value.IntValue(1), value.IntValue(1)); err == nil {
		t.Errorf("STRPOS on INT should fail")
	}
	if _, err := strfn.BindStartsWith(value.IntValue(1), value.IntValue(1)); err == nil {
		t.Errorf("STARTS_WITH on INT should fail")
	}
	if _, err := strfn.BindEndsWith(value.IntValue(1), value.IntValue(1)); err == nil {
		t.Errorf("ENDS_WITH on INT should fail")
	}
}

// ----- SPLIT --------------------------------------------------

func TestSplit(t *testing.T) {
	t.Parallel()

	// SPLIT('a,b,c') -> ['a','b','c'] (default delim is comma per BQ).
	got, _ := strfn.BindSplit(value.StringValue("a,b,c"))
	arr, _ := got.ToArray()
	if len(arr.Values) != 3 {
		t.Errorf("SPLIT default delim len: got %d", len(arr.Values))
	}
	// SPLIT with explicit delim.
	got, _ = strfn.BindSplit(value.StringValue("a|b|c"), value.StringValue("|"))
	arr, _ = got.ToArray()
	if len(arr.Values) != 3 {
		t.Errorf("SPLIT explicit delim len: got %d", len(arr.Values))
	}
	// SPLIT NULL -> empty array, per the BQ NULL contract for SPLIT.
	got, _ = strfn.BindSplit(nil)
	arr, _ = got.ToArray()
	if len(arr.Values) != 0 {
		t.Errorf("SPLIT NULL -> empty array")
	}
	// SPLIT BYTES requires explicit delim.
	got, _ = strfn.BindSplit(value.BytesValue("a|b"), value.BytesValue("|"))
	arr, _ = got.ToArray()
	if len(arr.Values) != 2 {
		t.Errorf("SPLIT BYTES len: got %d", len(arr.Values))
	}
	if _, err := strfn.BindSplit(value.BytesValue("a|b")); err == nil {
		t.Errorf("SPLIT BYTES no-delim should fail")
	}
}

// ----- Base32 round-trip ---------------------------------------

func TestBase32RoundTrip(t *testing.T) {
	t.Parallel()
	// FROM_BASE32 / TO_BASE32 are documented as inverses. Use the
	// canonical RFC 4648 example: 'foobar' <-> 'MZXW6YTBOI======'.
	encoded, _ := strfn.BindToBase32(value.BytesValue("foobar"))
	if !equalString(encoded, "MZXW6YTBOI======") {
		t.Errorf("TO_BASE32 foobar: got %v", encoded)
	}
	decoded, _ := strfn.BindFromBase32(value.StringValue("MZXW6YTBOI======"))
	if !equalBytes(decoded, []byte("foobar")) {
		t.Errorf("FROM_BASE32 foobar: got %v", decoded)
	}
}

// ----- REGEXP_REPLACE backreferences ---------------------------

// TestRegexpReplaceBackreferences pins the BigQuery replacement grammar
// (string_functions.md#regexp_replace): \0..\9 are single-digit group
// references (\0 = the whole match), \\ is a literal backslash, '$' is
// an ordinary literal, and any other escape is an error. Expected
// values are the upstream Description/Examples plus the documented
// single-digit and literal-backslash rules.
func TestRegexpReplaceBackreferences(t *testing.T) {
	t.Parallel()

	cases := []struct {
		desc    string
		input   string
		expr    string
		repl    string
		want    string
		wantErr bool
	}{
		// BQ docs Example: REGEXP_REPLACE('# Heading',
		// '^# ([a-zA-Z0-9\s]+$)', '<h1>\1</h1>') -> '<h1>Heading</h1>'.
		{"heading_example", "# Heading", `^# ([a-zA-Z0-9\s]+$)`, `<h1>\1</h1>`, "<h1>Heading</h1>", false},
		// Description example: REGEXP_REPLACE('abc','b(.)','X\1') -> aXc.
		// The backreference is the final token of the replacement.
		{"trailing_ref", "abc", "b(.)", `X\1`, "aXc", false},
		// \0 is the entire match.
		{"whole_match", "abc", "b(.)", `[\0]`, "a[bc]", false},
		// Single-digit index: \10 is group 1 then a literal '0'.
		{"single_digit", "abcdefghijk", "(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)", `\10`, "a0k", false},
		// Consecutive references.
		{"double_ref", "xyz", "(y)", `\1\1`, "xyyz", false},
		// \\ -> one literal backslash.
		{"literal_backslash", "abc", "(b)", `\\`, `a\c`, false},
		// '$' is a literal, not the Expand sigil.
		{"literal_dollar", "abc", "(b)", `p$q`, "ap$qc", false},
		{"literal_dollar_brace", "abc", "(b)", `${1}`, "a${1}c", false},
		// Invalid escapes: '\' must be followed by a digit or '\'.
		{"backslash_nondigit", "abc", "(b)", `\q`, "", true},
		{"trailing_backslash", "abc", "(b)", `x\`, "", true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got, err := strfn.REGEXP_REPLACE(value.StringValue(c.input), value.StringValue(c.expr), value.StringValue(c.repl))
			if c.wantErr {
				if err == nil {
					t.Fatalf("REGEXP_REPLACE(%q,%q,%q): expected error, got %v", c.input, c.expr, c.repl, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("REGEXP_REPLACE(%q,%q,%q): %v", c.input, c.expr, c.repl, err)
			}
			if !equalString(got, c.want) {
				t.Fatalf("REGEXP_REPLACE(%q,%q,%q) = %v, want %q", c.input, c.expr, c.repl, got, c.want)
			}
		})
	}
}

// ----- helpers --------------------------------------------------

func equalString(v value.Value, want string) bool {
	if v == nil {
		return false
	}
	got, err := v.ToString()
	if err != nil {
		return false
	}
	return got == want
}

func equalBytes(v value.Value, want []byte) bool {
	if v == nil {
		return false
	}
	got, err := v.ToBytes()
	if err != nil {
		return false
	}
	return reflect.DeepEqual(got, want)
}
