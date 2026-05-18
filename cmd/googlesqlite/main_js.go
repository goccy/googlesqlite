//go:build js && wasm

// Command googlesqlite, built for js/wasm, is the engine behind the
// googlesqlite Playground. It exposes a small JavaScript API on
// globalThis.googlesqlite so a web page can run GoogleSQL queries
// entirely in the browser.
//
// Every exported function returns a Promise and performs its work on a
// goroutine. This is required, not cosmetic: the engine performs
// asynchronous host I/O (for example time.LoadLocation reads the
// timezone database), and a synchronous js.Func that blocked on such
// I/O would deadlock the Go scheduler.
//
// The database is a single file on the Origin Private File System,
// reached through the "opfs" VFS (see opfsvfs_js.go). SQLite reads and
// writes it directly, so it persists across reloads with no explicit
// save step — the page does not export or import a database image.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/goccy/googlesqlite/internal/cli"
)

const (
	// playgroundDSN opens the Playground database on the OPFS-backed
	// VFS. journal_mode=memory and temp_store=memory keep SQLite from
	// opening a rollback journal or temporary files, so the VFS only
	// ever serves the single persistent database file (opfsvfs_js.go).
	playgroundDSN = "file:/googlesqlite.db?vfs=" + opfsVFSName +
		"&_pragma=journal_mode(memory)&_pragma=temp_store(memory)"

	// fallbackDSN backs the database with the in-process "memdb" VFS
	// when OPFS is unavailable — a non-secure context, e.g. plain HTTP
	// over a LAN IP, where navigator.storage is absent. The engine
	// still runs; the database just does not persist across reloads.
	fallbackDSN = "file:/googlesqlite.db?vfs=memdb"
)

// app holds the stateful Playground session. Every exported operation
// runs under mu, so concurrent JS calls are serialised.
type app struct {
	ctx     context.Context
	mu      sync.Mutex
	session *cli.Session
	history *cli.History
	sink    *strings.Builder
}

