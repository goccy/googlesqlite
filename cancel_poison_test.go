package googlesqlite_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/goccy/googlesqlite"
)

// TestCancelledTxDoesNotPoisonPooledConn is the regression guard for the
// bigquery-emulator #478 "connection is already closed" / "driver: bad
// connection" cascade.
//
// The emulator runs each request as db.Conn(ctx) + BeginTx(ctx) +
// query(ctx). Before the fix, BeginTx handed the request context to the
// inner *sql.Conn's transaction; when the request was cancelled
// mid-statement database/sql's tx watcher ran Rollback(discardConn=true)
// and closed the inner conn asynchronously. Our own operation had
// already returned context.Canceled (not sql.ErrConnDone), so the Conn
// was returned to the pool still looking healthy. The next caller drew
// that poisoned Conn and saw sql.ErrConnDone -> driver.ErrBadConn surface
// as a 500 — and because the emulator uses dedicated db.Conn() handles,
// database/sql performs no bad-conn retry, so the error reached the
// client.
//
// The fix decouples the inner transaction from the request context, so a
// cancellation can never close the inner conn out from under us. This
// test drives a burst of concurrently-cancelled transactions and then
// asserts that every subsequent probe on a freshly drawn connection
// succeeds — no ErrBadConn, no ErrConnDone.
func TestCancelledTxDoesNotPoisonPooledConn(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?cache=shared&_test=cancel_poison")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE t (id INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	// A query heavy enough that a few-millisecond deadline reliably
	// cancels it mid-statement.
	const slow = "SELECT COUNT(*) FROM (SELECT x FROM UNNEST(GENERATE_ARRAY(1,4000000)) x)"

	var poisoned int64
	for round := 0; round < 8; round++ {
		// Burst of concurrent BeginTx+slow-query, each cancelled in flight,
		// mirroring the emulator's db.Conn(ctx)+BeginTx(ctx) request shape.
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
				defer cancel()
				conn, err := db.Conn(ctx)
				if err != nil {
					return
				}
				defer conn.Close()
				tx, err := conn.BeginTx(ctx, nil)
				if err != nil {
					return
				}
				_, _ = tx.QueryContext(ctx, slow)
				_ = tx.Rollback()
			}()
		}
		wg.Wait()

		// Probe exactly as the emulator's next request would: a dedicated
		// connection (no bad-conn retry) running a trivial query.
		for p := 0; p < 25; p++ {
			conn, err := db.Conn(context.Background())
			if err != nil {
				if isPoison(err) {
					atomic.AddInt64(&poisoned, 1)
					t.Errorf("round %d probe %d: db.Conn: %v", round, p, err)
				}
				continue
			}
			var n int
			err = conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM t").Scan(&n)
			if err != nil && isPoison(err) {
				atomic.AddInt64(&poisoned, 1)
				t.Errorf("round %d probe %d: %v", round, p, err)
			}
			conn.Close()
		}
	}
	if poisoned > 0 {
		t.Fatalf("%d probe(s) hit a poisoned connection after concurrent cancellations; the pool must never serve a conn whose inner handle was closed by a cancelled tx", poisoned)
	}
}

func TestBeginTxHonorsContextWhileWaitingForLock(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "begin_tx_cancel.db")
	db, err := sql.Open("googlesqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(2)

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS begin_tx_cancel (id INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	holder, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		t.Fatalf("holder BeginTx: %v", err)
	}
	defer func() { _ = holder.Rollback() }()
	if _, err := holder.Exec("INSERT INTO begin_tx_cancel (id) VALUES (1)"); err != nil {
		t.Fatalf("holder INSERT: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	elapsed := time.Since(start)
	if err == nil {
		_ = tx.Rollback()
		t.Fatalf("contended BeginTx succeeded; want context deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("contended BeginTx error = %v; want context deadline", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("contended BeginTx took %s; want it to return on context deadline, not busy_timeout", elapsed)
	}
}

func isPoison(err error) bool {
	return errors.Is(err, driver.ErrBadConn) ||
		errors.Is(err, sql.ErrConnDone) ||
		err.Error() == "driver: bad connection" ||
		err.Error() == "sql: connection is already closed"
}
