package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	googlesql "github.com/goccy/go-googlesql"
)

// TestMain initialises the go-googlesql wasm runtime exactly once for
// every test in package internal. The TypeFactory / Analyzer / Catalog
// surfaces all talk through the wasm module; calls before Init panic
// with a nil-pointer dereference.
//
// The runtime is AOT-compiled (Compiler mode) with an on-disk
// compilation cache so the suite does not recompile the wasm on every
// run. GOOGLESQLITE_WASM_COMPILATION_MODE = "interpreter" forces the
// interpreter; GOOGLESQLITE_WASM_CACHE_DIR overrides the cache
// directory (CI points it at a persisted path).
func TestMain(m *testing.M) {
	var opts []googlesql.Option
	if strings.EqualFold(os.Getenv("GOOGLESQLITE_WASM_COMPILATION_MODE"), "interpreter") {
		opts = append(opts, googlesql.WithCompilationMode(googlesql.CompilationModeInterpreter))
	} else {
		cacheDir := os.Getenv("GOOGLESQLITE_WASM_CACHE_DIR")
		if cacheDir == "" {
			cacheDir = filepath.Join(os.TempDir(), "googlesqlite-wasm-cache")
		}
		opts = append(opts,
			googlesql.WithCompilationMode(googlesql.CompilationModeCompiler),
			googlesql.WithCompilationCache(cacheDir),
		)
	}
	if err := googlesql.Init(opts...); err != nil {
		fmt.Fprintf(os.Stderr, "wasm init failed: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	googlesql.Close()
	os.Exit(code)
}
