package googlesqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"

	googlesql "github.com/goccy/go-googlesql"
	sqlite3 "github.com/ncruces/go-sqlite3"
	sqlitedriver "github.com/ncruces/go-sqlite3/driver"

	internal "github.com/goccy/googlesqlite/internal"
	"github.com/goccy/googlesqlite/internal/sqlitex"
)

// dsnParts splits a googlesqlite DSN into the filename portion the SQLite
// engine sees and the query string we parse for our own options. SQLite URIs
// (those that start with "file:") carry their own query parameters such as
// `mode=memory` and `cache=shared` that SQLite itself interprets — those must
// pass through to the engine untouched, otherwise `file:foo?mode=memory`
// silently becomes an on-disk file named `file:foo` and accumulates state
// across test runs. For non-URI DSNs like `:memory:?dialect=bigquery`, we
// strip the first `?` since SQLite would otherwise treat the whole DSN as a
// filename.
func dsnParts(dsn string) (path, query string) {
	if strings.HasPrefix(dsn, "file:") {
		return dsn, ""
	}
	if i := strings.IndexByte(dsn, '?'); i >= 0 {
		return dsn[:i], dsn[i+1:]
	}
	return dsn, ""
}

// ensureGoogleSQLInit initialises the embedded GoogleSQL WASM runtime
// once per process. Subsequent calls are no-ops (Init uses sync.Once).
// It is invoked on the first sql.Open of a googlesqlite DB.
//
// The runtime is the wasm2go-transpiled module — already AOT-compiled
// into Go at generation time — so there is no runtime compilation
// mode to pick and no on-disk cache to configure.
var ensureGoogleSQLInit = sync.OnceValue(func() error {
	return googlesql.Init()
})

var (
	_ driver.Driver = &Driver{}
	_ driver.Conn   = &Conn{}
	_ driver.Tx     = &Tx{}
)

var (
	nameToCatalogMap = map[string]*internal.Catalog{}
	nameToDBMap      = map[string]*sql.DB{}
	nameToValueMapMu sync.Mutex
)

// internalDriverName is the database/sql driver name for the raw
// SQLite engine that backs every googlesqlite query. It registers
// every custom function and collation defined in
// internal.RegisterFunctions on each new connection.
const internalDriverName = "googlesqlite_sqlite3"

func init() {
	sql.Register("googlesqlite", &Driver{})
	sql.Register(internalDriverName, &internalSQLiteDriver{})
}

// internalSQLiteDriver wraps ncruces/go-sqlite3 and runs the
// googlesqlite function registration on every new connection.
type internalSQLiteDriver struct{}

func (internalSQLiteDriver) Open(name string) (driver.Conn, error) {
	path, _ := dsnParts(name)
	conn, err := (&sqlitedriver.SQLite{}).Open(path)
	if err != nil {
		return nil, err
	}
	rawConn, ok := conn.(sqlitedriver.Conn)
	if !ok {
		_ = conn.Close()
		return nil, fmt.Errorf("googlesqlite: unexpected ncruces conn type %T", conn)
	}
	raw := rawConn.Raw()
	// The predecessor's value layer emits double-quoted base64 string
	// literals into rewritten SQL. Modern SQLite (and the ncruces
	// build) treat double-quoted tokens as identifiers by default; we
	// need DQS enabled so the rewritten SQL parses as the predecessor
	// expected.
	if _, err := raw.Config(sqlite3.DBCONFIG_DQS_DDL, true); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("googlesqlite: enable DQS_DDL: %w", err)
	}
	if _, err := raw.Config(sqlite3.DBCONFIG_DQS_DML, true); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("googlesqlite: enable DQS_DML: %w", err)
	}
	// Allow user-registered functions to be referenced from inside
	// views and triggers. Without this, SQLite's safety check refuses
	// to evaluate googlesqlite_*-flavoured calls embedded in view
	// definitions.
	if _, err := raw.Config(sqlite3.DBCONFIG_TRUSTED_SCHEMA, true); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("googlesqlite: enable TRUSTED_SCHEMA: %w", err)
	}
	if err := internal.RegisterFunctions(raw); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("googlesqlite: register functions: %w", err)
	}
	// INFORMATION_SCHEMA virtual tables are bound to the per-DSN
	// catalog. The googlesqlite top-level Driver.Open populates
	// nameToCatalogMap[name] before triggering db.Conn(), which
	// is what eventually drives this Open. The catalog is therefore
	// always present at this point (sequential, single-DSN flow).
	nameToValueMapMu.Lock()
	catalog := nameToCatalogMap[name]
	nameToValueMapMu.Unlock()
	if catalog != nil {
		if err := internal.RegisterInfoSchemaModules(raw, catalog); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("googlesqlite: register info-schema modules: %w", err)
		}
	}
	sqlitex.SetVariableNumberLimit(raw, -1)
	return conn, nil
}

