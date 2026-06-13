package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"
)

// TestDropIfExistsMissingObject guards the bigquery-emulator regression
// where `DROP TABLE IF EXISTS <missing>` returned an error instead of
// succeeding as a no-op. The SQLite-level statement already carries IF
// EXISTS and runs cleanly, but DropStmtAction then unconditionally asked
// the catalog to delete the spec, which failed with "failed to find
// table spec from map" for an object that was never registered. IF
// EXISTS must make a missing object a no-op for TABLE, VIEW and FUNCTION.
func TestDropIfExistsMissingObject(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=drop_if_exists")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Dropping objects that were never created must succeed under IF EXISTS.
	for _, stmt := range []string{
		"DROP TABLE IF EXISTS never_created_table",
		"DROP VIEW IF EXISTS never_created_view",
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("%q on a missing object should be a no-op, got: %v", stmt, err)
		}
	}

	// And the IF EXISTS no-op must not wedge the connection: a real
	// create/drop round-trip afterwards still works.
	if _, err := db.ExecContext(ctx, "CREATE TABLE real_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE after no-op drop: %v", err)
	}
	// First drop removes it; a second IF EXISTS drop of the now-missing
	// table is again a no-op rather than an error.
	if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS real_t"); err != nil {
		t.Fatalf("DROP TABLE IF EXISTS of an existing table: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS real_t"); err != nil {
		t.Fatalf("second DROP TABLE IF EXISTS should be a no-op, got: %v", err)
	}

	// Without IF EXISTS, dropping a missing table must still be an error.
	if _, err := db.ExecContext(ctx, "DROP TABLE real_t"); err == nil {
		t.Fatal("DROP TABLE (no IF EXISTS) of a missing table should error, got nil")
	}
}
