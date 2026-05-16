package compliancetest

import (
	"strings"
	"testing"
)

func TestParseBasic(t *testing.T) {
	input := `[name=strings_basic]
SELECT "abc"
--
ARRAY<STRUCT<STRING>>[{"abc"}]
==
# Test all the accepted escaped characters.
[required_features=RADIANS_DEGREES_FUNCTIONS]
[name=radians_zero]
SELECT RADIANS(0);
--
ARRAY<STRUCT<DOUBLE>>[{0}]
==`
	cases, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}
	if cases[0].Name != "strings_basic" {
		t.Errorf("case 0 name = %q", cases[0].Name)
	}
	if !strings.Contains(cases[0].SQL, "SELECT \"abc\"") {
		t.Errorf("case 0 SQL = %q", cases[0].SQL)
	}
	if cases[1].Name != "radians_zero" {
		t.Errorf("case 1 name = %q", cases[1].Name)
	}
	if len(cases[1].Features) != 1 || cases[1].Features[0] != "RADIANS_DEGREES_FUNCTIONS" {
		t.Errorf("case 1 features = %v", cases[1].Features)
	}
}

func TestParseExpectedRowsScalar(t *testing.T) {
	rs, err := ParseExpectedRows(`ARRAY<STRUCT<STRING, INT64, DOUBLE, BOOL>>[{"abc", 42, 3.14, true}, {"xyz", -1, -0.5, false}]`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rs.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rs.Rows))
	}
	if rs.Rows[0][0] != "abc" || rs.Rows[0][1] != int64(42) || rs.Rows[0][2] != 3.14 || rs.Rows[0][3] != true {
		t.Errorf("row 0 mismatch: %v", rs.Rows[0])
	}
	if !ConvertibleToYAML(rs) {
		t.Errorf("should be convertible")
	}
}

func TestParseExpectedRowsArrayCellRaw(t *testing.T) {
	rs, err := ParseExpectedRows(`ARRAY<STRUCT<ARRAY<INT64>>>[{[1, 2, 3]}]`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ConvertibleToYAML(rs) {
		t.Errorf("array cell should not be convertible")
	}
}

func TestParseExpectedRowsNull(t *testing.T) {
	rs, err := ParseExpectedRows(`ARRAY<STRUCT<STRING>>[{NULL}]`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rs.Rows) != 1 || rs.Rows[0][0] != nil {
		t.Errorf("expected NULL, got %v", rs.Rows)
	}
}

func TestClassifyExpected(t *testing.T) {
	cases := []struct {
		in   string
		want ExpectedKind
	}{
		{`ARRAY<STRUCT<STRING>>[{"abc"}]`, ExpectedRows},
		{`ERROR: generic::out_of_range: x`, ExpectedError},
		{`{"abc"}`, ExpectedUnknown},
	}
	for _, c := range cases {
		got := ClassifyExpected(c.in)
		if got != c.want {
			t.Errorf("%q: got %v want %v", c.in, got, c.want)
		}
	}
}