func newDBAndCatalog(name string) (*sql.DB, *internal.Catalog, error) {
	if err := ensureGoogleSQLInit(); err != nil {
		return nil, nil, fmt.Errorf("googlesqlite: googlesql init: %w", err)
	}
	nameToValueMapMu.Lock()
	defer nameToValueMapMu.Unlock()
	db, exists := nameToDBMap[name]
	if exists {
		return db, nameToCatalogMap[name], nil
	}
	db, err := sql.Open(internalDriverName, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database by %s: %w", name, err)
	}
	catalog := internal.NewCatalog(db)
	nameToDBMap[name] = db
	nameToCatalogMap[name] = catalog
	return db, catalog, nil
}

type Driver struct {
	ConnectHook func(*Conn) error
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	c, err := d.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return c.Connect(context.Background())
}

// memoryConnectorSeq mints a unique suffix per :memory: connector so each
// sql.Open(":memory:") gets its own underlying *sql.DB and catalog. Without
// this, the DSN-keyed cache would hand every :memory: open the same *sql.DB,
// and SQLite's connection pool would persist DDL across sql.Opens — observable
// as "table already exists" on a second sql.Open(":memory:") run.
var memoryConnectorSeq atomic.Uint64

// connector is the googlesqlite driver.Connector. Holding the inner *sql.DB
// and catalog on the connector means every conn from the same *sql.DB shares
// state, while different *sql.DB instances get independent catalogs. The
// previous global DSN-keyed cache could not distinguish those two cases.
type connector struct {
	driver     *Driver
	name       string
	inner      *sql.DB
	catalog    *internal.Catalog
	innerOwned bool
	innerName  string
}

