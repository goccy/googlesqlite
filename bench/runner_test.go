// Package bench measures the cost of executing a corpus of GoogleSQL
// queries against the googlesqlite driver (ncruces + pure-Go SQLite).
// Run via
//
//	cd bench && go test -bench=. -benchmem -run=^$
//
// The purpose of this suite is to detect performance regressions in
// googlesqlite across revisions: capture a baseline, change the
// driver, and compare the new run against the baseline.
package bench

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goccy/googlesqlite"
)

// driverName is the database/sql driver name under benchmark.
const driverName = "googlesqlite"

// TestMain runs the benchmark in wazero Compiler mode so the numbers
// reflect steady-state performance rather than the interpreter.
//
// The embedded GoogleSQL wasm is AOT-compiled once, lazily, on the
// first connection. To keep that one-time compilation cost out of
// whichever corpus query runs first, TestMain warms the runtime
// before m.Run and measures the compile cost on its own — the figure
// is emitted as a `wasm-compile-ns` line that render_results.go
// reports separately. The corpus benchmarks below therefore all run
// against an already-compiled runtime.
func TestMain(m *testing.M) {
	os.Exit(runBenchmarks(m))
}

func runBenchmarks(m *testing.M) int {
	if os.Getenv(googlesqlite.EnvWasmCompilationMode) == "" {
		// A fresh cache directory per run makes the warm-up below a
		// genuine cold AOT compile rather than a precompiled-module
		// load, so the reported compile cost is consistent.
		cacheDir, err := os.MkdirTemp("", "googlesqlite-bench-wasm-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "bench: temp cache dir: %v\n", err)
			return 1
		}
		defer os.RemoveAll(cacheDir)
		os.Setenv(googlesqlite.EnvWasmCompilationMode, string(googlesqlite.WasmCompilationModeCompiler))
		os.Setenv(googlesqlite.EnvWasmCacheDir, cacheDir)
	}
	fmt.Printf("wasm-compile-ns %d\n", warmWasmRuntime())
	return m.Run()
}

// warmWasmRuntime forces the one-time wasm AOT compilation by opening a
// throwaway connection and running a trivial query, and returns how
// long that took in nanoseconds.
func warmWasmRuntime() int64 {
	start := time.Now()
	db, err := sql.Open(driverName, "file:bench_warmup?mode=memory&cache=shared")
	if err != nil {
		panic(fmt.Sprintf("bench: warm-up open: %v", err))
	}
	defer db.Close()
	var n int
	if err := db.QueryRow("SELECT 1").Scan(&n); err != nil {
		panic(fmt.Sprintf("bench: warm-up query: %v", err))
	}
	return time.Since(start).Nanoseconds()
}

type queryCase struct {
	category string // "window", "aggregate", "scalar", "mixed", "tpch_like"
	name     string
	setup    string // optional CREATE / INSERT block, separated from the timed query
	query    string // the timed SELECT
	prep     bool   // when true, the query is Prepare'd once and re-executed each iteration
}

// loadCorpus reads bench/corpus/<category>/<name>.sql files. Each file
// has two sections separated by `-- @query`:
//
//	<setup statements ...>
//	-- @query
//	<the timed query>
//
// If `-- @query` is absent the whole file is treated as the query
// (no setup).
func loadCorpus(t testing.TB) []queryCase {
	t.Helper()
	root := "corpus"
	var cases []queryCase
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		parts := strings.SplitN(string(data), "\n-- @query\n", 2)
		var setup, query string
		if len(parts) == 2 {
			setup = strings.TrimSpace(parts[0])
			query = strings.TrimSpace(parts[1])
		} else {
			query = strings.TrimSpace(parts[0])
		}
		category, name := splitCorpusPath(rel)
		cases = append(cases, queryCase{
			category: category,
			name:     name,
			setup:    setup,
			query:    query,
			prep:     true,
		})
		return nil
	})
	if err != nil {
		t.Fatalf("loadCorpus: %v", err)
	}
	sort.Slice(cases, func(i, j int) bool {
		if cases[i].category != cases[j].category {
			return cases[i].category < cases[j].category
		}
		return cases[i].name < cases[j].name
	})
	return cases
}

