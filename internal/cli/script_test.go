package cli

import (
	"reflect"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "simple",
			in:   "SELECT 1; SELECT 2;",
			want: []string{"SELECT 1", "SELECT 2"},
		},
		{
			name: "no trailing semicolon",
			in:   "SELECT 1",
			want: []string{"SELECT 1"},
		},
		{
			name: "semicolon inside single-quoted string",
			in:   `SELECT 'a;b'; SELECT 2`,
			want: []string{`SELECT 'a;b'`, "SELECT 2"},
		},
		{
			name: "semicolon inside double-quoted string",
			in:   `SELECT "x;y;z"`,
			want: []string{`SELECT "x;y;z"`},
		},
		{
			name: "escaped quote inside string",
			in:   `SELECT 'it\'s; fine'; SELECT 2`,
			want: []string{`SELECT 'it\'s; fine'`, "SELECT 2"},
		},
		{
			name: "raw string keeps backslash literal",
			in:   `SELECT r'a\'; SELECT 2`,
			want: []string{`SELECT r'a\'`, "SELECT 2"},
		},
		{
			name: "triple-quoted string with semicolons",
			in:   "SELECT '''a;\nb;c'''; SELECT 2",
			want: []string{"SELECT '''a;\nb;c'''", "SELECT 2"},
		},
		{
			name: "line comment with semicolon",
			in:   "SELECT 1 -- a;b\n; SELECT 2",
			want: []string{"SELECT 1 -- a;b", "SELECT 2"},
		},
		{
			name: "hash comment with semicolon",
			in:   "SELECT 1 # a;b\n; SELECT 2",
			want: []string{"SELECT 1 # a;b", "SELECT 2"},
		},
		{
			name: "block comment with semicolon",
			in:   "SELECT 1 /* a;b */; SELECT 2",
			want: []string{"SELECT 1 /* a;b */", "SELECT 2"},
		},
		{
			name: "backtick identifier with semicolon",
			in:   "SELECT `a;b` FROM t; SELECT 2",
			want: []string{"SELECT `a;b` FROM t", "SELECT 2"},
		},
		{
			name: "empty statements skipped",
			in:   "SELECT 1;;; ; SELECT 2;",
			want: []string{"SELECT 1", "SELECT 2"},
		},
		{
			name: "comment-only input",
			in:   "-- just a comment\n",
			want: nil,
		},
		{
			name: "group mode marker preserved",
			in:   `SELECT 1 \G`,
			want: []string{`SELECT 1 \G`},
		},
		{
			name: "bytes literal with semicolon",
			in:   `SELECT b'a;b'; SELECT 2`,
			want: []string{`SELECT b'a;b'`, "SELECT 2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitStatements(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitStatements(%q)\n got: %#v\nwant: %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsQueryStatement(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"SELECT 1", true},
		{"select 1", true},
		{"  WITH t AS (SELECT 1) SELECT * FROM t", true},
		{"(SELECT 1)", true},
		{"-- lead comment\nSELECT 1", true},
		{"/* c */ SELECT 1", true},
		{"FROM t |> WHERE x > 1", true},
		{"GRAPH g MATCH (n) RETURN n", true},
		{"CREATE TABLE t (x INT64)", false},
		{"INSERT INTO t VALUES (1)", false},
		{"UPDATE t SET x = 1", false},
		{"DELETE FROM t WHERE x = 1", false},
		{"DROP TABLE t", false},
	}
	for _, tt := range tests {
		if got := isQueryStatement(tt.in); got != tt.want {
			t.Errorf("isQueryStatement(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
