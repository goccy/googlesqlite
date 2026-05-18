// Package cli is the shared core of the googlesqlite command-line
// tool. It executes GoogleSQL statements through the database/sql
// driver and renders results as display-width-aware tables.
//
// The native REPL (cmd/googlesqlite, native build) and the wasm
// Playground entrypoint (cmd/googlesqlite, js/wasm build) both build
// on this package, so it deliberately avoids terminal- and
// syscall/js-specific dependencies: nothing here imports readline or
// syscall/js.
package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	_ "github.com/goccy/googlesqlite" // register the googlesqlite driver
	"github.com/goccy/googlesqlite/internal"
	_ "github.com/ncruces/go-sqlite3/vfs/memdb" // register the memdb VFS
)

const (
	// driverName is the public googlesqlite database/sql driver.
	driverName = "googlesqlite"
	// rawDriverName is the raw SQLite engine behind the driver. The CLI
	// uses it only to read the googlesqlite_catalog bookkeeping table
	// for the .tables / .functions dot-commands.
	rawDriverName = "googlesqlite_sqlite3"

	// DefaultMemoryDSN is the CLI's default data source: an in-memory
	// database on the ncruces "memdb" VFS. Unlike a bare :memory: DSN,
	// a memdb database is shared between connections in the process, so
	// the short-lived connection that .tables / .functions open to read
	// the catalog sees the same data as the session connection.
	DefaultMemoryDSN = "file:/googlesqlite_cli.db?vfs=memdb"
)

// Result is the structured outcome of executing a single statement.
type Result struct {
	// Statement is the original GoogleSQL text, with any trailing \G
	// marker removed.
	Statement string
	// GroupMode is true when the statement carried a trailing \G.
	GroupMode bool
	// IsQuery is true when the statement produced a result set.
	IsQuery bool
	// Columns holds the result column names (queries only).
	Columns []string
	// ColumnKinds holds the GoogleSQL type kind of each column, used
	// to colourise values.
	ColumnKinds []googlesql.TypeKind
	// Rows holds the scanned result rows (queries only).
	Rows [][]any
	// RowsAffected is the affected-row count (non-queries only).
	RowsAffected int64
	// SQLiteQuery is the translated SQLite text the engine executed.
	// It is always captured; the CLI shows it only in debug mode.
	SQLiteQuery string
	// Elapsed is the wall-clock execution time.
	Elapsed time.Duration
	// Err is the execution error, if any.
	Err error
}

// Runner owns one googlesqlite session: a single pinned connection so
// schema and data created by one statement are visible to the next.
type Runner struct {
	dsn  string
	db   *sql.DB
	conn *sql.Conn
}

// NewRunner opens a googlesqlite database at dsn and pins one
// connection for the whole session. A single connection is required
// because a fresh connection from the pool could land on a different
// in-memory database.
func NewRunner(ctx context.Context, dsn string) (*Runner, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	// One connection only: every statement must see the same session
	// state, and the pinned conn below keeps an in-memory DSN alive.
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Runner{dsn: dsn, db: db, conn: conn}, nil
}

// DSN returns the data source name the runner was opened with.
func (r *Runner) DSN() string { return r.dsn }

// Close releases the pinned connection and the underlying pool.
func (r *Runner) Close() error {
	if r.conn != nil {
		_ = r.conn.Close()
	}
	return r.db.Close()
}

// Exec runs a single statement and returns its structured Result. A
// trailing \G selects group display mode. The translated SQLite query
// is captured into Result.SQLiteQuery via the internal SQL collector.
func (r *Runner) Exec(ctx context.Context, statement string) Result {
	res := Result{Statement: statement}
	stmt := strings.TrimSpace(statement)
	if trimmed, ok := strings.CutSuffix(stmt, `\G`); ok {
		res.GroupMode = true
		stmt = strings.TrimSpace(trimmed)
	}
	res.Statement = stmt
	if stmt == "" {
		return res
	}

	ctx, collector := internal.NewSQLCollectorContext(ctx)
	start := time.Now()
	if isQueryStatement(stmt) {
		res.IsQuery = true
		res.Err = r.runQuery(ctx, stmt, &res)
	} else {
		res.Err = r.runExec(ctx, stmt, &res)
	}
	res.Elapsed = time.Since(start)
	if queries := collector.Queries(); len(queries) > 0 {
		res.SQLiteQuery = strings.Join(queries, ";\n")
	}
	return res
}

func (r *Runner) runQuery(ctx context.Context, stmt string, res *Result) error {
	rows, err := r.conn.QueryContext(ctx, stmt)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	res.Columns = cols

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	res.ColumnKinds = make([]googlesql.TypeKind, len(colTypes))
	for i, ct := range colTypes {
		res.ColumnKinds[i] = parseTypeKind(ct.DatabaseTypeName())
	}

	for rows.Next() {
		cells := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range cells {
			ptrs[i] = &cells[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		res.Rows = append(res.Rows, cells)
	}
	return rows.Err()
}

func (r *Runner) runExec(ctx context.Context, stmt string, res *Result) error {
	result, err := r.conn.ExecContext(ctx, stmt)
	if err != nil {
		return err
	}
	if n, err := result.RowsAffected(); err == nil {
		res.RowsAffected = n
	}
	return nil
}

// parseTypeKind decodes the JSON type descriptor that the driver
// returns from ColumnType.DatabaseTypeName into a GoogleSQL type kind.
// An undecodable descriptor yields TypeKindTypeUnknown.
func parseTypeKind(dbTypeName string) googlesql.TypeKind {
	var t internal.Type
	if err := json.Unmarshal([]byte(dbTypeName), &t); err != nil {
		return googlesql.TypeKindTypeUnknown
	}
	return googlesql.TypeKind(t.Kind)
}
