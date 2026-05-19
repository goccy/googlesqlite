package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"
)

// TestRegression_AggregateOverEmptyGroup asserts that an aggregate
// evaluated over zero input rows returns each aggregate's documented
// zero value rather than NULL.
//
// Origin: aggregates were registered through ncruces'
// CreateAggregateFunction, whose iter.Seq wrapper starts the user
// coroutine lazily on the first Step. When a group has zero rows
// SQLite invokes only the final callback, which stops the
// never-started coroutine without ever running the aggregate, so the
// function never calls ctx.Result* and SQLite reports NULL. As a
// result COUNT(*) over an empty table returned NULL instead of 0.
// RegisterAggregator now registers through CreateWindowFunction, which
// constructs a fresh instance and calls Value even when Step never
// ran. See internal/sqlitex.RegisterAggregator.
func TestRegression_AggregateOverEmptyGroup(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=aggemptygroup")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE t (n INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	// COUNT-family aggregates return 0, not NULL, over an empty input.
	for _, tc := range []struct {
		name string
		sql  string
	}{
		{"count_star", "SELECT COUNT(*) FROM t"},
		{"count_column", "SELECT COUNT(n) FROM t"},
		{"count_star_where_false", "SELECT COUNT(*) FROM t WHERE FALSE"},
		{"countif", "SELECT COUNTIF(n > 0) FROM t"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got sql.NullInt64
			if err := conn.QueryRowContext(ctx, tc.sql).Scan(&got); err != nil {
				t.Fatalf("%s: %v", tc.sql, err)
			}
			if !got.Valid {
				t.Fatalf("%s returned NULL; want 0", tc.sql)
			}
			if got.Int64 != 0 {
				t.Fatalf("%s = %d; want 0", tc.sql, got.Int64)
			}
		})
	}

	// SUM/AVG/MIN/MAX over an empty input are legitimately NULL — the
	// fix must not turn those into a zero value.
	for _, tc := range []struct {
		name string
		sql  string
	}{
		{"sum", "SELECT SUM(n) FROM t"},
		{"avg", "SELECT AVG(n) FROM t"},
		{"min", "SELECT MIN(n) FROM t"},
		{"max", "SELECT MAX(n) FROM t"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got sql.NullInt64
			if err := conn.QueryRowContext(ctx, tc.sql).Scan(&got); err != nil {
				t.Fatalf("%s: %v", tc.sql, err)
			}
			if got.Valid {
				t.Fatalf("%s = %d; want NULL", tc.sql, got.Int64)
			}
		})
	}

	// A non-empty table still aggregates correctly, and GROUP BY (where
	// every emitted group has at least one row) is unaffected.
	if _, err := conn.ExecContext(ctx, "INSERT INTO t (n) VALUES (1), (2), (3)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	var count, sum sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*), SUM(n) FROM t").Scan(&count, &sum); err != nil {
		t.Fatalf("aggregate over filled table: %v", err)
	}
	if !count.Valid || count.Int64 != 3 {
		t.Fatalf("COUNT(*) over filled table = %v; want 3", count)
	}
	if !sum.Valid || sum.Int64 != 6 {
		t.Fatalf("SUM(n) over filled table = %v; want 6", sum)
	}
}