func (c *connector) Connect(_ context.Context) (dc driver.Conn, e error) {
	defer recoverPanicAsError(&e)
	conn, err := newConn(c.inner, c.catalog)
	if err != nil {
		return nil, err
	}
	if c.driver.ConnectHook != nil {
		if err := c.driver.ConnectHook(conn); err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (c *connector) Driver() driver.Driver { return c.driver }

func (c *connector) Close() error {
	if !c.innerOwned {
		return nil
	}
	nameToValueMapMu.Lock()
	delete(nameToCatalogMap, c.innerName)
	nameToValueMapMu.Unlock()
	return c.inner.Close()
}

func (d *Driver) OpenConnector(name string) (driver.Connector, error) {
	if err := ensureGoogleSQLInit(); err != nil {
		return nil, fmt.Errorf("googlesqlite: googlesql init: %w", err)
	}
	if isPrivateInMemoryDSN(name) {
		innerName := freshMemoryDSN(name)
		inner, err := sql.Open(internalDriverName, innerName)
		if err != nil {
			return nil, fmt.Errorf("failed to open database by %s: %w", name, err)
		}
		catalog := internal.NewCatalog(inner)
		nameToValueMapMu.Lock()
		nameToCatalogMap[innerName] = catalog
		nameToValueMapMu.Unlock()
		return &connector{driver: d, name: name, inner: inner, catalog: catalog, innerOwned: true, innerName: innerName}, nil
	}
	db, catalog, err := newDBAndCatalog(name)
	if err != nil {
		return nil, err
	}
	return &connector{driver: d, name: name, inner: db, catalog: catalog}, nil
}

// isPrivateInMemoryDSN reports whether the DSN names a private in-memory
// database that should be re-created on every sql.Open. Bare ":memory:" is
// the obvious case; SQLite URIs of the form `file:NAME?mode=memory&...`
// without `cache=shared` are also private per-connection in-memory DBs.
// The flip side — `file::memory:?cache=shared` (used by some consumers)
// — is *intentionally* shared across opens, so we keep that in the DSN
// cache.
func isPrivateInMemoryDSN(name string) bool {
	path, _ := dsnParts(name)
	if path == ":memory:" {
		return true
	}
	if !strings.HasPrefix(name, "file:") {
		return false
	}
	i := strings.IndexByte(name, '?')
	if i < 0 {
		return false
	}
	hasModeMemory := false
	hasSharedCache := false
	for _, kv := range strings.Split(name[i+1:], "&") {
		switch kv {
		case "mode=memory":
			hasModeMemory = true
		case "cache=shared":
			hasSharedCache = true
		}
	}
	return hasModeMemory && !hasSharedCache
}

// freshMemoryDSN appends a unique query parameter so the inner sql.Open call
// produces a distinct cache key per outer sql.Open. SQLite ignores unknown
// URI parameters, so adding `_googlesqlite_id` is safe for both bare
// ":memory:" DSNs and SQLite URI-form DSNs.
func freshMemoryDSN(name string) string {
	id := memoryConnectorSeq.Add(1)
	suffix := fmt.Sprintf("_googlesqlite_id=%d", id)
	if strings.IndexByte(name, '?') >= 0 {
		return name + "&" + suffix
	}
	return name + "?" + suffix
}

var _ driver.DriverContext = (*Driver)(nil)

type Conn struct {
	conn           *sql.Conn
	tx             *sql.Tx
	analyzer       *internal.Analyzer
	catalog        *internal.Catalog
	systemVars     map[string]string
	scriptVars     map[string]string
	materializeCTE bool

	// dead is set when an operation on this Conn observes that the
	// inner *sql.Conn has been closed (sql.ErrConnDone). database/sql
	// pools Conn values across requests; without an explicit signal it
	// happily hands a Conn whose inner handle has been closed by a
	// cancelled tx's discardConn rollback to the next caller, which
	// then surfaces a confusing "sql: connection is already closed"
	// 500 on every subsequent request (see IsValid and the
	// translateConnDone defers below). The dead flag is read by IsValid
	// and written from any goroutine that drove the failing operation;
	// concurrent writes are harmless because they all store the same
	// value (true).
	dead atomic.Bool
}

// IsValid implements database/sql/driver.Validator. database/sql calls
// it on every pool fetch before handing the Conn to the next caller —
// returning false makes the pool discard us and pull a fresh one
// instead of letting us surface a confusing "sql: connection is
// already closed" error inside the next request. The flag is set by
// translateConnDone whenever an earlier operation observed the inner
// handle had been closed (typically by a request whose context was
// cancelled mid-statement; database/sql's tx watcher then ran
// Rollback(discardConn=true) on the inner connection).
func (c *Conn) IsValid() bool {
	return !c.dead.Load()
}

// translateConnDone is the single point that converts the inner
// "connection is closed" sentinel into the driver-level "drop me"
// sentinel.
//
// The inner *sql.Conn we hold can become done while still appearing
// healthy to database/sql's outer pool: a cancelled request triggers
// the inner-tx watcher's Rollback(discardConn=true), which closes the
// inner conn, but the outer pool doesn't know because the failure
// hasn't surfaced yet. The first operation on the poisoned Conn then
// fails with sql.ErrConnDone — which to a caller looks like a server
// error, not a transient retryable condition. Mapping it to
// driver.ErrBadConn tells database/sql to drop the Conn from the pool
// and retry on a fresh one (when the call has not yet sent data),
// hiding the transient from the caller; setting dead=true means the
// next IsValid call returns false so the Conn never even gets that
// far again.
//
// Wrapped errors are inspected because ExecContext / QueryContext may
// fold cleanup-time errors into an internal.ErrorGroup; the group
// implements Unwrap() []error so errors.Is reaches the sentinel.
func (c *Conn) translateConnDone(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrConnDone) {
		c.dead.Store(true)
		return driver.ErrBadConn
	}
	return err
}

// Static contract check: database/sql's pool only calls IsValid via a
// type assertion against driver.Validator. Failing this assertion at
// compile time keeps a future refactor that drops the method visible
// instead of silently regressing the #478 fix.
var _ driver.Validator = (*Conn)(nil)

func newConn(db *sql.DB, catalog *internal.Catalog) (*Conn, error) {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get sqlite3 connection: %w", err)
	}
	analyzer, err := internal.NewAnalyzer(catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}
	return &Conn{
		conn:           conn,
		analyzer:       analyzer,
		catalog:        catalog,
		systemVars:     map[string]string{},
		scriptVars:     map[string]string{},
		materializeCTE: true,
	}, nil
}

