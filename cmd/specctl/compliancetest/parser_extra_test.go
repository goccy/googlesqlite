// Additional tests for the compliancetest parser.
// The fixtures mirror shapes seen in upstream GoogleSQL compliance
// *.test files (docs/third_party/googlesql-compliance/...) — comment-only
// blocks, attribute lines, error-shaped expected blocks, escape
// sequences, etc.
package compliancetest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile_HappyPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.test")
	body := `[name=basic]
SELECT 1
--
ARRAY<STRUCT<INT64>>[{1}]
==`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	cases, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(cases) != 1 || cases[0].Name != "basic" {
		t.Fatalf("got %+v; want one case named basic", cases)
	}
}

func TestParseFile_NonExistent(t *testing.T) {
	t.Parallel()
	if _, err := ParseFile(filepath.Join(t.TempDir(), "missing.test")); err == nil {
		t.Errorf("expected error opening missing file")
	}
}

// A case with comment + multiple required_features lines + arbitrary
// attributes + CRLF line endings is a common compliance-format shape.
func TestParseCase_CommentFeaturesAttrsCRLF(t *testing.T) {
	t.Parallel()
	body := "# leading comment line 1\r\n# line 2\r\n[required_features=A,B]\r\n[required_features=C]\r\n[name=multi]\r\n[parameters=@x int64]\r\nSELECT @x\r\n--\r\nARRAY<STRUCT<INT64>>[{1}]\r\n==\r\n"
	cases, err := Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(cases) != 1 {
		t.Fatalf("got %d cases; want 1", len(cases))
	}
	c := cases[0]
	if c.Name != "multi" {
		t.Errorf("Name = %q; want multi", c.Name)
	}
	if c.Attrs["parameters"] != "@x int64" {
		t.Errorf("Attrs[parameters] = %q; want @x int64", c.Attrs["parameters"])
	}
	if len(c.Features) != 3 || c.Features[0] != "A" || c.Features[2] != "C" {
		t.Errorf("Features = %v; want [A B C]", c.Features)
	}
	if !strings.Contains(c.Comment, "leading comment line 1") {
		t.Errorf("Comment missing expected content: %q", c.Comment)
	}
}

// A case missing the `--` separator must be silently dropped.
func TestParseCase_MissingSeparator(t *testing.T) {
	t.Parallel()
	body := "[name=broken]\nSELECT 1\nNo separator here\n==\n[name=ok]\nSELECT 2\n--\nARRAY<STRUCT<INT64>>[{2}]\n==\n"
	cases, err := Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(cases) != 1 || cases[0].Name != "ok" {
		t.Fatalf("got %+v; want one case named ok", cases)
	}
}

// A case with empty SQL body must be dropped, while a sibling
// well-formed case in the same file is still emitted.
func TestParseCase_EmptySQL(t *testing.T) {
	t.Parallel()
	body := "[name=empty]\n--\nARRAY<STRUCT<INT64>>[{1}]\n==\n[name=ok]\nSELECT 1\n--\nARRAY<STRUCT<INT64>>[{1}]\n=="
	cases, err := Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(cases) != 1 || cases[0].Name != "ok" {
		t.Errorf("expected only the ok case; got %+v", cases)
	}
}

// Attribute lines without an `=` are ignored without dropping the
// case (we still need its SQL and expected).
func TestParseCase_AttrWithoutEquals(t *testing.T) {
	t.Parallel()
	body := "[no_equals]\n[name=ok]\nSELECT 1\n--\nARRAY<STRUCT<INT64>>[{1}]\n=="
	cases, err := Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(cases) != 1 || cases[0].Name != "ok" {
		t.Fatalf("got %+v; want one case named ok", cases)
	}
}

// Parse over an empty stream returns no cases and no error.
func TestParse_Empty(t *testing.T) {
	t.Parallel()
	cases, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	if len(cases) != 0 {
		t.Errorf("expected 0 cases; got %d", len(cases))
	}
}

// Parse over a stream that has content but no parseable case returns
// an error ("no cases parsed").
func TestParse_MalformedNoCase(t *testing.T) {
	t.Parallel()
	if _, err := Parse(strings.NewReader("garbage line\nanother\n")); err == nil {
		t.Errorf("expected error for malformed content")
	}
}

// ClassifyExpected: extra blank-prefix should still classify correctly.
func TestClassifyExpected_WhitespacePrefix(t *testing.T) {
	t.Parallel()
	if got := ClassifyExpected("   ARRAY<STRUCT<INT64>>[{1}]"); got != ExpectedRows {
		t.Errorf("got %v; want ExpectedRows", got)
	}
	if got := ClassifyExpected("   ERROR: x"); got != ExpectedError {
		t.Errorf("got %v; want ExpectedError", got)
	}
}