func splitCorpusPath(rel string) (category, name string) {
	parts := strings.SplitN(filepath.ToSlash(rel), "/", 2)
	if len(parts) != 2 {
		return "misc", strings.TrimSuffix(rel, ".sql")
	}
	return parts[0], strings.TrimSuffix(parts[1], ".sql")
}

// openWithSetup opens a fresh in-memory DB for the given driver name
// and runs setup. Returns the *sql.DB and an optional prepared *sql.Stmt
// when prep is true. The caller is responsible for Close().
func openWithSetup(b *testing.B, driverName string, c queryCase) (*sql.DB, *sql.Stmt) {
	b.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniqueDSN())
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		b.Fatalf("open(%s): %v", driverName, err)
	}
	// CREATE TEMP TABLE is per-connection in SQLite; pin the pool to a
	// single connection so setup statements and the timed query share
	// the same in-process catalog.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		b.Fatalf("ping(%s): %v", driverName, err)
	}
	if c.setup != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		if _, err := db.ExecContext(ctx, c.setup); err != nil {
			_ = db.Close()
			b.Fatalf("setup(%s): %v\nsetup:\n%s", driverName, err, c.setup)
		}
	}
	if c.prep {
		stmt, err := db.Prepare(c.query)
		if err != nil {
			_ = db.Close()
			b.Fatalf("prepare(%s): %v\nquery:\n%s", driverName, err, c.query)
		}
		return db, stmt
	}
	return db, nil
}

// uniqueDSN returns a unique in-memory DSN per call. Without this,
// shared-cache memory databases would collide between concurrent
// benchmarks and corrupt each other's catalogs.
var dsnSeq atomic_uint64

func uniqueDSN() string {
	return fmt.Sprintf("googlesqlite_bench_%d", dsnSeq.add())
}

// atomic_uint64 keeps the runner self-contained without pulling in
// sync/atomic generics every place; the seq is monotonic and
// thread-safe enough for benchmarks.
type atomic_uint64 struct {
	mu sync.Mutex
	v  uint64
}

func (a *atomic_uint64) add() uint64 {
	a.mu.Lock()
	a.v++
	v := a.v
	a.mu.Unlock()
	return v
}

func runQuery(b *testing.B, db *sql.DB, stmt *sql.Stmt, query string) {
	b.Helper()
	var rows *sql.Rows
	var err error
	if stmt != nil {
		rows, err = stmt.Query()
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		b.Fatalf("query: %v", err)
	}
	cols, err := rows.Columns()
	if err != nil {
		_ = rows.Close()
		b.Fatalf("columns: %v", err)
	}
	scanArgs := make([]any, len(cols))
	scanPtrs := make([]any, len(cols))
	for i := range scanArgs {
		scanPtrs[i] = &scanArgs[i]
	}
	for rows.Next() {
		if err := rows.Scan(scanPtrs...); err != nil {
			_ = rows.Close()
			b.Fatalf("scan: %v", err)
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		b.Fatalf("rows.Err: %v", err)
	}
	_ = rows.Close()
}

// BenchmarkCorpus — top-level entry point. It spawns one
// sub-benchmark per corpus query, all run against googlesqlite.
//
// Run with `go test -bench=. -benchmem` from the bench/ directory.

func BenchmarkCorpus(b *testing.B) {
	cases := loadCorpus(b)
	for _, c := range cases {
		c := c
		b.Run(filepath.Join(c.category, c.name), func(b *testing.B) {
			db, stmt := openWithSetup(b, driverName, c)
			defer func() {
				if stmt != nil {
					_ = stmt.Close()
				}
				_ = db.Close()
			}()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runQuery(b, db, stmt, c.query)
			}
		})
	}
}