// SetMaxNamePath specifies the maximum value of name path.
// If the name path in the query is the maximum value, the name path set as prefix is not used.
// Effective only when a value greater than zero is specified ( default zero ).
func (c *Conn) SetMaxNamePath(num int) {
	c.analyzer.SetMaxNamePath(num)
}

// SetNamePath set path to name path to be set as prefix.
// If max name path is specified, an error is returned if the number is exceeded.
func (c *Conn) SetNamePath(path []string) error {
	return c.analyzer.SetNamePath(path)
}

func (s *Conn) CheckNamedValue(value *driver.NamedValue) error {
	// Encode every value into the googlesqlite value layout so the
	// rest of the pipeline (formatter, reshapeInsertArgs, scanner)
	// always sees the same shape. Primitives that the underlying
	// SQLite driver can store natively (int64, float64, bool) are
	// returned unchanged by EncodeValue.
	return internal.CheckNamedValue(value)
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.PrepareContext(context.Background(), query)
	return stmt, err
}

// recoverPanicAsError converts an arbitrary panic value into an error
// recorded on `e`. It is meant to be invoked from a deferred call in
// every driver-public entry point so that a bug anywhere inside
// analyzer/formatter/runtime never escapes as a panic — a `nil`
// `*sql.Conn`/`*sql.Stmt` being returned by a panicking driver call
// is what triggers `database/sql`'s deferred close to deadlock on the
// per-conn closemu lock.
//
// The first line of the stack trace is preserved on the returned
// error so the failure remains debuggable; downstream callers
// (database/sql) get an ordinary error and can clean up normally.
func recoverPanicAsError(e *error) {
	r := recover()
	if r == nil {
		return
	}
	stack := debug.Stack()
	if existing := *e; existing != nil {
		*e = fmt.Errorf("googlesqlite: panic %v (prior error: %w)\n%s", r, existing, stack)
	} else {
		*e = fmt.Errorf("googlesqlite: panic %v\n%s", r, stack)
	}
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (s driver.Stmt, e error) {
	defer func() { e = c.translateConnDone(e) }()
	defer recoverPanicAsError(&e)
	conn := internal.NewConnWithOptions(c.conn, c.tx, c.systemVars, c.scriptVars, c.materializeCTE)
	actionFuncs, err := c.analyzer.Analyze(ctx, conn, query, nil)
	if err != nil {
		return nil, err
	}
	var stmt driver.Stmt
	for _, actionFunc := range actionFuncs {
		action, err := actionFunc()
		if err != nil {
			return nil, err
		}
		got, err := action.Prepare(ctx, conn)
		if err != nil {
			return nil, err
		}
		stmt = got
	}
	return stmt, nil
}

func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, e error) {
	defer func() { e = c.translateConnDone(e) }()
	defer recoverPanicAsError(&e)
	conn := internal.NewConnWithOptions(c.conn, c.tx, c.systemVars, c.scriptVars, c.materializeCTE)
	actionFuncs, err := c.analyzer.Analyze(ctx, conn, query, args)
	if err != nil {
		return nil, err
	}
	var actions []internal.StmtAction
	defer func() {
		eg := new(internal.ErrorGroup)
		eg.Add(e)
		for _, action := range actions {
			eg.Add(action.Cleanup(ctx, conn))
		}
		if eg.HasError() {
			e = eg
		}
	}()

	var result driver.Result
	for _, actionFunc := range actionFuncs {
		action, err := actionFunc()
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
		r, err := action.ExecContext(ctx, conn)
		if err != nil {
			return nil, err
		}
		result = r
	}
	return result, nil
}

