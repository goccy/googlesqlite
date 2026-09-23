package googlesqlite_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestAssertStatement covers ASSERT expression [AS description], per
// https://cloud.google.com/bigquery/docs/reference/standard-sql/debugging-statements#assert
// ("If expression evaluates to FALSE or NULL, the statement generates an
// error. If AS description is present, description will appear in the
// error message."). ASSERT used to be a no-op, so every one of these
// false or NULL conditions passed silently.
func TestAssertStatement(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=assert_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE t AS SELECT x FROM UNNEST([1, 2, 3]) AS x"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	for _, tc := range []struct {
		name    string
		stmt    string
		wantErr string // empty: the statement must succeed
	}{
		{"true literal", "ASSERT TRUE AS 'unused'", ""},
		{"false literal", "ASSERT FALSE AS 'false must fail'", "false must fail"},
		{"null", "ASSERT CAST(NULL AS BOOL) AS 'null must fail'", "null must fail"},
		{"no description", "ASSERT 1 = 2", "Assertion failed"},
		{"table subquery holds", "ASSERT (SELECT COUNT(*) FROM t) = 3 AS 'three rows'", ""},
		{"table subquery fails", "ASSERT (SELECT COUNT(*) FROM t) = 4 AS 'four rows'", "four rows"},
		{"subquery yields NULL", "ASSERT (SELECT MAX(x) FROM t WHERE x > 9) = 1 AS 'empty max is NULL'", "empty max is NULL"},
		// The two examples of the reference page.
		{"docs example passes", "ASSERT ((SELECT COUNT(*) > 5 FROM UNNEST([1, 2, 3, 4, 5, 6]))) AS 'Table must contain more than 5 rows.'", ""},
		{"docs example fails", "ASSERT ((SELECT COUNT(*) > 10 FROM UNNEST([1, 2, 3, 4, 5, 6]))) AS 'Table must contain more than 10 rows.'", "Table must contain more than 10 rows."},
		{"exists passes", "ASSERT EXISTS((SELECT X FROM UNNEST([7877, 7879, 7883, 7901, 7907]) AS X WHERE X = 7907)) AS 'Column X must contain the value 7907.'", ""},
		{"exists fails", "ASSERT EXISTS((SELECT X FROM UNNEST([7877, 7879, 7883, 7901, 7907]) AS X WHERE X = 7919)) AS 'Column X must contain the value 7919'", "Column X must contain the value 7919"},
	} {
		_, err := conn.ExecContext(ctx, tc.stmt)
		switch {
		case tc.wantErr == "" && err != nil:
			t.Errorf("%s: %s: unexpected error %v", tc.name, tc.stmt, err)
		case tc.wantErr != "" && err == nil:
			t.Errorf("%s: %s: succeeded; want an error containing %q", tc.name, tc.stmt, tc.wantErr)
		case tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr):
			t.Errorf("%s: %s: error %q does not contain %q", tc.name, tc.stmt, err, tc.wantErr)
		}
	}
}

// TestAssertStopsScript checks that a failed ASSERT aborts the rest of a
// multi-statement script, so a later statement never runs.
func TestAssertStopsScript(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=assert_stops_script")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "CREATE TABLE before_assert AS SELECT 1 AS x; ASSERT FALSE AS 'stop here'; CREATE TABLE after_assert AS SELECT 1 AS x")
	if err == nil || !strings.Contains(err.Error(), "stop here") {
		t.Fatalf("script error = %v; want one containing %q", err, "stop here")
	}
	if _, err := conn.ExecContext(ctx, "SELECT * FROM before_assert"); err != nil {
		t.Fatalf("statement before the ASSERT did not run: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "SELECT * FROM after_assert"); err == nil {
		t.Fatal("statement after the failed ASSERT ran")
	}
}
