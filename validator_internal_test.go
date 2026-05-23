package googlesqlite

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	internal "github.com/goccy/googlesqlite/internal"
)

// TestConn_IsValid_And_TranslateConnDone covers the two pieces
// database/sql's pool relies on to drop a Conn whose inner *sql.Conn
// was closed under us (typically by a cancelled transaction's
// discardConn rollback): IsValid (cheap, called on every pool fetch)
// and the translateConnDone helper that maps sql.ErrConnDone to
// driver.ErrBadConn and flips the dead flag.
//
// Driving translateConnDone directly with the sentinel — rather than
// trying to time a real cancellation — keeps this test deterministic;
// the end-to-end behaviour is exercised by bigquery-emulator's
// TestIssue478ConnectionCorruptionOnCancel, which depends on the same
// translation working through database/sql's two-layer pool.
func TestConn_IsValid_And_TranslateConnDone(t *testing.T) {
	t.Run("fresh Conn is valid", func(t *testing.T) {
		c := &Conn{}
		if !c.IsValid() {
			t.Fatal("a freshly-allocated Conn must report IsValid()=true; otherwise database/sql would never put us in the pool")
		}
	})

	t.Run("nil error passes through and does not flip dead", func(t *testing.T) {
		c := &Conn{}
		if got := c.translateConnDone(nil); got != nil {
			t.Fatalf("translateConnDone(nil) = %v; want nil", got)
		}
		if !c.IsValid() {
			t.Fatal("Conn was marked dead after a nil error; the flag is only meant for the ErrConnDone path")
		}
	})

	t.Run("unrelated error passes through and does not flip dead", func(t *testing.T) {
		c := &Conn{}
		sentinel := errors.New("boom")
		if got := c.translateConnDone(sentinel); !errors.Is(got, sentinel) {
			t.Fatalf("translateConnDone of an unrelated error mutated it: got %v", got)
		}
		if !c.IsValid() {
			t.Fatal("Conn was marked dead after an unrelated error; only ErrConnDone may flip the flag")
		}
	})

	t.Run("ErrConnDone is mapped to ErrBadConn and flips dead", func(t *testing.T) {
		c := &Conn{}
		out := c.translateConnDone(sql.ErrConnDone)
		if !errors.Is(out, driver.ErrBadConn) {
			t.Fatalf("translateConnDone(sql.ErrConnDone) = %v; want driver.ErrBadConn so database/sql discards-and-retries", out)
		}
		if c.IsValid() {
			t.Fatal("Conn must be marked dead after observing ErrConnDone, so the next pool fetch is skipped via IsValid")
		}
	})

	t.Run("wrapped ErrConnDone is mapped to ErrBadConn", func(t *testing.T) {
		c := &Conn{}
		wrapped := fmt.Errorf("upper-layer context: %w", sql.ErrConnDone)
		out := c.translateConnDone(wrapped)
		if !errors.Is(out, driver.ErrBadConn) {
			t.Fatalf("translateConnDone(wrapped) = %v; want driver.ErrBadConn (errors.Is must traverse Unwrap)", out)
		}
		if c.IsValid() {
			t.Fatal("Conn must be marked dead even when ErrConnDone arrives wrapped")
		}
	})

	t.Run("ErrorGroup carrying ErrConnDone is mapped to ErrBadConn", func(t *testing.T) {
		// ExecContext / QueryContext fold cleanup-time errors into an
		// internal.ErrorGroup; the group's Unwrap() []error is what
		// lets errors.Is reach the buried sentinel. Without that
		// hookup the dead-conn signal would be silently lost when any
		// cleanup error landed on the same path.
		c := &Conn{}
		eg := &internal.ErrorGroup{}
		eg.Add(sql.ErrConnDone)
		eg.Add(errors.New("cleanup also failed"))
		out := c.translateConnDone(eg)
		if !errors.Is(out, driver.ErrBadConn) {
			t.Fatalf("translateConnDone(ErrorGroup) = %v; want driver.ErrBadConn (Unwrap() []error must surface the sentinel)", out)
		}
		if c.IsValid() {
			t.Fatal("Conn must be marked dead when an ErrorGroup carries ErrConnDone")
		}
	})

	t.Run("dead flag is sticky across translate calls", func(t *testing.T) {
		c := &Conn{}
		_ = c.translateConnDone(sql.ErrConnDone)
		if c.IsValid() {
			t.Fatal("setup: expected dead after first ErrConnDone")
		}
		// A subsequent benign error must not un-deaden us: once a Conn
		// has been seen poisoned the flag stays set so the pool never
		// hands it out again.
		_ = c.translateConnDone(nil)
		if c.IsValid() {
			t.Fatal("Conn un-deadened after a nil translate; the flag must be sticky")
		}
		_ = c.translateConnDone(errors.New("unrelated"))
		if c.IsValid() {
			t.Fatal("Conn un-deadened after an unrelated translate; the flag must be sticky")
		}
	})
}