func main() {
	registerOPFSVFS()

	a := &app{
		ctx:     context.Background(),
		history: &cli.History{},
		sink:    &strings.Builder{},
	}
	// worker.ts publishes the OPFS database handle before starting this
	// program; its absence means OPFS is unavailable, so fall back to
	// the non-persistent in-memory VFS.
	dsn := fallbackDSN
	if js.Global().Get(dbHandleGlobal).Truthy() {
		dsn = playgroundDSN
	}
	runner, err := cli.NewRunner(a.ctx, dsn)
	if err != nil {
		consoleError("googlesqlite: init failed: " + err.Error())
		return
	}
	a.session = cli.NewSession(runner, a.sink)

	js.Global().Set("googlesqlite", js.ValueOf(map[string]any{
		"exec":          js.FuncOf(a.exec),
		"execScript":    js.FuncOf(a.exec),
		"getHistory":    js.FuncOf(a.getHistory),
		"loadHistory":   js.FuncOf(a.loadHistory),
		"clearHistory":  js.FuncOf(a.clearHistory),
		"exportHistory": js.FuncOf(a.exportHistory),
		"listTables":    js.FuncOf(a.listTables),
		"setDebug":      js.FuncOf(a.setDebug),
	}))

	// Let the host page know the API is installed.
	if ready := js.Global().Get("onGoogleSQLiteReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}

	// Keep the Go runtime alive so the exported callbacks remain valid.
	select {}
}

// promise wraps a unit of work as a JavaScript Promise. The work runs
// on a fresh goroutine, so it may freely block on asynchronous host
// I/O without stalling the JS event loop.
func (a *app) promise(work func() (any, error)) any {
	var handler js.Func
	handler = js.FuncOf(func(_ js.Value, pa []js.Value) any {
		resolve, reject := pa[0], pa[1]
		go func() {
			defer handler.Release()
			defer func() {
				if r := recover(); r != nil {
					reject.Invoke(jsError(fmt.Sprintf("panic: %v", r)))
				}
			}()
			res, err := work()
			if err != nil {
				reject.Invoke(jsError(err.Error()))
				return
			}
			resolve.Invoke(res)
		}()
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}

// exec runs the SQL passed as the first argument. The text may contain
// several statements and dot-commands; each result is appended to the
// history. The resolved value is { output, results }.
func (a *app) exec(_ js.Value, args []js.Value) any {
	sql := ""
	if len(args) > 0 {
		sql = args[0].String()
	}
	return a.promise(func() (any, error) {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.sink.Reset()
		results := a.session.RunInput(a.ctx, sql, false, nil)
		jsResults := make([]any, len(results))
		for i, res := range results {
			a.history.Add(res)
			jsResults[i] = entryToJS(cli.NewHistoryEntry(res))
		}
		return map[string]any{
			"output":  a.sink.String(),
			"results": jsResults,
		}, nil
	})
}

// getHistory resolves with the query history as an array of entry
// objects.
func (a *app) getHistory(_ js.Value, _ []js.Value) any {
	return a.promise(func() (any, error) {
		a.mu.Lock()
		defer a.mu.Unlock()
		entries := make([]any, len(a.history.Entries))
		for i, e := range a.history.Entries {
			entries[i] = entryToJS(e)
		}
		return entries, nil
	})
}

// loadHistory restores a history previously produced by
// exportHistory("json").
func (a *app) loadHistory(_ js.Value, args []js.Value) any {
	jsonText := ""
	if len(args) > 0 {
		jsonText = args[0].String()
	}
	return a.promise(func() (any, error) {
		var h cli.History
		if err := json.Unmarshal([]byte(jsonText), &h); err != nil {
			return nil, err
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		a.history = &h
		return map[string]any{"ok": true}, nil
	})
}

// clearHistory drops every history entry.
func (a *app) clearHistory(_ js.Value, _ []js.Value) any {
	return a.promise(func() (any, error) {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.history.Clear()
		return map[string]any{"ok": true}, nil
	})
}

// exportHistory renders the history in the format named by the first
// argument: "json", "sql", "markdown" or "csv" (default "json").
func (a *app) exportHistory(_ js.Value, args []js.Value) any {
	format := "json"
	if len(args) > 0 {
		format = args[0].String()
	}
	return a.promise(func() (any, error) {
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.history.Export(format)
	})
}

// listTables resolves with the names of the tables and views that
// currently exist in the session.
func (a *app) listTables(_ js.Value, _ []js.Value) any {
	return a.promise(func() (any, error) {
		a.mu.Lock()
		defer a.mu.Unlock()
		names, err := a.session.TableNames(a.ctx)
		if err != nil {
			return nil, err
		}
		out := make([]any, len(names))
		for i, name := range names {
			out[i] = name
		}
		return out, nil
	})
}

// setDebug toggles whether rendered output includes the translated
// SQLite query.
func (a *app) setDebug(_ js.Value, args []js.Value) any {
	on := len(args) > 0 && args[0].Truthy()
	return a.promise(func() (any, error) {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.session.Debug = on
		return map[string]any{"debug": a.session.Debug}, nil
	})
}

// entryToJS converts a history entry into a JS-marshalable map.
func entryToJS(e cli.HistoryEntry) map[string]any {
	rows := make([]any, len(e.Rows))
	for i, row := range e.Rows {
		rows[i] = stringsToJS(row)
	}
	return map[string]any{
		"statement":    e.Statement,
		"sqliteQuery":  e.SQLiteQuery,
		"isQuery":      e.IsQuery,
		"columns":      stringsToJS(e.Columns),
		"rows":         rows,
		"rowsAffected": float64(e.RowsAffected),
		"error":        e.Error,
		"elapsedMs":    e.ElapsedMs,
		"timestamp":    e.Timestamp.Format(time.RFC3339Nano),
	}
}

func stringsToJS(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

// jsError builds a JavaScript Error object so a rejected Promise
// carries a real Error to the caller.
func jsError(msg string) js.Value {
	return js.Global().Get("Error").New(msg)
}

func consoleError(msg string) {
	if c := js.Global().Get("console"); c.Type() == js.TypeObject {
		c.Call("error", msg)
	}
}
