package googlesqlite_test

import (
	"context"
	"database/sql"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestTableFunctionSurvivesGC is the regression guard for bigquery-emulator
// TVF queries failing at random with "wasm trap", "No matching signature for
// table valued function", or a crashed process. The catalog kept only the
// embedded TableValuedFunction, so a GC finalized the FixedOutputSchemaTVF
// wrapper and freed the native TVF while the catalog still pointed at it.
func TestTableFunctionSurvivesGC(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=tvf_survives_gc")
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

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE Customer (CustomerId INT64, Name STRING)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO Customer (CustomerId, Name) VALUES (1, 'A'), (2, 'B'), (3, 'C'), (4, 'D')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE TABLE FUNCTION CustomerRange(MinId INT64, MaxId INT64) AS (
		SELECT *
		FROM Customer
		WHERE CustomerId >= MinId AND CustomerId <= MaxId
	)`); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION: %v", err)
	}

	// Two cycles, with a pause for the finalizer goroutine to run.
	for range 2 {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	rows, err := conn.QueryContext(ctx,
		"SELECT CustomerId, Name FROM CustomerRange(2, 3) ORDER BY CustomerId")
	if err != nil {
		t.Fatalf("SELECT FROM TVF: %v", err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var id int64
		var n string
		if err := rows.Scan(&id, &n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if strings.Join(got, ",") != "B,C" {
		t.Fatalf("rows = %v; want [B C]", got)
	}
}
