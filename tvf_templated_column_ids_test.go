package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestTVFTemplatedBodyColumnIDs calls a templated (ANY TABLE) TVF whose
// body has an unaliased aggregate from a query that aggregates too. The
// body is analyzed separately at each call site, so its column ids
// restart and `$agg1#4` inside it is also the id of the caller's
// COUNT(*). Formatting the body must not consume the caller's columns.
func TestTVFTemplatedBodyColumnIDs(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=tvf_templated_column_ids")
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

	for _, stmt := range []string{
		"CREATE TABLE ta AS SELECT v FROM UNNEST([3, 1]) AS v",
		"CREATE TABLE tb AS SELECT v FROM UNNEST([2]) AS v",
		"CREATE TABLE FUNCTION mins(a ANY TABLE, b ANY TABLE) AS (SELECT MIN(v) AS m FROM a UNION ALL SELECT MIN(v) FROM b)",
		`CREATE TABLE FUNCTION first_min(a ANY TABLE, b ANY TABLE) AS (
		  WITH firsts AS (SELECT MIN(v) AS m FROM a UNION ALL SELECT MIN(v) FROM b)
		  SELECT m FROM (SELECT MIN(m) AS m FROM firsts) WHERE m IS NOT NULL)`,
	} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}

	for _, tc := range []struct {
		query string
		want  int64
	}{
		{"SELECT COUNT(*) FROM mins(TABLE ta, TABLE tb)", 2},
		{"SELECT SUM(m) FROM mins(TABLE ta, TABLE tb)", 3},
		{"SELECT (SELECT COUNT(*) FROM first_min(TABLE ta, TABLE tb))", 1},
		{"SELECT (SELECT m FROM first_min(TABLE ta, TABLE tb))", 1},
	} {
		var got int64
		if err := conn.QueryRowContext(ctx, tc.query).Scan(&got); err != nil {
			t.Errorf("%s: %v", tc.query, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.query, got, tc.want)
		}
	}
}
