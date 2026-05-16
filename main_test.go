package googlesqlite_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goccy/googlesqlite"
)

// TestMain selects the wazero Compiler mode with a stable on-disk
// compilation cache for the black-box suite, so it does not
// re-AOT-compile the embedded wasm on every run. The driver reads the
// EnvWasm* environment variables on the first sql.Open; setting them
// here (when unset) keeps a bare `go test` fast without exposing a
// programmatic configuration API. CI sets these variables already, in
// which case the existing values are kept.
func TestMain(m *testing.M) {
	// Pin the process timezone to UTC for the whole suite. Timestamp
	// tests need a deterministic zone; fixing it once here (instead of
	// per-test t.Setenv/os.Setenv) keeps those tests free of process-
	// global mutation so they can run with t.Parallel().
	os.Setenv("TZ", "UTC")
	time.Local = time.UTC

	if os.Getenv(googlesqlite.EnvWasmCompilationMode) == "" {
		os.Setenv(googlesqlite.EnvWasmCompilationMode, string(googlesqlite.WasmCompilationModeCompiler))
		if os.Getenv(googlesqlite.EnvWasmCacheDir) == "" {
			os.Setenv(googlesqlite.EnvWasmCacheDir, filepath.Join(os.TempDir(), "googlesqlite-wasm-cache"))
		}
	}
	os.Exit(m.Run())
}
