package googlesqlite_test

import (
	"context"
	"database/sql"
	"runtime"
	"testing"
	"time"

	_ "github.com/goccy/googlesqlite"
)

// TestTVFSurvivesGC calls a TVF after forcing garbage collection. The
// catalog must keep the TVF's native handle alive: if the Go wrapper
// that owns it is collected, its finalizer frees the handle while the
// catalog still points at it, and the next call reads freed memory
// (a SIGSEGV, or a garbled "No matching signature for table valued
// function : ." error).
func TestTVFSurvivesGC(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=tvf_survives_gc")
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

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE FUNCTION `p`.`d`.`plus_one`(k INT64) AS (SELECT k + 1 AS n)"); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION: %v", err)
	}
	for i := 0; i < 5; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond) // let finalizers run
		var n int64
		if err := conn.QueryRowContext(ctx, "SELECT n FROM `p`.`d`.`plus_one`(1)").Scan(&n); err != nil {
			t.Fatalf("call %d after GC: %v", i, err)
		}
		if n != 2 {
			t.Fatalf("call %d after GC: n = %d, want 2", i, n)
		}
	}
}
