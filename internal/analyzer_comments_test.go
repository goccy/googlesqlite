package internal

import "testing"

func TestBlankCommentsPreservesOffsetsAndLiterals(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"SELECT 1 -- it's\nFROM t", "SELECT 1        \nFROM t"},
		{"# PARTITION BY x\nSELECT 1", "                \nSELECT 1"},
		{"SELECT /* it's\nmulti */ 1", "SELECT        \n         1"},
		{"SELECT '-- a /* b */ #c' AS s", "SELECT '-- a /* b */ #c' AS s"},
		{`SELECT "it's" /* ' */, r'\d--', b"#" AS x`, `SELECT "it's"        , r'\d--', b"#" AS x`},
		{"SELECT `a--b` /*x*/ FROM `c#d`", "SELECT `a--b`       FROM `c#d`"},
		{"SELECT 1 /* unterminated", "SELECT 1 /* unterminated"},
	} {
		got := blankComments(tc.in)
		if got != tc.want {
			t.Errorf("blankComments(%q):\n got  %q\n want %q", tc.in, got, tc.want)
		}
		if len(got) != len(tc.in) {
			t.Errorf("blankComments(%q) changed the length: %d -> %d", tc.in, len(tc.in), len(got))
		}
	}
}
