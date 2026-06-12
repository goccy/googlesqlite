package internal

import (
	"fmt"
	"os"
	"testing"

	googlesql "github.com/goccy/go-googlesql"
)

// TestMain initialises the go-googlesql wasm runtime exactly once for
// every test in package internal. The TypeFactory / Analyzer / Catalog
// surfaces all talk through the wasm module; calls before Init panic
// with a nil-pointer dereference.
//
// The runtime is the wasm2go-transpiled module — already AOT-compiled
// into Go at generation time — so there is no runtime compilation
// mode to pick and no on-disk cache to configure. The wazero-era
// GOOGLESQLITE_WASM_COMPILATION_MODE / GOOGLESQLITE_WASM_CACHE_DIR
// knobs no longer apply.
func TestMain(m *testing.M) {
	if err := googlesql.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "wasm init failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