func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Rows, e error) {
	defer func() { e = c.translateConnDone(e) }()
	defer recoverPanicAsError(&e)
	conn := internal.NewConnWithOptions(c.conn, c.tx, c.systemVars, c.scriptVars, c.materializeCTE)
	actionFuncs, err := c.analyzer.Analyze(ctx, conn, query, args)
	if err != nil {
		return nil, err
	}
	var (
		actions []internal.StmtAction
		rows    *internal.Rows
	)
	defer func() {
		if rows != nil {
			// If we call cleanup action at the end of QueryContext function,
			// there is a possibility that the deleted table will be referenced when scanning from Rows,
			// so cleanup action should be executed in the Close() process of Rows.
			// For that, let Rows have a reference to actions ( and connection ).
			rows.SetActions(actions)
		}
	}()
	for _, actionFunc := range actionFuncs {
		action, err := actionFunc()
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
		queryRows, err := action.QueryContext(ctx, conn)
		if err != nil {
			return nil, err
		}
		rows = queryRows
	}
	return rows, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (t driver.Tx, e error) {
	defer func() { e = c.translateConnDone(e) }()
	// Honour an already-cancelled request context for the BEGIN itself…
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// …but deliberately do NOT hand the request context to the inner
	// transaction. database/sql starts a watcher goroutine for any tx
	// whose context can be cancelled, and on cancellation it runs
	// Rollback(discardConn=true), which closes the inner *sql.Conn we
	// hold. That close happens asynchronously, after our own operation
	// has already returned the caller's ctx error (context.Canceled, not
	// sql.ErrConnDone) — so translateConnDone never marks us dead and the
	// outer pool hands this now-poisoned Conn to the next caller, which
	// then fails with "connection is already closed" / driver.ErrBadConn
	// (bigquery-emulator #478). Binding the inner tx to a context the
	// request cannot cancel removes that asynchronous close entirely; the
	// inner transaction is instead torn down deterministically by our own
	// Commit/Rollback. Per-statement cancellation is unaffected because
	// ExecContext/QueryContext still receive and honour the request ctx,
	// so an in-flight query is still interrupted on cancel.
	tx, err := c.conn.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.IsolationLevel(opts.Isolation),
		ReadOnly:  opts.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	c.tx = tx
	return &Tx{
		tx:   tx,
		conn: c,
	}, nil
}

func (c *Conn) Begin() (t driver.Tx, e error) {
	defer func() { e = c.translateConnDone(e) }()
	tx, err := c.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	c.tx = tx
	return &Tx{
		tx:   tx,
		conn: c,
	}, nil
}

type Tx struct {
	tx   *sql.Tx
	conn *Conn
}

func (tx *Tx) Commit() error {
	defer func() {
		tx.conn.tx = nil
	}()
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	defer func() {
		tx.conn.tx = nil
	}()
	return tx.tx.Rollback()
}
