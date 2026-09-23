package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestFormatTLiteralTypes covers FORMAT('%T', ...) for the types whose
// literal carries a type prefix or a specific form, per the "%t and %T
// behavior" table of the FORMAT reference: NUMERIC "1.5", BIGNUMERIC,
// JSON '...', INTERVAL "..." YEAR TO SECOND, and the one-field and
// zero-field STRUCT forms, also nested inside arrays and structs. A JSON
// or STRING literal takes single quotes only when that avoids escaping a
// double quote. The scalar expected values are measured on BigQuery; the
// DATETIME cases inside arrays and structs follow from the element rule.
func TestFormatTLiteralTypes(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=format_t_literal_types")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		query string
		want  string
	}{
		{"numeric", `SELECT FORMAT('%T', NUMERIC '1.5')`, `NUMERIC "1.5"`},
		{"negative numeric", `SELECT FORMAT('%T', NUMERIC '-1.5')`, `NUMERIC "-1.5"`},
		{"numeric %t unchanged", `SELECT FORMAT('%t', NUMERIC '1.5')`, `1.5`},
		{"bignumeric", `SELECT FORMAT('%T', BIGNUMERIC '1.5')`, `BIGNUMERIC "1.5"`},
		{"json", `SELECT FORMAT('%T', JSON '{"name":"apple","stock":3}')`, `JSON '{"name":"apple","stock":3}'`},
		{"json %t unchanged", `SELECT FORMAT('%t', JSON '{"name":"apple","stock":3}')`, `{"name":"apple","stock":3}`},
		{"interval", `SELECT FORMAT('%T', INTERVAL '1-2 3 4:5:6.789' YEAR TO SECOND)`, `INTERVAL "1-2 3 4:5:6.789" YEAR TO SECOND`},
		{"interval %t unchanged", `SELECT FORMAT('%t', INTERVAL '1-2 3 4:5:6.789' YEAR TO SECOND)`, `1-2 3 4:5:6.789`},
		{"date", `SELECT FORMAT('%T', DATE '2011-02-03')`, `DATE "2011-02-03"`},
		{"timestamp", `SELECT FORMAT('%T', TIMESTAMP '2011-02-03 04:05:06+00')`, `TIMESTAMP "2011-02-03 04:05:06+00"`},
		{"bytes", `SELECT FORMAT('%T', b'abc\x01\x02')`, `b"abc\x01\x02"`},
		{"numeric in array", `SELECT FORMAT('%T', [NUMERIC '1.5', NUMERIC '2'])`, `[NUMERIC "1.5", NUMERIC "2"]`},
		{"numeric in struct", `SELECT FORMAT('%T', STRUCT(NUMERIC '1.5' AS a, 'x' AS b))`, `(NUMERIC "1.5", "x")`},
		{"one-field struct", `SELECT FORMAT('%T', STRUCT(NUMERIC '1.5' AS a))`, `STRUCT(NUMERIC "1.5")`},
		{"one-field struct %t unchanged", `SELECT FORMAT('%t', STRUCT(NUMERIC '1.5' AS a))`, `(1.5)`},
		{"json string value with both quote kinds", `SELECT FORMAT('%T', JSON '"it\'s"')`, `JSON "\"it's\""`},
		{"string with a double quote", `SELECT FORMAT('%T', 'a"b')`, `'a"b'`},
		{"string with both quote kinds", `SELECT FORMAT('%T', 'a"b\'c')`, `"a\"b'c"`},
		{"datetime", `SELECT FORMAT('%T', DATETIME '2011-02-03 04:05:06')`, `DATETIME "2011-02-03 04:05:06"`},
		{"datetime with fraction", `SELECT FORMAT('%T', DATETIME '2011-02-03 04:05:06.123')`, `DATETIME "2011-02-03 04:05:06.123"`},
		{"time", `SELECT FORMAT('%T', TIME '12:34:56')`, `TIME "12:34:56"`},
		{"numeric integral", `SELECT FORMAT('%T', NUMERIC '2')`, `NUMERIC "2"`},
		{"zero-field struct", `SELECT FORMAT('%T', STRUCT())`, `STRUCT()`},
		{"string without quotes", `SELECT FORMAT('%T', 'abc')`, `"abc"`},
		{"datetime %t", `SELECT FORMAT('%t', DATETIME '2011-02-03 04:05:06')`, `2011-02-03 04:05:06`},
		{"datetime %t in array", `SELECT FORMAT('%t', [DATETIME '2011-02-03 04:05:06'])`, `[2011-02-03 04:05:06]`},
		{"datetime %t in struct", `SELECT FORMAT('%t', STRUCT(DATETIME '2011-02-03 04:05:06' AS a, 1 AS b))`, `(2011-02-03 04:05:06, 1)`},
		{"datetime in array", `SELECT FORMAT('%T', [DATETIME '2011-02-03 04:05:06'])`, `[DATETIME "2011-02-03 04:05:06"]`},
		{"json in array", `SELECT FORMAT('%T', [JSON '1'])`, `[JSON "1"]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			if err := db.QueryRowContext(ctx, tc.query).Scan(&got); err != nil {
				t.Fatalf("%s: %v", tc.query, err)
			}
			if got != tc.want {
				t.Errorf("%s = %q, want %q", tc.query, got, tc.want)
			}
		})
	}
}
