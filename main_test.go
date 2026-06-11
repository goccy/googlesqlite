package googlesqlite_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/goccy/googlesqlite"
)

// nopWriteCloser discards everything written. The spec-corpus EXPORT DATA
// case ('mem://specsuite/out.csv') only asserts that the statement runs
// and yields no rows; the actual bytes do not need to be inspected here
// (TestExportDataMemRoundTrip and the unit tests that build their own
// capture map exercise the byte path). Registering this once in TestMain
// avoids the spec runner having to wire URI writers per case.
type nopWriteCloser struct{}

func (nopWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (nopWriteCloser) Close() error                { return nil }

// TestMain pins the process timezone to UTC for the whole suite.
// Timestamp tests need a deterministic zone; fixing it once here
// (instead of per-test t.Setenv/os.Setenv) keeps those tests free of
// process-global mutation so they can run with t.Parallel().
//
// The embedded GoogleSQL wasm runtime is the wasm2go-transpiled
// module — AOT-compiled into Go at generation time — so there is no
// runtime compilation mode to pick and no on-disk cache to configure.
func TestMain(m *testing.M) {
	os.Setenv("TZ", "UTC")
	time.Local = time.UTC

	// Wire a discard `mem://` writer so the spec-corpus EXPORT DATA case
	// has a destination to write to. Unit tests that need to inspect the
	// captured bytes register their own scheme via registerMemScheme(t)
	// (see export_data_test.go) — that helper picks a per-test scheme so
	// concurrent tests do not contend on a shared map.
	googlesqlite.RegisterExportURIWriter("mem", func(_ context.Context, _ string, _ googlesqlite.ExportWriterOpts) (io.WriteCloser, error) {
		return nopWriteCloser{}, nil
	})

	os.Exit(m.Run())
}
