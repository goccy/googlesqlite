package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"
)

// TestTempTableScopedToFailedScript checks that a TEMP table created by a
// script that later fails does not outlive the script. BigQuery scopes temp
// tables to their script, so a later script on the same connection that
// creates a temp table of the same name must see the new definition.
func TestTempTableScopedToFailedScript(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=temp_table_scope_after_failure")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("db.Conn: %v", err)
	}
	defer conn.Close()

	// The failing script goes through both entry points: ExecContext and
	// QueryContext each own the cleanup of the statements they ran.
	if _, err := conn.ExecContext(ctx,
		"CREATE TEMP TABLE raw AS SELECT 1 AS a; SELECT no_such_col FROM raw",
	); err == nil {
		t.Fatal("ExecContext: first script should fail on the unknown column, got nil")
	}
	rows, err := conn.QueryContext(ctx,
		"CREATE TEMP TABLE raw AS SELECT 1 AS a; SELECT no_such_col FROM raw",
	)
	if err == nil {
		rows.Close()
		t.Fatal("QueryContext: first script should fail on the unknown column, got nil")
	}

	var b int64
	if err := conn.QueryRowContext(ctx,
		"CREATE TEMP TABLE raw AS SELECT 2 AS b; SELECT b FROM raw",
	).Scan(&b); err != nil {
		t.Fatalf("second script: %v", err)
	}
	if b != 2 {
		t.Fatalf("b = %d, want 2", b)
	}
}