// ParseExpectedRows error paths.
func TestParseExpectedRows_Errors(t *testing.T) {
	t.Parallel()
	if _, err := ParseExpectedRows("not an array<struct>"); err == nil {
		t.Errorf("expected error for non-ARRAY prefix")
	}
	if _, err := ParseExpectedRows("ARRAY<STRUCT<INT64>>{1, 2}"); err == nil {
		t.Errorf("expected error for missing [")
	}
	if _, err := ParseExpectedRows("ARRAY<STRUCT<INT64>>[{1}"); err == nil {
		t.Errorf("expected error for missing ]")
	}
}

// parseScalarCell: every recognised literal form.
func TestParseScalarCell_ScalarForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want any
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`null`, nil},
		{`NULL`, nil},
		{`true`, true},
		{`TRUE`, true},
		{`false`, false},
		{`FALSE`, false},
		{`-7`, int64(-7)},
		{`2.5`, 2.5},
	}
	for _, c := range cases {
		got := parseScalarCell(c.in)
		if got != c.want {
			t.Errorf("%q: got %v (%T); want %v (%T)", c.in, got, got, c.want, c.want)
		}
	}
	// NaN / Inf / hex byte literals are raw.
	for _, raw := range []string{"NaN", "inf", "-inf", `b"\x00"`, "[1,2]", "{a:1}"} {
		got := parseScalarCell(raw)
		if _, ok := got.(rawValue); !ok {
			t.Errorf("%q: got %v (%T); want rawValue", raw, got, got)
		}
	}
}

// ConvertibleToYAML rejects rows that include any rawValue cell.
func TestConvertibleToYAML_Mixed(t *testing.T) {
	t.Parallel()
	rs := ExpectedRowSet{Rows: [][]any{{int64(1), "abc"}, {int64(2), rawValue("[1,2]")}}}
	if ConvertibleToYAML(rs) {
		t.Errorf("rawValue in row should disqualify")
	}
}

// unquoteString must decode every recognised escape sequence.
func TestUnquoteString_AllEscapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{`abc`, "abc"},
		{`a\nb`, "a\nb"},
		{`a\tb`, "a\tb"},
		{`a\rb`, "a\rb"},
		{`a\"b`, `a"b`},
		{`a\'b`, "a'b"},
		{`a\\b`, `a\b`},
		{`a\0b`, "a\x00b"},
		{`a\x41b`, "aAb"},
		// Unknown escapes are preserved verbatim.
		{`a\zb`, `a\zb`},
	}
	for _, c := range cases {
		got, err := unquoteString(c.in)
		if err != nil {
			t.Errorf("%q: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %q; want %q", c.in, got, c.want)
		}
	}
	// Dangling backslash -> error.
	if _, err := unquoteString(`abc\`); err == nil {
		t.Errorf("expected error for dangling backslash")
	}
	// Truncated \x -> error.
	if _, err := unquoteString(`\x`); err == nil {
		t.Errorf("expected error for truncated \\x")
	}
	// Invalid hex -> error.
	if _, err := unquoteString(`\xZZ`); err == nil {
		t.Errorf("expected error for invalid hex")
	}
}

// ParseExpectedRows with a quoted string that contains a brace
// embedded inside must not confuse splitTopLevel.
func TestParseExpectedRows_StringContainsBrace(t *testing.T) {
	t.Parallel()
	rs, err := ParseExpectedRows(`ARRAY<STRUCT<STRING>>[{"a{b}c"}]`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(rs.Rows) != 1 || rs.Rows[0][0] != "a{b}c" {
		t.Errorf("got %v; want one row with a{b}c", rs.Rows)
	}
}

// ParseExpectedRows with an escaped quote inside a literal must not
// trip the in-string detector.
func TestParseExpectedRows_StringWithEscape(t *testing.T) {
	t.Parallel()
	rs, err := ParseExpectedRows(`ARRAY<STRUCT<STRING>>[{"a\"b"}]`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(rs.Rows) != 1 {
		t.Fatalf("got %d rows; want 1", len(rs.Rows))
	}
	// Note: parseScalarCell calls unquoteString on the literal body
	// `a\"b`, which yields `a"b`.
	if rs.Rows[0][0] != `a"b` {
		t.Errorf("got %q; want %q", rs.Rows[0][0], `a"b`)
	}
}
