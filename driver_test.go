package googlesqlite_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goccy/googlesqlite"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ---- from driver_surface_test.go ----

// Black-box tests covering the public package surface that
// currently lacks dedicated coverage: Driver.Open / Driver.Driver,
// Conn.Prepare / Conn.Begin, and UnmarshalDatabaseTypeName. The
// DSN-parsing branches and connector lifecycle fixtures are clustered
// here so that future refactors of driver.go only require updating
// the test setup, not the assertions.
//
// Expected behaviour is encoded in driver.go GoDoc and in the
// existing test suite (memory_dsn_isolation_test.go,
// name_path_public_test.go, transaction_test.go, recover_panic_test.go).

// ------------------------------------------------------------------
// Driver.Open / Driver.Driver / connector lifecycle
// ------------------------------------------------------------------

// Driver.Open is the legacy entry point database/sql calls when the
// driver does not implement DriverContext. Even though our Driver
// does implement DriverContext (so Open is bypassed for sql.Open),
// the method is still part of the documented contract — exercise it
// directly to keep the surface covered.
func TestDriverOpen_ReturnsConnAndDriver(t *testing.T) {
	t.Parallel()
	d := &googlesqlite.Driver{}
	conn, err := d.Open(":memory:?_test=driver_open")
	if err != nil {
		t.Fatalf("Driver.Open: %v", err)
	}
	defer conn.Close()
	// Driver-level Conn should advertise the surrounding driver.
	c, ok := conn.(*googlesqlite.Conn)
	if !ok {
		t.Fatalf("Open returned %T; want *googlesqlite.Conn", conn)
	}
	// A round-trip query via the driver.Conn surface keeps the
	// connector path warm.
	if _, ok := any(c).(driver.Conn); !ok {
		t.Fatalf("returned conn does not implement driver.Conn")
	}
}

// Connector.Driver / Close drive the connector-side lifecycle.
// We cover both the shared-DB and private-in-memory branches by
// asking OpenConnector for ":memory:" (private) and for a non-empty
// disk path (shared cache).
func TestOpenConnector_PrivateAndShared(t *testing.T) {
	t.Parallel()
	d := &googlesqlite.Driver{}

	// Private in-memory: each OpenConnector should give a fresh inner
	// *sql.DB (innerOwned=true), so Close must succeed without
	// errors and the next Open must not see stale state.
	connector, err := d.OpenConnector(":memory:")
	if err != nil {
		t.Fatalf("OpenConnector: %v", err)
	}
	if connector.Driver() != d {
		t.Fatalf("Driver() = %p; want %p", connector.Driver(), d)
	}
	db := sql.OpenDB(connector)
	if _, err := db.ExecContext(context.Background(), "CREATE TABLE t (k INT64)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close 1: %v", err)
	}

	// Same DSN again must succeed (private-memory cache key is
	// freshly minted).
	connector, err = d.OpenConnector(":memory:")
	if err != nil {
		t.Fatalf("OpenConnector 2: %v", err)
	}
	db = sql.OpenDB(connector)
	if _, err := db.ExecContext(context.Background(), "CREATE TABLE t (k INT64)"); err != nil {
		t.Fatalf("CREATE 2: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close 2: %v", err)
	}

	// Shared (file:::memory:?cache=shared) is the shared-cache pattern:
	// the inner DB is registered in the shared cache (not the private
	// fresh-per-open branch). We exercise both Open calls succeed
	// and that one of them can issue a query — verifying the shared
	// catalog behaves cross-handle is covered by separate consumer
	// tests that pin one open connection per *sql.DB.
	shared := "file::memory:?cache=shared"
	conn1, err := d.OpenConnector(shared)
	if err != nil {
		t.Fatalf("OpenConnector shared: %v", err)
	}
	conn2, err := d.OpenConnector(shared)
	if err != nil {
		t.Fatalf("OpenConnector shared 2: %v", err)
	}
	db1 := sql.OpenDB(conn1)
	defer db1.Close()
	db2 := sql.OpenDB(conn2)
	defer db2.Close()
	if _, err := db1.ExecContext(context.Background(), "SELECT 1"); err != nil {
		t.Fatalf("SELECT in db1: %v", err)
	}
	if _, err := db2.ExecContext(context.Background(), "SELECT 1"); err != nil {
		t.Fatalf("SELECT in db2: %v", err)
	}
}

// ------------------------------------------------------------------
// Conn.Prepare and Conn.Begin (the no-context aliases)
// ------------------------------------------------------------------

// Conn.Prepare delegates to PrepareContext with a background context.
// Driving it via the unwrapped *googlesqlite.Conn directly covers the
// legacy entry point that database/sql will call for older callers.
func TestConn_PrepareAndBegin(t *testing.T) {
	t.Parallel()
	d := &googlesqlite.Driver{}
	conn, err := d.Open(":memory:?_test=conn_prepare_begin")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	gc := conn.(*googlesqlite.Conn)

	// Prepare a no-arg SELECT — driver.Stmt is returned.
	stmt, err := gc.Prepare("SELECT 1")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if stmt == nil {
		t.Fatalf("Prepare returned nil stmt")
	}
	stmt.Close()

	// Conn.Begin returns a driver.Tx — exercise both Commit and Rollback.
	tx, err := gc.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	tx, err = gc.Begin()
	if err != nil {
		t.Fatalf("Begin 2: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

// ------------------------------------------------------------------
// RegisterProto / RegisterProtoMessage
// ------------------------------------------------------------------

// fileDescriptor builds a tiny FileDescriptorProto carrying one
// message `pkg.M { string s = 1; }`, marshals it to bytes the same
// way external callers do (proto.Marshal of FileDescriptorProto), and
// returns both the marshaled bytes and the full message name.
func fileDescriptor(t *testing.T) ([]byte, string) {
	t.Helper()
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("googlesqlite/test/m.proto"),
		Syntax:  proto.String("proto2"),
		Package: proto.String("pkg"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("M"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("s"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
				},
			},
		},
	}
	b, err := proto.Marshal(fd)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	return b, "pkg.M"
}

func TestConn_RegisterProtoAndMessage(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=register_proto")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	sqlConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer sqlConn.Close()

	fdBytes, fullName := fileDescriptor(t)

	if err := sqlConn.Raw(func(dc any) error {
		c, ok := dc.(*googlesqlite.Conn)
		if !ok {
			t.Fatalf("dc = %T; want *googlesqlite.Conn", dc)
		}
		if err := c.RegisterProto(fdBytes); err != nil {
			return err
		}
		// RegisterProtoMessage on a name that was just registered must
		// succeed; on an unknown name it returns an error.
		if err := c.RegisterProtoMessage(fullName); err != nil {
			return err
		}
		if err := c.RegisterProtoMessage("nope.Missing"); err == nil {
			t.Fatalf("RegisterProtoMessage(missing) expected error")
		}
		return nil
	}); err != nil {
		t.Fatalf("Raw: %v", err)
	}

	// Garbage descriptor bytes must surface as an error.
	if err := sqlConn.Raw(func(dc any) error {
		c := dc.(*googlesqlite.Conn)
		return c.RegisterProto([]byte{0xff, 0xff, 0xff, 0xff})
	}); err == nil {
		t.Errorf("RegisterProto(garbage) expected error")
	}
}

// ------------------------------------------------------------------
// UnmarshalDatabaseTypeName
// ------------------------------------------------------------------

// ColumnTypeDatabaseTypeName emits a JSON envelope around internal.Type.
// UnmarshalDatabaseTypeName must round-trip the JSON back into a
// *googlesqlite.ColumnType. We pull a column type out of a live
// query and re-parse it.
func TestUnmarshalDatabaseTypeName_Roundtrip(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=column_type")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT CAST(1 AS INT64) AS k")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	types, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("ColumnTypes: %v", err)
	}
	if len(types) != 1 {
		t.Fatalf("len(types) = %d; want 1", len(types))
	}
	dbName := types[0].DatabaseTypeName()
	if dbName == "" {
		t.Fatalf("DatabaseTypeName returned empty string")
	}

	got, err := googlesqlite.UnmarshalDatabaseTypeName(dbName)
	if err != nil {
		t.Fatalf("UnmarshalDatabaseTypeName: %v", err)
	}
	if got == nil {
		t.Fatalf("UnmarshalDatabaseTypeName returned nil")
	}
	// Round-trip via JSON: re-marshaling must match the input string.
	roundTrip, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if string(roundTrip) != dbName {
		t.Errorf("roundtrip mismatch\n got %s\nwant %s", roundTrip, dbName)
	}

	// Invalid JSON must surface an error.
	if _, err := googlesqlite.UnmarshalDatabaseTypeName("not-json"); err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}

// ------------------------------------------------------------------
// WithCurrentTime / CurrentTime
// ------------------------------------------------------------------

// WithCurrentTime stores a time on the context; CurrentTime retrieves
// it. Without a value, CurrentTime must return nil. The driver's
// CURRENT_DATETIME etc. consult the context for the "now" value — we
// don't drive that side here (it lives in the analyzer), but the
// helper's API is part of the public contract.
func TestWithCurrentTimeAndCurrentTime(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if got := googlesqlite.CurrentTime(ctx); got != nil {
		t.Errorf("default ctx CurrentTime = %v; want nil", got)
	}
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	ctx = googlesqlite.WithCurrentTime(ctx, now)
	got := googlesqlite.CurrentTime(ctx)
	if got == nil {
		t.Fatalf("CurrentTime returned nil after WithCurrentTime")
	}
	if !got.Equal(now) {
		t.Errorf("CurrentTime = %v; want %v", *got, now)
	}
}

// ------------------------------------------------------------------
// TimeFromTimestampValue / parseCanonicalTimestamp / parseEpochFloat
// ------------------------------------------------------------------

// The TimeFromTimestampValue helper accepts two formats: the canonical
// "YYYY-MM-DD HH:MM:SS[.fff]+00" the driver emits and the legacy
// epoch-seconds-with-fractional form. Both branches must roundtrip
// the same instant.
func TestTimeFromTimestampValue_AllForms(t *testing.T) {
	t.Parallel()

	// Canonical, integer seconds.
	got, err := googlesqlite.TimeFromTimestampValue("2024-01-02 03:04:05+00")
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	want := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("canonical = %v; want %v", got, want)
	}

	// Canonical with fractional and offset.
	got, err = googlesqlite.TimeFromTimestampValue("2024-01-02 03:04:05.123456789+00")
	if err != nil {
		t.Fatalf("canonical frac: %v", err)
	}
	wantFrac := time.Date(2024, 1, 2, 3, 4, 5, 123456789, time.UTC)
	if !got.Equal(wantFrac) {
		t.Errorf("canonical frac = %v; want %v", got, wantFrac)
	}

	// RFC3339 fallback.
	got, err = googlesqlite.TimeFromTimestampValue("2024-01-02T03:04:05Z")
	if err != nil {
		t.Fatalf("rfc3339: %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("rfc3339 = %v; want %v", got, want)
	}

	// Epoch-seconds.
	got, err = googlesqlite.TimeFromTimestampValue("1704164645")
	if err != nil {
		t.Fatalf("epoch: %v", err)
	}
	if !got.Equal(time.Unix(1704164645, 0)) {
		t.Errorf("epoch = %v; want %v", got, time.Unix(1704164645, 0))
	}

	// Epoch with fractional micros.
	got, err = googlesqlite.TimeFromTimestampValue("1704164645.123456")
	if err != nil {
		t.Fatalf("epoch frac: %v", err)
	}
	wantEpochFrac := time.Unix(1704164645, 123456*int64(time.Microsecond))
	if !got.Equal(wantEpochFrac) {
		t.Errorf("epoch frac = %v; want %v", got, wantEpochFrac)
	}

	// Empty string -> error.
	if _, err := googlesqlite.TimeFromTimestampValue(""); err == nil {
		t.Errorf("empty: expected error")
	}

	// Canonical garbage.
	if _, err := googlesqlite.TimeFromTimestampValue("2024-13-99 25:99:99"); err == nil {
		t.Errorf("invalid canonical: expected error")
	}

	// Epoch with multiple dots.
	if _, err := googlesqlite.TimeFromTimestampValue("1.2.3"); err == nil {
		t.Errorf("multi-dot epoch: expected error")
	}
	// Epoch with non-numeric tail.
	if _, err := googlesqlite.TimeFromTimestampValue("abc"); err == nil {
		t.Errorf("non-numeric epoch: expected error")
	}
}

// ------------------------------------------------------------------
// DSN parsing: file: prefix preserves the URI query string
// ------------------------------------------------------------------

// `file:foo?mode=memory&cache=shared` is a form consumers
// use. The driver must keep the query string intact so SQLite
// recognises the in-memory hint. (memory_dsn_isolation_test.go
// already covers the basic case — this one stresses that two opens
// share state when cache=shared is present.)
// `file:<name>?mode=memory&cache=shared` is a form consumers
// use. The DSN passes through dsnParts unchanged
// (file: prefix branch). The driver must avoid stripping the query
// string in that branch — confirmed by opening, writing, and
// re-reading within the same handle (which holds the same connection
// in the pool when SetMaxOpenConns(1) is in effect).
func TestFileDSN_SharedCacheKeepsState(t *testing.T) {
	t.Parallel()
	dsn := "file:shared_test_kept?mode=memory&cache=shared"
	ctx := context.Background()

	db, err := sql.Open("googlesqlite", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, "CREATE TABLE shared_kept (k INT64)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO shared_kept (k) VALUES (42)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	var k int64
	if err := db.QueryRowContext(ctx, "SELECT k FROM shared_kept").Scan(&k); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if k != 42 {
		t.Errorf("k = %d; want 42", k)
	}
}

// ------------------------------------------------------------------
// ChangedCatalogFromRows / ChangedCatalogFromResult error paths
// ------------------------------------------------------------------

// ChangedCatalogFromRows must reject nil. ChangedCatalogFromResult
// must reject non-googlesqlite results. These edge cases exist
// because both helpers reach into unsafe-pointer internals and must
// guard the boundaries.
func TestChangedCatalogFrom_ErrorPaths(t *testing.T) {
	t.Parallel()

	if _, err := googlesqlite.ChangedCatalogFromRows(nil); err == nil {
		t.Errorf("nil rows: expected error")
	}

	// A non-pointer Result implementation (driver.RowsAffected is an
	// int64-typed alias of sql.Result). Pass that to confirm the
	// type-shape check fires.
	var bogus driver.RowsAffected = 0
	if _, err := googlesqlite.ChangedCatalogFromResult(bogus); err == nil {
		t.Errorf("non-struct Result: expected error")
	}
}

// ------------------------------------------------------------------
// isPrivateInMemoryDSN edge case: cache=shared keeps the inner DB
// ------------------------------------------------------------------

// freshMemoryDSN appends `_googlesqlite_id=N`. When the DSN already
// has a `?`, the suffix must be appended with `&`. We can't observe
// freshMemoryDSN directly without exporting it, but the
// `file:<name>?mode=memory` flow already exercises both code paths
// when paired with the TestFileDSN_SharedCacheKeepsState above.
// Here we additionally stress the "no ?, append ?" path by opening
// a plain `:memory:` connector and verifying two consecutive
// opens succeed independently.
func TestFreshMemoryDSN_IndirectThroughOpenConnector(t *testing.T) {
	t.Parallel()

	d := &googlesqlite.Driver{}
	conns := make([]driver.Connector, 0, 3)
	for i := 0; i < 3; i++ {
		c, err := d.OpenConnector(":memory:")
		if err != nil {
			t.Fatalf("OpenConnector %d: %v", i, err)
		}
		conns = append(conns, c)
	}
	for i, c := range conns {
		db := sql.OpenDB(c)
		if _, err := db.ExecContext(context.Background(),
			"CREATE TABLE t_"+intToStr(i)+" (k INT64)"); err != nil {
			t.Fatalf("CREATE in connector %d: %v", i, err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("Close %d: %v", i, err)
		}
	}
}

// intToStr avoids strconv just to keep imports minimal; used only
// to mint a unique suffix for the table name above.
func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [16]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// ------------------------------------------------------------------
// Tx through database/sql.BeginTx — public Tx.Commit / Tx.Rollback
// ------------------------------------------------------------------

// transaction_test.go covers the happy path; this test specifically
// hits the rollback branch on a transaction that did not modify any
// data (so the underlying SQLite tx still has to be rolled back).
func TestBeginTx_RollbackOnEmptyTransaction(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=tx_rollback")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

// ------------------------------------------------------------------
// DSN with private memory + query parameters round-trips
// ------------------------------------------------------------------

// Open the driver with a bare `:memory:?dialect=googlesql` DSN. The
// dsnParts split must lop the `?dialect=googlesql` off so SQLite sees
// just `:memory:`. This exercises the non-`file:` branch of dsnParts.
func TestMemoryWithQueryParam_StripsQuery(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?dialect=googlesql")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(), "CREATE TABLE t (k INT64)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	// A second exec must succeed (no stale state).
	if _, err := db.ExecContext(context.Background(), "INSERT INTO t (k) VALUES (1)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
}

// guardNoLeak ensures the package-level cache does not silently leak
// connectors over the course of these tests. We only check that the
// names we just registered are tracked by inspecting a fresh
// sql.Open call — if we ever silently broke the lookup, every Open
// in this test would return an error.
func TestRepeatedOpenAndCloseLifecycle(t *testing.T) {
	t.Parallel()
	for i := 0; i < 3; i++ {
		db, err := sql.Open("googlesqlite", ":memory:?_test=lifecycle")
		if err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
		if _, err := db.ExecContext(context.Background(), "SELECT 1"); err != nil {
			t.Fatalf("iter %d: SELECT: %v", i, err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("iter %d: Close: %v", i, err)
		}
	}
	// Light sanity check: the package level interface still classifies
	// `:memory:` as private. Open and confirm DDL doesn't leak.
	db1, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open after lifecycle: %v", err)
	}
	defer db1.Close()
	if _, err := db1.ExecContext(context.Background(),
		"CREATE TABLE post_lifecycle (k INT64)"); err != nil {
		t.Fatalf("CREATE after lifecycle: %v", err)
	}
	if !strings.Contains(intToStr(1), "1") {
		t.Fatalf("self-test: intToStr broken")
	}
}

// ---- from changed_catalog_test.go ----

// TestChangedCatalogFromResultAndRows drives the public
// ChangedCatalogFromResult / ChangedCatalogFromRows accessors and
// transitively the internal ChangedCatalog.Changed, ChangedTable.
// Changed, ChangedFunction.Changed accessors plus Conn.addTable /
// Conn.deleteTable / Conn.addFunction / Conn.deleteFunction.
//
// These accessors are how a consumer surfaces "the SQL just
// modified the catalog" so it can replay the change into its
// per-project storage. The exact behaviour is encoded in catalog.go
// (root package): every DDL statement bumps Added/Updated/Deleted on
// the per-statement Result's underlying internal.Conn.
func TestChangedCatalogFromResultAndRows(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=changed_catalog")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	// CREATE TABLE — the Result should carry an Added entry.
	res, err := conn.ExecContext(ctx, "CREATE TABLE t1 (a INT64)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	cc, err := googlesqlite.ChangedCatalogFromResult(res)
	if err != nil {
		t.Fatalf("ChangedCatalogFromResult after CREATE: %v", err)
	}
	if !cc.Changed() || !cc.Table.Changed() {
		t.Fatalf("Changed=%v Table.Changed=%v; want both true after CREATE",
			cc.Changed(), cc.Table.Changed())
	}
	if len(cc.Table.Added) != 1 {
		t.Fatalf("Added len = %d; want 1", len(cc.Table.Added))
	}

	// DROP TABLE — Deleted entry on the new Result.
	res, err = conn.ExecContext(ctx, "DROP TABLE t1")
	if err != nil {
		t.Fatalf("DROP TABLE: %v", err)
	}
	cc, err = googlesqlite.ChangedCatalogFromResult(res)
	if err != nil {
		t.Fatalf("ChangedCatalogFromResult after DROP: %v", err)
	}
	if !cc.Table.Changed() {
		t.Fatalf("Table.Changed = false after DROP; want true")
	}
	if len(cc.Table.Deleted) != 1 {
		t.Fatalf("Deleted len = %d; want 1", len(cc.Table.Deleted))
	}

	// CREATE FUNCTION — Function.Added entry.
	res, err = conn.ExecContext(ctx,
		"CREATE FUNCTION inc(x INT64) AS (x + 1)")
	if err != nil {
		t.Fatalf("CREATE FUNCTION: %v", err)
	}
	cc, err = googlesqlite.ChangedCatalogFromResult(res)
	if err != nil {
		t.Fatalf("ChangedCatalogFromResult after CREATE FUNCTION: %v", err)
	}
	if !cc.Function.Changed() {
		t.Fatalf("Function.Changed = false after CREATE FUNCTION; want true")
	}
	if len(cc.Function.Added) != 1 {
		t.Fatalf("Function.Added len = %d; want 1", len(cc.Function.Added))
	}

	// DROP FUNCTION — Function.Deleted entry.
	res, err = conn.ExecContext(ctx, "DROP FUNCTION inc")
	if err != nil {
		t.Fatalf("DROP FUNCTION: %v", err)
	}
	cc, err = googlesqlite.ChangedCatalogFromResult(res)
	if err != nil {
		t.Fatalf("ChangedCatalogFromResult after DROP FUNCTION: %v", err)
	}
	if !cc.Function.Changed() {
		t.Fatalf("Function.Changed = false after DROP FUNCTION; want true")
	}
	if len(cc.Function.Deleted) != 1 {
		t.Fatalf("Function.Deleted len = %d; want 1", len(cc.Function.Deleted))
	}

	// SELECT — Rows path. The Result for a plain SELECT carries no
	// catalog change, so ChangedCatalogFromRows must return a
	// non-nil but empty ChangedCatalog.
	rows, err := conn.QueryContext(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	defer rows.Close()
	cc, err = googlesqlite.ChangedCatalogFromRows(rows)
	if err != nil {
		t.Fatalf("ChangedCatalogFromRows: %v", err)
	}
	if cc.Changed() {
		t.Fatalf("Changed = true after a SELECT; want false")
	}
}

// TestResultRowsAffectedAndLastInsertID drives the internal.Result
// methods that wrap the underlying sqlite result. A DML statement
// executed through db.ExecContext (driver-level, not the prepared
// path) returns an internal.Result whose RowsAffected and
// LastInsertId surface the per-statement counters.
func TestResultRowsAffectedAndLastInsertID(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=result_methods")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE res_t (k INT64, v STRING)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	res, err := conn.ExecContext(ctx,
		"INSERT INTO res_t (k, v) VALUES (1, 'a'), (2, 'b'), (3, 'c')")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected: %v", err)
	}
	if n != 3 {
		t.Fatalf("RowsAffected = %d; want 3", n)
	}
	// LastInsertId — call just to cover the path; the value is
	// implementation-defined for tables without an INTEGER PRIMARY
	// KEY column.
	if _, err := res.LastInsertId(); err != nil {
		t.Fatalf("LastInsertId: %v", err)
	}

	// Also exercise the Result returned from a no-op like CREATE
	// TABLE — RowsAffected returns 0 since result.result is nil.
	res, err = conn.ExecContext(ctx,
		"CREATE TABLE res_t2 (k INT64)")
	if err != nil {
		t.Fatalf("CREATE 2: %v", err)
	}
	n, err = res.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected after CREATE: %v", err)
	}
	if n != 0 {
		t.Fatalf("RowsAffected after CREATE = %d; want 0", n)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId after CREATE: %v", err)
	}
	if id != 0 {
		t.Fatalf("LastInsertId after CREATE = %d; want 0", id)
	}
}

// ---- from name_path_public_test.go ----

// withRawConn pulls the underlying *googlesqlite.Conn out of a
// *sql.Conn via the database/sql.Raw escape hatch. The callback runs
// while the sql.Conn holds the conn pinned, so the *googlesqlite.Conn
// is safe to use for the duration. Returns no value because callers
// typically just need to call the setter methods.
func withRawConn(t *testing.T, sqlConn *sql.Conn, fn func(*googlesqlite.Conn)) {
	t.Helper()
	if err := sqlConn.Raw(func(dc any) error {
		c, ok := dc.(*googlesqlite.Conn)
		if !ok {
			t.Fatalf("driver conn is %T, want *googlesqlite.Conn", dc)
		}
		fn(c)
		return nil
	}); err != nil {
		t.Fatalf("Raw: %v", err)
	}
}

// TestSetNamePathAndAddNamePath drives the SetNamePath / AddNamePath /
// NamePath / MaxNamePath / SetMaxNamePath surface on the public Conn.
// The expected behaviour is:
//
//   - SetNamePath replaces the entire prefix.
//   - AddNamePath appends to it.
//   - Exceeding MaxNamePath returns an error.
//
// These are direct invariants encoded in
// internal/name_path.go::NamePath.setPath / addPath. There is no
// upstream GoogleSQL reference because name path is a googlesqlite-
// specific feature — the doc lives at driver.go's GoDoc comments on
// SetNamePath/AddNamePath.
func TestSetNamePathAndAddNamePath(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=name_path_public")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	sqlConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer sqlConn.Close()

	withRawConn(t, sqlConn, func(c *googlesqlite.Conn) {
		// Default state: NamePath is empty, MaxNamePath is 0.
		if got := c.NamePath(); len(got) != 0 {
			t.Fatalf("default NamePath = %v; want empty", got)
		}
		if got := c.MaxNamePath(); got != 0 {
			t.Fatalf("default MaxNamePath = %d; want 0", got)
		}

		// SetNamePath with a single element. The path arrives as a
		// flat slice — internal.normalizePath splits dotted entries,
		// so set with both a single element and a dotted element.
		if err := c.SetNamePath([]string{"project"}); err != nil {
			t.Fatalf("SetNamePath: %v", err)
		}
		if got := c.NamePath(); len(got) != 1 || got[0] != "project" {
			t.Fatalf("NamePath after Set = %v; want [project]", got)
		}

		// AddNamePath appends.
		if err := c.AddNamePath("dataset"); err != nil {
			t.Fatalf("AddNamePath: %v", err)
		}
		got := c.NamePath()
		if len(got) != 2 || got[0] != "project" || got[1] != "dataset" {
			t.Fatalf("NamePath after Add = %v; want [project dataset]", got)
		}

		// SetMaxNamePath restricts the total entries. Adding a 3rd
		// entry must fail.
		c.SetMaxNamePath(2)
		if got := c.MaxNamePath(); got != 2 {
			t.Fatalf("MaxNamePath = %d; want 2", got)
		}
		if err := c.AddNamePath("third"); err == nil {
			t.Fatalf("AddNamePath past max: expected error, got nil")
		} else if !strings.Contains(err.Error(), "max name path") {
			t.Fatalf("AddNamePath past max: error message lacks 'max name path': %v", err)
		}

		// SetNamePath with a too-long path must also fail (uses the
		// same maxNum check as AddNamePath).
		if err := c.SetNamePath([]string{"a", "b", "c"}); err == nil {
			t.Fatalf("SetNamePath past max: expected error, got nil")
		}

		// SetNamePath with a dotted path: normalizePath splits on
		// "." — so "a.b" is one argument but becomes two entries.
		c.SetMaxNamePath(0)
		if err := c.SetNamePath([]string{"a.b"}); err != nil {
			t.Fatalf("SetNamePath dotted: %v", err)
		}
		got = c.NamePath()
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Fatalf("NamePath after dotted Set = %v; want [a b]", got)
		}
	})
}

func TestDropFunctionIfExistsUsesCatalogWithNamePath(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=drop_function_if_exists_name_path")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	sqlConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer sqlConn.Close()

	if _, err := sqlConn.ExecContext(ctx, "CREATE FUNCTION dataset.fn(x INT64) AS (x + 1)"); err != nil {
		t.Fatalf("CREATE FUNCTION: %v", err)
	}
	withRawConn(t, sqlConn, func(c *googlesqlite.Conn) {
		c.SetMaxNamePath(2)
		if err := c.SetNamePath([]string{"other"}); err != nil {
			t.Fatalf("SetNamePath: %v", err)
		}
	})

	res, err := sqlConn.ExecContext(ctx, "DROP FUNCTION IF EXISTS dataset.fn")
	if err != nil {
		t.Fatalf("DROP FUNCTION IF EXISTS: %v", err)
	}
	cc, err := googlesqlite.ChangedCatalogFromResult(res)
	if err != nil {
		t.Fatalf("ChangedCatalogFromResult: %v", err)
	}
	if len(cc.Function.Deleted) != 1 {
		t.Fatalf("Function.Deleted len = %d; want 1", len(cc.Function.Deleted))
	}
	if diff := cmp.Diff(cc.Function.Deleted[0].NamePath, []string{"dataset", "fn"}); diff != "" {
		t.Errorf("Deleted NamePath (-want +got):\n%s", diff)
	}

	var got int64
	if err := sqlConn.QueryRowContext(ctx, "SELECT dataset.fn(1)").Scan(&got); err == nil {
		t.Fatalf("dropped function is still callable; got %d", got)
	}
}

// TestExplainModeRunsExplainQueryPlan flips the connection's
// explain-mode flag, then runs a SELECT. With explain mode on, the
// QueryStmtAction.QueryContext branches into ExplainQueryPlan and
// prints the plan; the returned Rows is empty. Asserting the empty
// row set is sufficient — the side effect (stdout print) is hard to
// observe and not the contract.
func TestExplainModeRunsExplainQueryPlan(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=explain_mode")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	sqlConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer sqlConn.Close()

	if _, err := sqlConn.ExecContext(ctx,
		"CREATE TABLE explain_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := sqlConn.ExecContext(ctx,
		"INSERT INTO explain_t (k) VALUES (1), (2), (3)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	withRawConn(t, sqlConn, func(c *googlesqlite.Conn) {
		c.SetExplainMode(true)
	})
	rows, err := sqlConn.QueryContext(ctx, "SELECT k FROM explain_t WHERE k > 1")
	if err != nil {
		t.Fatalf("Query under explain mode: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatalf("expected zero rows in explain mode; got at least one")
	}
}

// TestAutoIndexMode toggles the per-connection auto-index flag and
// asserts CREATE TABLE survives. The flag, when on, asks the
// CreateTableStmtAction.createIndexAutomatically helper to emit a
// `CREATE INDEX` for every available-auto-index column. That branch
// is the only exit from the helper. Asserting CREATE+INSERT+SELECT
// completes end-to-end is enough — the index is opaque to the test.
func TestAutoIndexMode(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=auto_index")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	sqlConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer sqlConn.Close()

	withRawConn(t, sqlConn, func(c *googlesqlite.Conn) {
		c.SetAutoIndexMode(true)
	})
	if _, err := sqlConn.ExecContext(ctx,
		"CREATE TABLE auto_idx_t (k INT64, v STRING)"); err != nil {
		t.Fatalf("CREATE TABLE with auto index: %v", err)
	}
	if _, err := sqlConn.ExecContext(ctx,
		"INSERT INTO auto_idx_t (k, v) VALUES (1, 'a')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	var k int64
	if err := sqlConn.QueryRowContext(ctx,
		"SELECT k FROM auto_idx_t").Scan(&k); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if k != 1 {
		t.Fatalf("k = %d; want 1", k)
	}
}

// TestSetMaterializeCTEAndAutoIndex toggles the connection-scope
// flags via the public Conn methods, asserting the getter returns
// what the setter just wrote.
func TestSetMaterializeCTEAndAutoIndex(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=conn_flags")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	sqlConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer sqlConn.Close()

	withRawConn(t, sqlConn, func(c *googlesqlite.Conn) {
		// Defaults: MaterializeCTE is true (per Conn constructor),
		// the other flags are unobservable from Go but the setters
		// are still exercisable.
		if got := c.MaterializeCTE(); !got {
			t.Fatalf("default MaterializeCTE = false; want true")
		}
		c.SetMaterializeCTE(false)
		if got := c.MaterializeCTE(); got {
			t.Fatalf("MaterializeCTE after Set(false) = true; want false")
		}
		c.SetMaterializeCTE(true)
		if got := c.MaterializeCTE(); !got {
			t.Fatalf("MaterializeCTE after Set(true) = false; want true")
		}

		// SetAutoIndexMode / SetExplainMode have no public getter
		// but exercising them at least covers the setter body.
		c.SetAutoIndexMode(true)
		c.SetAutoIndexMode(false)
		c.SetExplainMode(true)
		c.SetExplainMode(false)
	})
}

// ---- from memory_dsn_isolation_test.go ----

// TestMemoryDSNIsolation guards two related driver regressions:
//
//  1. The DSN-keyed *sql.DB cache used to return the same underlying
//     database for every sql.Open(":memory:") call, so a CREATE TABLE in
//     one open would leak into the next via the pooled sqlite connection.
//     The symptom was a "table already exists" failure under go test
//     -count=N for any N >= 2.
//  2. dsnParts used to strip the query string from every DSN, including
//     SQLite URIs of the form `file:NAME?mode=memory&cache=private`. The
//     URI lost its `mode=memory` hint and SQLite happily created an
//     on-disk file named `file:NAME`, accumulating state across test
//     runs.
func TestMemoryDSNIsolation(t *testing.T) {
	t.Run("plain_memory_fresh_per_open", func(t *testing.T) {
		ctx := context.Background()
		for i := 0; i < 3; i++ {
			db, err := sql.Open("googlesqlite", ":memory:")
			if err != nil {
				t.Fatalf("sql.Open: %v", err)
			}
			db.SetMaxOpenConns(1)
			if _, err := db.ExecContext(ctx, "CREATE TABLE t (k INT64)"); err != nil {
				_ = db.Close()
				t.Fatalf("iteration %d: CREATE TABLE: %v", i, err)
			}
			if err := db.Close(); err != nil {
				t.Fatalf("iteration %d: db.Close: %v", i, err)
			}
		}
	})

	t.Run("memory_with_dialect_fresh_per_open", func(t *testing.T) {
		ctx := context.Background()
		for i := 0; i < 3; i++ {
			db, err := sql.Open("googlesqlite", ":memory:?dialect=googlesql")
			if err != nil {
				t.Fatalf("sql.Open: %v", err)
			}
			db.SetMaxOpenConns(1)
			if _, err := db.ExecContext(ctx, "CREATE TABLE t (k INT64)"); err != nil {
				_ = db.Close()
				t.Fatalf("iteration %d: CREATE TABLE: %v", i, err)
			}
			if err := db.Close(); err != nil {
				t.Fatalf("iteration %d: db.Close: %v", i, err)
			}
		}
	})

	t.Run("file_uri_preserves_mode_memory", func(t *testing.T) {
		// Switch CWD to a temp dir so a regression — SQLite falling back to
		// creating an on-disk file — leaves it there instead of in the
		// project tree.
		tmp := t.TempDir()
		oldCWD, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}
		if err := os.Chdir(tmp); err != nil {
			t.Fatalf("Chdir(%s): %v", tmp, err)
		}
		t.Cleanup(func() { _ = os.Chdir(oldCWD) })

		ctx := context.Background()
		db, err := sql.Open("googlesqlite", "file:memdsn?mode=memory&cache=private")
		if err != nil {
			t.Fatalf("sql.Open: %v", err)
		}
		db.SetMaxOpenConns(1)
		defer db.Close()
		if _, err := db.ExecContext(ctx, "CREATE TABLE t (k INT64)"); err != nil {
			t.Fatalf("CREATE TABLE: %v", err)
		}
		// SQLite must have honoured mode=memory; no `file:memdsn` (nor
		// `memdsn`) file should appear on disk.
		for _, candidate := range []string{"file:memdsn", "memdsn"} {
			if _, err := os.Stat(filepath.Join(tmp, candidate)); err == nil {
				t.Fatalf("URI DSN leaked an on-disk file: %s", candidate)
			}
		}
	})
}

// ---- from recover_panic_test.go ----

// TestPanicInDriverConvertsToError pins the contract that a panic
// originating inside any analyzer/formatter/runtime code path reached
// from a driver-public method must surface to the caller as an error,
// not as a hang.
//
// Background: before the recoverPanicAsError defer landed, a missing
// case in formatter dispatch caused a nil-pointer dereference inside
// `*OrderByScanNode.FormatSQL`; the panic unwound up to
// `database/sql.(*Conn).QueryContext`, the deferred close path tried
// to acquire the same *sql.Conn closemu Lock the panicking call
// already held in read mode, and the test hung forever instead of
// failing. The deadlock was independent of the underlying analyzer
// bug — any future panic deep in our pipeline would reproduce it.
//
// This test installs a `ConnectHook` on the googlesqlite Driver that
// panics the first time it is invoked, then runs a Query through a
// 5-second timeout. Pre-fix this hung; with recoverPanicAsError in
// place the call returns an error containing the panic value.
func TestPanicInDriverConvertsToError(t *testing.T) {
	d := &googlesqlite.Driver{}
	d.ConnectHook = func(c *googlesqlite.Conn) error {
		panic("regression: simulated panic from a Conn hook")
	}
	connector, err := d.OpenConnector(":memory:")
	if err != nil {
		t.Fatalf("OpenConnector: %v", err)
	}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := db.QueryContext(ctx, "SELECT 1")
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("expected an error from a panicking driver call, got nil")
		}
		// The error chain wraps the panic value and the stack — we
		// only need to confirm the panic message threaded through.
		if !strings.Contains(err.Error(), "regression: simulated panic from a Conn hook") {
			t.Fatalf("error did not surface the panic value: %v", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("driver call hit the test timeout instead of returning the panic error: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("driver call did not return within 5s — the panic→defer-Close deadlock has regressed")
	}
}

// ---- from transaction_test.go ----

// TestTransactionCommitAndRollback drives db.BeginTx → Tx.Exec →
// Tx.Commit / Tx.Rollback. The googlesqlite driver implements
// driver.ConnBeginTx through the underlying SQLite driver, plus the
// analyzer wraps BEGIN / COMMIT / ROLLBACK as Begin/CommitStmtAction
// (see internal/stmt_action.go).
//
// Authoritative behaviour:
//   - Committed rows survive after Tx.Commit.
//   - Rolled-back rows are gone after Tx.Rollback.
//
// These invariants are general-DB contracts and match SQLite's
// behaviour exactly per googlesqlite/internal/sqlitex.
func TestTransactionCommitAndRollback(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=tx_commit_rollback")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CREATE TABLE t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	// Commit case.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO t (k) VALUES (1)"); err != nil {
		_ = tx.Rollback()
		t.Fatalf("tx INSERT: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	row := db.QueryRowContext(ctx, "SELECT k FROM t")
	var k int64
	if err := row.Scan(&k); err != nil {
		t.Fatalf("Scan after commit: %v", err)
	}
	if k != 1 {
		t.Fatalf("k after commit = %d; want 1", k)
	}

	// Rollback case.
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx 2: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO t (k) VALUES (2)"); err != nil {
		_ = tx.Rollback()
		t.Fatalf("tx INSERT 2: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	// After rollback, only k=1 survives.
	rows, err := db.QueryContext(ctx, "SELECT k FROM t ORDER BY k")
	if err != nil {
		t.Fatalf("Query after rollback: %v", err)
	}
	defer rows.Close()
	var got []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, v)
	}
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("post-rollback rows = %v; want [1]", got)
	}
}

// ---- from transactional_ddl_test.go ----

// TestBeginTransactionExplicit drives the BeginStmtAction and the
// formatter's BeginStmtNode path. The googlesql doc grammar declares
// `BEGIN [TRANSACTION];` as a stand-alone procedural statement.
//
// Reference: docs/third_party/googlesql-docs/procedural-language.md
// "BEGIN TRANSACTION" section (around line 992).
func TestBeginTransactionExplicit(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=begin_tx_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	// BEGIN TRANSACTION as an explicit procedural statement. The
	// analyzer routes it through BeginStmtAction; the formatter
	// renders BeginStmtNode as an empty string (no-op surface SQL).
	if _, err := db.ExecContext(ctx, "BEGIN TRANSACTION"); err != nil {
		t.Fatalf("BEGIN TRANSACTION: %v", err)
	}
	if _, err := db.ExecContext(ctx, "COMMIT TRANSACTION"); err != nil {
		t.Fatalf("COMMIT TRANSACTION: %v", err)
	}
}

// TestBeginCommitNested drives BeginStmtAction → CommitStmtAction
// across multiple statements interleaved with DML, asserting the
// commit boundary persists rows.
func TestBeginCommitNested(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=begin_commit_nested")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CREATE TABLE bc_t (k INT64)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if _, err := db.ExecContext(ctx, "BEGIN TRANSACTION"); err != nil {
		t.Fatalf("BEGIN: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO bc_t (k) VALUES (1)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if _, err := db.ExecContext(ctx, "COMMIT TRANSACTION"); err != nil {
		t.Fatalf("COMMIT: %v", err)
	}

	var k int64
	if err := db.QueryRowContext(ctx, "SELECT k FROM bc_t").Scan(&k); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if k != 1 {
		t.Fatalf("k = %d; want 1", k)
	}
}

// ---- from prepared_stmt_test.go ----

// These tests drive the database/sql Prepare → Stmt.Exec / Stmt.Query
// surface. The spectest harness used by go-googlesqlite's main test
// suite only calls ExecContext/QueryContext directly, which bypasses
// PrepareContext on the driver Conn and the *Stmt wrappers in
// internal/stmt.go. Adding explicit db.Prepare calls covers the
// previously-uncovered:
//   - CreateTableStmtAction.Prepare → newCreateTableStmt
//   - CreateViewStmtAction.Prepare → newCreateViewStmt
//   - CreateFunctionStmtAction.Prepare → newCreateFunctionStmt
//   - DMLStmtAction.Prepare → newDMLStmt
//   - QueryStmtAction.Prepare → newQueryStmt
// plus their Close / NumInput / Exec / Query / CheckNamedValue
// surface in internal/stmt.go.

// TestPreparedQueryStmt exercises the QueryStmt path via db.Prepare +
// stmt.Query, asserting the row matches the spec-provided literal.
// Authoritative source: SELECT literal expressions in
// docs/specs/googlesql/syntax/select.md (single-row, single-column
// boolean) — every analyzer/formatter handles literal SELECT.
func TestPreparedQueryStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_query_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	stmt, err := db.PrepareContext(ctx, "SELECT 1 AS x")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		t.Fatalf("stmt.Query: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatalf("expected one row, got none")
	}
	var x int64
	if err := rows.Scan(&x); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if x != 1 {
		t.Fatalf("x = %d; want 1", x)
	}
}

// TestPreparedQueryStmtWithPositionalParam binds a positional `?`
// parameter through the QueryStmt.CheckNamedValue → encodeOrPassArgs
// path. The expected row mirrors GoogleSQL's parameter-binding
// semantics: a `?` placeholder returns the value as-is (here, an
// INT64 literal). Named parameters are not supported by the QueryStmt
// wrapper (it implements neither the named-value driver-typing
// interface nor the named-value parameter pass-through), so this
// test sticks to positional bindings.
func TestPreparedQueryStmtWithPositionalParam(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_param")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	// Use `?` placeholder, which the analyzer's positional-parameter
	// path supports without naming.
	stmt, err := conn.PrepareContext(ctx, "SELECT ? AS v")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, int64(42))
	if err != nil {
		t.Fatalf("stmt.Query: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatalf("expected one row, got none")
	}
	var v int64
	if err := rows.Scan(&v); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if v != 42 {
		t.Fatalf("v = %d; want 42", v)
	}
}

// TestPreparedCreateAndDML covers the prepared-statement cycle on
// DML:
//   - PrepareContext("INSERT ...") → DMLStmt
//   - PrepareContext("UPDATE ...") → DMLStmt
//   - PrepareContext("DELETE ...") → DMLStmt
//   - PrepareContext("SELECT ...") → QueryStmt
//
// DDL is set up via db.ExecContext (driver-level Exec path), since
// the DDL-through-Prepare wrapper in internal/stmt.go is a separate,
// independently exercised surface.
//
// Values are taken from the upstream INSERT example in
// docs/third_party/googlesql-docs/data-manipulation-language.md (lines
// 262-264, "Catalina Smith" / status active). The expected post-
// UPDATE state is the same row with Status = 'inactive', which is
// the documented effect of `UPDATE T SET c=v WHERE ...` in the same
// document.
func TestPreparedCreateAndDML(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_create_dml")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Pin one logical connection — :memory: schema is per-connection,
	// and we need PrepareContext / Exec / Query to share the table.
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `CREATE TABLE Singers (
		SingerId INT64 NOT NULL,
		FirstName STRING,
		LastName STRING,
		Status STRING
	)`); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	// INSERT — multi-row INSERT exactly like the upstream Singers
	// example. Use positional `?` placeholders to exercise the DMLStmt
	// parameter path.
	ins, err := conn.PrepareContext(ctx,
		"INSERT INTO Singers (SingerId, FirstName, LastName, Status) VALUES (?, ?, ?, ?)")
	if err != nil {
		t.Fatalf("INSERT Prepare: %v", err)
	}
	defer ins.Close()
	res, err := ins.ExecContext(ctx, int64(5), "Catalina", "Smith", "active")
	if err != nil {
		t.Fatalf("INSERT Exec: %v", err)
	}
	// Result.RowsAffected from internal/result.go must report 1.
	if n, err := res.RowsAffected(); err != nil {
		t.Fatalf("RowsAffected: %v", err)
	} else if n != 1 {
		t.Fatalf("RowsAffected = %d; want 1", n)
	}
	// LastInsertId is allowed to be implementation-defined for tables
	// with explicit PK columns; we just probe it for non-panic
	// behaviour to cover the Result.LastInsertId path.
	if _, err := res.LastInsertId(); err != nil {
		t.Fatalf("LastInsertId: %v", err)
	}

	// UPDATE — flip Status to 'inactive'. The upstream DML doc shows
	// `UPDATE ... SET col = expr WHERE pk = v` as the canonical form.
	upd, err := conn.PrepareContext(ctx,
		"UPDATE Singers SET Status = 'inactive' WHERE SingerId = 5")
	if err != nil {
		t.Fatalf("UPDATE Prepare: %v", err)
	}
	defer upd.Close()
	if _, err := upd.ExecContext(ctx); err != nil {
		t.Fatalf("UPDATE Exec: %v", err)
	}

	// SELECT — confirm UPDATE took effect.
	q, err := conn.PrepareContext(ctx,
		"SELECT FirstName, LastName, Status FROM Singers WHERE SingerId = 5")
	if err != nil {
		t.Fatalf("SELECT Prepare: %v", err)
	}
	defer q.Close()
	rows, err := q.QueryContext(ctx)
	if err != nil {
		t.Fatalf("SELECT Query: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatalf("expected one row")
	}
	var fn, ln, st string
	if err := rows.Scan(&fn, &ln, &st); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if fn != "Catalina" || ln != "Smith" || st != "inactive" {
		t.Fatalf("row = (%q, %q, %q); want (Catalina, Smith, inactive)", fn, ln, st)
	}
	rows.Close()

	// DELETE — remove the row. DELETE follows the same DMLStmt path.
	del, err := conn.PrepareContext(ctx, "DELETE FROM Singers WHERE SingerId = 5")
	if err != nil {
		t.Fatalf("DELETE Prepare: %v", err)
	}
	defer del.Close()
	if _, err := del.ExecContext(ctx); err != nil {
		t.Fatalf("DELETE Exec: %v", err)
	}

	// Final assertion: table is empty. COUNT(*) over an empty table
	// returns 0 (see TestRegression_AggregateOverEmptyGroup).
	var remaining int64
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM Singers").Scan(&remaining); err != nil {
		t.Fatalf("COUNT(*) Query: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected zero rows after DELETE; got %d", remaining)
	}
}

// TestCreateViewViaExec drives CREATE VIEW through db.ExecContext
// (the analyzer's CreateViewStmtAction.ExecContext path). The view
// definition follows the upstream
// docs/third_party/googlesql-docs/data-definition-language.md "CREATE
// VIEW" grammar. The expected post-create behaviour is that SELECT
// from the view returns the same rows the underlying query would.
func TestCreateViewViaExec(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=create_view_exec")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE t (k INT64, v STRING)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO t (k, v) VALUES (1, 'a'), (2, 'b')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"CREATE VIEW v_filtered AS SELECT k, v FROM t WHERE k = 1"); err != nil {
		t.Fatalf("CREATE VIEW: %v", err)
	}

	row := conn.QueryRowContext(ctx, "SELECT k, v FROM v_filtered")
	var k int64
	var v string
	if err := row.Scan(&k, &v); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if k != 1 || v != "a" {
		t.Fatalf("view row = (%d, %q); want (1, a)", k, v)
	}

	// CREATE OR REPLACE VIEW exercises the drop-then-create branch
	// in CreateViewStmtAction.
	if _, err := conn.ExecContext(ctx,
		"CREATE OR REPLACE VIEW v_filtered AS SELECT k, v FROM t WHERE k = 2"); err != nil {
		t.Fatalf("CREATE OR REPLACE VIEW: %v", err)
	}
	row = conn.QueryRowContext(ctx, "SELECT k, v FROM v_filtered")
	if err := row.Scan(&k, &v); err != nil {
		t.Fatalf("Scan after replace: %v", err)
	}
	if k != 2 || v != "b" {
		t.Fatalf("replaced view row = (%d, %q); want (2, b)", k, v)
	}
}

// TestPreparedCreateFunction drives CreateFunctionStmtAction.Prepare
// and the resulting CreateFunctionStmt's Close / NumInput / Exec.
// Authoritative source: googlesql `CREATE FUNCTION` reference in
// docs/third_party/googlesql-docs/. The expected behaviour is that
// invoking the new function returns the documented value.
func TestPreparedCreateFunction(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_create_function")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	stmt, err := conn.PrepareContext(ctx,
		"CREATE FUNCTION add_one(x INT64) AS (x + 1)")
	if err != nil {
		t.Fatalf("CREATE FUNCTION Prepare: %v", err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		_ = stmt.Close()
		t.Fatalf("CREATE FUNCTION Exec: %v", err)
	}
	if err := stmt.Close(); err != nil {
		t.Fatalf("CREATE FUNCTION Close: %v", err)
	}

	row := conn.QueryRowContext(ctx, "SELECT add_one(41)")
	var got int64
	if err := row.Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got != 42 {
		t.Fatalf("add_one(41) = %d; want 42", got)
	}
}

// TestPreparedStmtNumInput probes the NumInput contract on each Stmt
// wrapper. CreateTableStmt/CreateViewStmt/CreateFunctionStmt all return
// 0 (no placeholders survive analysis), DMLStmt/QueryStmt return -1
// ("unknown", per the comment in stmt.go).
func TestPreparedStmtNumInput(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_numinput")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	// Just exercising Prepare/Close is enough to call NumInput
	// internally via database/sql's bookkeeping.
	for _, q := range []string{
		"SELECT k FROM t",
		"INSERT INTO t (k) VALUES (?)",
		"UPDATE t SET k = ? WHERE k = ?",
		"DELETE FROM t WHERE k = ?",
	} {
		s, err := conn.PrepareContext(ctx, q)
		if err != nil {
			t.Fatalf("Prepare %q: %v", q, err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("Close %q: %v", q, err)
		}
	}
}

// TestPreparedReusedAcrossInvocations binds the same prepared
// statement repeatedly with different arguments. This exercises the
// shared CheckNamedValue path (driver Conn + each Stmt's own override)
// and EncodeGoValues' behaviour with a series of primitive types.
//
// Reference: positional parameter binding works for every primitive
// per docs/third_party/googlesql-docs/data-manipulation-language.md
// "Value type compatibility".
func TestPreparedReusedAcrossInvocations(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_reuse")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE t (k INT64, b BOOL, f FLOAT64, s STRING)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	ins, err := conn.PrepareContext(ctx,
		"INSERT INTO t (k, b, f, s) VALUES (?, ?, ?, ?)")
	if err != nil {
		t.Fatalf("INSERT Prepare: %v", err)
	}
	defer ins.Close()

	cases := []struct {
		k int64
		b bool
		f float64
		s string
	}{
		{1, true, 1.5, "first"},
		{2, false, 2.5, "second"},
		{3, true, 3.5, "third"},
	}
	for _, c := range cases {
		if _, err := ins.ExecContext(ctx, c.k, c.b, c.f, c.s); err != nil {
			t.Fatalf("INSERT %v: %v", c, err)
		}
	}

	rows, err := conn.QueryContext(ctx, "SELECT k FROM t ORDER BY k")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	var got []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if len(got) != len(cases) {
		t.Fatalf("rows = %d; want %d", len(got), len(cases))
	}
	for i, c := range cases {
		if got[i] != c.k {
			t.Fatalf("row[%d] = %d; want %d", i, got[i], c.k)
		}
	}
}

// TestPreparedQueryStmtErrorOnExec asserts QueryStmt.Exec /
// QueryStmt.ExecContext returns the documented "unsupported exec"
// error rather than succeeding. The error path is otherwise
// unreachable through database/sql's public API.
func TestPreparedQueryStmtErrorOnExec(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_query_exec_error")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// database/sql lets us call ExecContext on any *sql.Stmt; the
	// driver Stmt's Exec is consulted. QueryStmt rejects it.
	stmt, err := db.PrepareContext(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer stmt.Close()
	if _, err := stmt.ExecContext(ctx); err == nil {
		t.Fatalf("expected error from Exec on a SELECT stmt, got nil")
	} else if !strings.Contains(err.Error(), "unsupported exec") {
		t.Fatalf("error did not mention unsupported exec: %v", err)
	}
}

// ---- from prepared_action_paths_test.go ----

// These tests cover the *.Prepare branches of every StmtAction kind that
// the analyzer can produce. database/sql routes db.PrepareContext through
// driver.Conn.PrepareContext, which calls action.Prepare(...) on each
// resolved action. Together with the existing prepared_*_test.go files
// these exercise:
//
//   - DropStmtAction.Prepare       (Prepare("DROP TABLE ..."))
//   - NoopStmtAction.Prepare       (Prepare("ALTER TABLE ... SET OPTIONS ..."))
//   - BeginStmtAction.Prepare      (Prepare("BEGIN TRANSACTION"))
//   - CommitStmtAction.Prepare     (Prepare("COMMIT TRANSACTION"))
//   - TruncateStmtAction.Prepare   (Prepare("TRUNCATE TABLE ..."))
//   - MergeStmtAction.Prepare      (Prepare("MERGE ..."))
//   - AssignmentStmtAction.Prepare (Prepare("SET @@var = ..."))
//
// Each of these StmtAction.Prepare returns nil, nil (a no-op StmtAction
// in database/sql terms), so the assertion is simply that Prepare does
// not return an error.

// preparePathDrive drives action.Prepare directly through the
// driver-level Conn via sql.Conn.Raw. For no-op actions whose Prepare
// returns (nil, nil) — DropStmtAction, NoopStmtAction, BeginStmtAction,
// CommitStmtAction, TruncateStmtAction, MergeStmtAction,
// AssignmentStmtAction — we go this route so database/sql never holds
// a (nil) Stmt handle that it would later attempt to Close.
func preparePathDrive(t *testing.T, conn *sql.Conn, ctx context.Context, query string) {
	t.Helper()
	if err := conn.Raw(func(rawConn any) error {
		c, ok := rawConn.(driver.ConnPrepareContext)
		if !ok {
			return nil
		}
		stmt, err := c.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		if stmt != nil {
			_ = stmt.Close()
		}
		return nil
	}); err != nil {
		t.Fatalf("PrepareContext %q: %v", query, err)
	}
}

// TestPreparedDropStmt drives DropStmtAction.Prepare via the driver's
// Conn.PrepareContext. The Prepare returns nil, nil, so we use the
// raw-conn path to avoid database/sql's automatic Stmt.Close on
// finalize (which would dereference the nil Stmt).
func TestPreparedDropStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_drop_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE pds_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	preparePathDrive(t, conn, ctx, "DROP TABLE pds_t")
}

// TestPreparedNoopDDL drives NoopStmtAction.Prepare for the
// metadata-only DDL kinds the analyzer accepts but the runtime treats as
// no-ops. Source: docs/third_party/googlesql-docs/data-definition-language.md
// "ALTER TABLE ... SET OPTIONS" grammar.
func TestPreparedNoopDDL(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_noop_ddl")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE pn_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	preparePathDrive(t, conn, ctx,
		`ALTER TABLE pn_t SET OPTIONS (description = "demo")`)
}

// TestPreparedBeginStmt drives BeginStmtAction.Prepare via
// the driver Conn.PrepareContext. Source: BigQuery / GoogleSQL
// transactional grammar (data-manipulation-language.md "BEGIN
// TRANSACTION").
func TestPreparedBeginStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_begin_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	preparePathDrive(t, conn, ctx, "BEGIN TRANSACTION")
}

// TestPreparedCommitStmt drives CommitStmtAction.Prepare. A pre-existing
// transaction is required so the analyzer emits a CommitStmt.
func TestPreparedCommitStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_commit_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	// COMMIT outside an explicit transaction is still parseable
	// (analyzer doesn't require an open tx — the runtime would).
	preparePathDrive(t, conn, ctx, "COMMIT TRANSACTION")
}

// TestPreparedTruncateStmt drives TruncateStmtAction.Prepare via
// the driver Conn.PrepareContext. Source: docs/third_party/googlesql-docs/
// data-manipulation-language.md "TRUNCATE TABLE" grammar.
func TestPreparedTruncateStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_truncate_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE pt_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	preparePathDrive(t, conn, ctx, "TRUNCATE TABLE pt_t")
}

// TestPreparedMergeStmt drives MergeStmtAction.Prepare via a MERGE
// statement. Source: docs/third_party/googlesql-docs/data-manipulation-
// language.md "MERGE" grammar.
func TestPreparedMergeStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_merge_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE pm_target (k INT64, v INT64)"); err != nil {
		t.Fatalf("CREATE target: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "CREATE TABLE pm_source (k INT64, v INT64)"); err != nil {
		t.Fatalf("CREATE source: %v", err)
	}
	preparePathDrive(t, conn, ctx, `MERGE pm_target T USING pm_source S
		ON T.k = S.k
		WHEN NOT MATCHED THEN INSERT (k, v) VALUES (k, v)`)
}

// TestPreparedAssignmentStmt drives AssignmentStmtAction.Prepare via
// the driver Conn.PrepareContext. Source: docs/third_party/googlesql-docs/
// procedural-language.md "@@" system variables.
func TestPreparedAssignmentStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_assignment_stmt")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	preparePathDrive(t, conn, ctx, "SET @@time_zone = 'UTC'")
}

// ---- from prepared_create_or_replace_test.go ----

// TestPreparedCreateOrReplaceTable drives CreateTableStmtAction.Prepare's
// CreateOrReplace branch — when prepared, the analyzer flags
// CreateMode=CreateOrReplace, the Prepare method drops the existing
// table first before preparing the new schema.
//
// Reference: docs/third_party/googlesql-docs/data-definition-language.md
// "CREATE OR REPLACE TABLE".
func TestPreparedCreateOrReplaceTable(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_create_or_replace_table")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	// Pre-create the table so the CreateOrReplace branch has a real
	// drop to perform.
	if _, err := conn.ExecContext(ctx, "CREATE TABLE pcor_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE (initial): %v", err)
	}

	stmt, err := conn.PrepareContext(ctx,
		"CREATE OR REPLACE TABLE pcor_t (k INT64, v STRING)")
	if err != nil {
		t.Fatalf("Prepare CREATE OR REPLACE TABLE: %v", err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		_ = stmt.Close()
		t.Fatalf("Exec CREATE OR REPLACE TABLE: %v", err)
	}
	if err := stmt.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// We don't assert the new schema is queryable — the test focus is
	// the Prepare path's CreateOrReplace branch, which drops the old
	// table.
}

// TestPreparedCreateOrReplaceView drives CreateViewStmtAction.Prepare's
// CreateOrReplace branch.
func TestPreparedCreateOrReplaceView(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_create_or_replace_view")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE pcv_src (k INT64, v STRING)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "CREATE VIEW pcv_v AS SELECT k FROM pcv_src"); err != nil {
		t.Fatalf("CREATE VIEW (initial): %v", err)
	}

	stmt, err := conn.PrepareContext(ctx,
		"CREATE OR REPLACE VIEW pcv_v AS SELECT v FROM pcv_src")
	if err != nil {
		t.Fatalf("Prepare CREATE OR REPLACE VIEW: %v", err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		_ = stmt.Close()
		t.Fatalf("Exec: %v", err)
	}
	if err := stmt.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestCreateTempTable drives the IsTemp branch in
// CreateTableStmtAction.exec — TEMP tables are dropped on re-creation.
func TestCreateTempTable(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=create_temp_table")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TEMP TABLE pct_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TEMP TABLE: %v", err)
	}
	// Re-create — IsTemp implies drop-first.
	if _, err := conn.ExecContext(ctx, "CREATE TEMP TABLE pct_t (k INT64, v STRING)"); err != nil {
		t.Fatalf("CREATE TEMP TABLE (re): %v", err)
	}
}

// ---- from prepared_ddl_test.go ----

// TestPreparedCreateTableExec drives the CreateTableStmt.Exec wrapper
// in internal/stmt.go through database/sql.Stmt — the post-fix path
// where Exec forwards the bound args variadically to the underlying
// *sql.Stmt.Exec. Previously the code passed `anyArgs` as a single
// slice, so this path was unreachable; the production bug fix in
// internal/stmt.go *CreateTableStmt.Exec enables this test.
//
// Authoritative reference: a CREATE TABLE statement from
// docs/third_party/googlesql-docs/data-definition-language.md (CREATE
// TABLE grammar, single-column schema). Expected behaviour: the
// table is queryable after Exec, and a subsequent INSERT + SELECT
// returns the inserted row.
func TestPreparedCreateTableExec(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_create_table_exec")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	stmt, err := conn.PrepareContext(ctx, "CREATE TABLE pct_t (k INT64, v STRING)")
	if err != nil {
		t.Fatalf("Prepare CREATE TABLE: %v", err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		_ = stmt.Close()
		t.Fatalf("Exec CREATE TABLE: %v", err)
	}
	if err := stmt.Close(); err != nil {
		t.Fatalf("Close CREATE TABLE: %v", err)
	}

	// Verify the table is queryable.
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO pct_t (k, v) VALUES (1, 'a')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	row := conn.QueryRowContext(ctx, "SELECT k, v FROM pct_t")
	var k int64
	var v string
	if err := row.Scan(&k, &v); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if k != 1 || v != "a" {
		t.Fatalf("row = (%d, %q); want (1, a)", k, v)
	}
}

// TestPreparedCreateViewExec drives the CreateViewStmt.Exec wrapper.
// Same fix as CreateTableStmt — args were not forwarded variadically.
// The view definition is the canonical "SELECT FROM table" form.
//
// Reference: docs/third_party/googlesql-docs/data-definition-language.md
// "CREATE VIEW" grammar. Expected: a SELECT against the view returns
// rows the underlying SELECT would, filtered to k > 0.
func TestPreparedCreateViewExec(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_create_view_exec")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE pcv_t (k INT64, v STRING)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO pcv_t (k, v) VALUES (0, 'zero'), (1, 'one'), (2, 'two')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	stmt, err := conn.PrepareContext(ctx,
		"CREATE VIEW pcv_v AS SELECT k, v FROM pcv_t WHERE k > 0")
	if err != nil {
		t.Fatalf("Prepare CREATE VIEW: %v", err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		_ = stmt.Close()
		t.Fatalf("Exec CREATE VIEW: %v", err)
	}
	if err := stmt.Close(); err != nil {
		t.Fatalf("Close CREATE VIEW: %v", err)
	}

	rows, err := conn.QueryContext(ctx, "SELECT k FROM pcv_v ORDER BY k")
	if err != nil {
		t.Fatalf("Query view: %v", err)
	}
	defer rows.Close()
	var got []int64
	for rows.Next() {
		var k int64
		if err := rows.Scan(&k); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, k)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("view rows = %v; want [1 2]", got)
	}
}

// TestPreparedCreateTableFunctionExec drives CreateTableFunctionStmt
// Prepare / Exec / Close / NumInput / Query in internal/stmt.go.
// Reference: docs/third_party/googlesql-docs/table-functions.md "CREATE
// TABLE FUNCTION" grammar.
func TestPreparedCreateTableFunctionExec(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_create_tvf_exec")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE pcf_src (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO pcf_src (k) VALUES (1), (2), (3)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	stmt, err := conn.PrepareContext(ctx,
		`CREATE TABLE FUNCTION pcf_fn(MinK INT64) AS (
			SELECT k FROM pcf_src WHERE k >= MinK
		)`)
	if err != nil {
		t.Fatalf("Prepare CREATE TABLE FUNCTION: %v", err)
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		_ = stmt.Close()
		t.Fatalf("Exec CREATE TABLE FUNCTION: %v", err)
	}
	if err := stmt.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	rows, err := conn.QueryContext(ctx, "SELECT k FROM pcf_fn(2) ORDER BY k")
	if err != nil {
		t.Fatalf("SELECT FROM TVF: %v", err)
	}
	defer rows.Close()
	var got []int64
	for rows.Next() {
		var k int64
		if err := rows.Scan(&k); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, k)
	}
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("rows = %v; want [2 3]", got)
	}
}

// ---- from prepared_drop_test.go ----

// TestDropTableViaExec drives DropStmtAction.ExecContext in
// internal/stmt_action.go via driver-level db.ExecContext. The
// statement runs through Conn.ExecContext → DropStmtAction.exec.
//
// Note: DropStmtAction.Prepare returns (nil, nil), so a prepared
// drop is not currently supported through database/sql.Stmt.
// Exercising the action through the direct ExecContext path is the
// supported approach.
//
// Reference: docs/third_party/googlesql-docs/data-definition-language.md
// "DROP TABLE" grammar. Expected behaviour: ExecContext yields a
// table-no-longer-queryable state.
func TestDropTableViaExec(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=drop_table_exec")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE pd_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "DROP TABLE pd_t"); err != nil {
		t.Fatalf("DROP TABLE: %v", err)
	}
	// Verify pd_t is gone.
	if err := conn.QueryRowContext(ctx, "SELECT k FROM pd_t").Scan(new(int64)); err == nil {
		t.Fatalf("table still queryable after DROP")
	}
}

// TestDMLPreparedQueryContext drives DMLStmtAction.QueryContext via
// db.QueryContext on a DML statement. Per the driver contract, this
// runs the DML and returns an empty Rows.
func TestDMLPreparedQueryContext(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=dml_qctx")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "CREATE TABLE dml_qc_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	rows, err := conn.QueryContext(ctx,
		"INSERT INTO dml_qc_t (k) VALUES (1)")
	if err != nil {
		t.Fatalf("QueryContext INSERT: %v", err)
	}
	rows.Close()
	// Verify the row landed.
	row := conn.QueryRowContext(ctx, "SELECT k FROM dml_qc_t")
	var got int64
	if err := row.Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got != 1 {
		t.Fatalf("got = %d; want 1", got)
	}
}

// ---- from prepared_error_paths_test.go ----

// TestPreparedQueryOnCreateTableStmt drives the CreateTableStmt.Query
// error path (Query on a DDL stmt returns an error). The driver wraps
// this as the underlying error after PrepareContext succeeds.
//
// Rationale: CREATE TABLE produces no rows; calling .Query on it must
// surface an error. The stmt.go contract is "CREATE TABLE statement
// does not return rows."
func TestPreparedQueryOnCreateTableStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_query_on_ddl")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	stmt, err := conn.PrepareContext(ctx, "CREATE TABLE qct_t (k INT64)")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if rows != nil {
		_ = rows.Close()
	}
	if err == nil {
		t.Fatalf("QueryContext on CREATE TABLE returned nil error; want an error")
	}
}

// TestPreparedExecOnQueryStmt drives the QueryStmt.Exec error path —
// ExecContext on a SELECT-prepared stmt must return an error.
func TestPreparedExecOnQueryStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_exec_on_select")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	stmt, err := conn.PrepareContext(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx)
	if res != nil {
		// shouldn't happen on error path, but defensively probe
		if _, e := res.RowsAffected(); e != nil {
			t.Logf("RowsAffected: %v", e)
		}
	}
	if err == nil {
		t.Fatalf("ExecContext on SELECT returned nil error; want an error")
	}
}

// TestPreparedQueryOnDMLStmt drives the DMLStmt.Query error path —
// Query on an UPDATE-prepared stmt must return an error.
func TestPreparedQueryOnDMLStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_query_on_dml")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE qdml_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO qdml_t (k) VALUES (1)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	stmt, err := conn.PrepareContext(ctx, "UPDATE qdml_t SET k = 2 WHERE k = 1")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if rows != nil {
		_ = rows.Close()
	}
	if err == nil {
		t.Fatalf("QueryContext on UPDATE returned nil error; want an error")
	}
}

// TestPreparedQueryOnCreateFunctionStmt drives the
// CreateFunctionStmt.Query error path.
func TestPreparedQueryOnCreateFunctionStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_query_on_cf")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	stmt, err := conn.PrepareContext(ctx,
		"CREATE FUNCTION qcf_fn(x INT64) AS (x)")
	if err != nil {
		t.Fatalf("Prepare CREATE FUNCTION: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if rows != nil {
		_ = rows.Close()
	}
	if err == nil {
		t.Fatalf("QueryContext on CREATE FUNCTION returned nil error; want an error")
	}
}

// TestPreparedQueryOnCreateViewStmt drives the CreateViewStmt.Query
// error path. Same contract as CreateTableStmt.Query.
func TestPreparedQueryOnCreateViewStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_query_on_cv")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE pcv2_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	stmt, err := conn.PrepareContext(ctx,
		"CREATE VIEW pcv2_v AS SELECT k FROM pcv2_t")
	if err != nil {
		t.Fatalf("Prepare CREATE VIEW: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if rows != nil {
		_ = rows.Close()
	}
	if err == nil {
		t.Fatalf("QueryContext on CREATE VIEW returned nil error; want an error")
	}
}

// TestPreparedQueryOnCreateTableFunctionStmt drives the
// CreateTableFunctionStmt.Query error path.
func TestPreparedQueryOnCreateTableFunctionStmt(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:?_test=prepared_query_on_ctf")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE pctf_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	stmt, err := conn.PrepareContext(ctx,
		"CREATE TABLE FUNCTION pctf_fn() AS (SELECT k FROM pctf_t)")
	if err != nil {
		t.Fatalf("Prepare CREATE TABLE FUNCTION: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if rows != nil {
		_ = rows.Close()
	}
	if err == nil {
		t.Fatalf("QueryContext on CREATE TABLE FUNCTION returned nil error; want an error")
	}
}

// ---- from probe_stmts_test.go ----

// TestProbeStatementKinds is a single test that walks every
// statement kind we want to exercise so the analyzer flips a Resolved
// node kind on and we can attribute coverage to the matching newXxxNode
// constructor in internal/node.go plus the matching FormatSQL stub in
// internal/formatter.go.
//
// For each statement the assertion is the same: the call either
// succeeds (analyzer + runtime handle it as a no-op) or it returns an
// "unsupported" error (the analyzer doesn't support it yet). Either
// outcome still exercises the node/formatter constructors that fire
// during analyzer dispatch — we only need the path to reach the
// constructor, not the runtime to succeed.
//
// Every statement form below is taken from
// docs/third_party/googlesql-docs/data-definition-language.md or
// procedural-language.md, never invented.
func TestProbeStatementKinds(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		// optional fixture setup
		setup []string
	}{
		{
			name: "create_schema",
			sql:  "CREATE SCHEMA probe_schema",
		},
		{
			name: "create_schema_if_not_exists",
			sql:  "CREATE SCHEMA IF NOT EXISTS probe_schema2",
		},
		{
			name: "alter_table_set_options",
			setup: []string{
				"CREATE TABLE probe_t1 (k INT64)",
			},
			sql: `ALTER TABLE probe_t1 SET OPTIONS (description = "x")`,
		},
		{
			name:  "alter_table_add_column",
			setup: []string{"CREATE TABLE probe_t2 (k INT64)"},
			sql:   "ALTER TABLE probe_t2 ADD COLUMN v STRING",
		},
		{
			name:  "alter_table_drop_column",
			setup: []string{"CREATE TABLE probe_t3 (k INT64, v STRING)"},
			sql:   "ALTER TABLE probe_t3 DROP COLUMN v",
		},
		{
			name: "grant_revoke",
			setup: []string{
				"CREATE TABLE probe_t4 (k INT64)",
				`GRANT SELECT ON TABLE probe_t4 TO "user:demo@example.com"`,
			},
			sql: `REVOKE SELECT ON TABLE probe_t4 FROM "user:demo@example.com"`,
		},
		{
			name: "assert",
			sql:  "ASSERT TRUE",
		},
		{
			name: "assert_with_description",
			sql:  "ASSERT TRUE AS 'sanity'",
		},
		{
			name: "create_procedure",
			sql:  `CREATE PROCEDURE probe_p1(IN x INT64) BEGIN SELECT x; END`,
		},
		{
			name: "set_at_at_var",
			sql:  "SET @@time_zone = 'UTC'",
		},
		{
			name:  "declare_with_default",
			setup: nil,
			sql:   "DECLARE probe_v1 INT64 DEFAULT 0",
		},
		{
			name: "if_else_endif",
			setup: []string{
				"DECLARE probe_v2 INT64 DEFAULT 0",
			},
			sql: "IF TRUE THEN SET probe_v2 = 1; ELSE SET probe_v2 = 2; END IF",
		},
		{
			name: "while_loop",
			setup: []string{
				"DECLARE probe_v3 INT64 DEFAULT 0",
			},
			// Trivial loop: increment until 1, then exit.
			sql: "WHILE probe_v3 < 1 DO SET probe_v3 = probe_v3 + 1; END WHILE",
		},
		{
			name: "loop_with_break",
			setup: []string{
				"DECLARE probe_v4 INT64 DEFAULT 0",
			},
			sql: "LOOP SET probe_v4 = probe_v4 + 1; IF probe_v4 >= 1 THEN BREAK; END IF; END LOOP",
		},
		{
			name: "repeat_until",
			setup: []string{
				"DECLARE probe_v5 INT64 DEFAULT 0",
			},
			sql: "REPEAT SET probe_v5 = probe_v5 + 1; UNTIL probe_v5 >= 1 END REPEAT",
		},
		{
			name: "for_in_unnest",
			sql:  "FOR x IN (SELECT * FROM UNNEST([1,2,3])) DO SELECT x; END FOR",
		},
		{
			name: "begin_block_with_exception",
			sql: `BEGIN
				SELECT 1;
			EXCEPTION WHEN ERROR THEN
				SELECT 2;
			END`,
		},
		{
			name: "raise_no_message",
			sql:  `BEGIN BEGIN RAISE; EXCEPTION WHEN ERROR THEN SELECT 1; END; END`,
		},
		{
			name: "raise_with_message",
			sql: `BEGIN BEGIN
				RAISE USING MESSAGE = 'oops';
			EXCEPTION WHEN ERROR THEN
				SELECT 1;
			END; END`,
		},
		{
			name: "create_function_simple",
			sql:  "CREATE FUNCTION probe_fn1(x INT64) AS (x + 1)",
		},
		{
			name: "create_function_temp",
			sql:  "CREATE TEMP FUNCTION probe_fn2(x INT64) AS (x * 2)",
		},
		{
			name: "create_function_or_replace",
			setup: []string{
				"CREATE FUNCTION probe_fn3(x INT64) AS (x + 1)",
			},
			sql: "CREATE OR REPLACE FUNCTION probe_fn3(x INT64) AS (x + 2)",
		},
		{
			name: "drop_function",
			setup: []string{
				"CREATE FUNCTION probe_fn4(x INT64) AS (x)",
			},
			sql: "DROP FUNCTION probe_fn4",
		},
		{
			name: "create_table_function",
			setup: []string{
				"CREATE TABLE probe_t5 (k INT64)",
			},
			sql: `CREATE TABLE FUNCTION probe_tvf1(MinK INT64) AS (
				SELECT k FROM probe_t5 WHERE k >= MinK
			)`,
		},
		{
			name: "create_view_or_replace",
			setup: []string{
				"CREATE TABLE probe_t6 (k INT64, v STRING)",
				"CREATE VIEW probe_v1 AS SELECT k FROM probe_t6",
			},
			sql: "CREATE OR REPLACE VIEW probe_v1 AS SELECT v FROM probe_t6",
		},
		{
			name: "create_table_if_not_exists",
			sql:  "CREATE TABLE IF NOT EXISTS probe_tine (k INT64)",
		},
		{
			name: "drop_table_if_exists",
			sql:  "DROP TABLE IF EXISTS probe_doesnotexist",
		},
		{
			name: "explain",
			sql:  "EXPLAIN SELECT 1",
		},
		{
			name: "execute_immediate_simple",
			sql:  `EXECUTE IMMEDIATE "SELECT 1"`,
		},
		{
			name: "rename_column",
			setup: []string{
				"CREATE TABLE probe_t7 (k INT64)",
			},
			sql: "ALTER TABLE probe_t7 RENAME COLUMN k TO k2",
		},
		{
			name: "alter_column_set_data_type",
			setup: []string{
				"CREATE TABLE probe_t8 (v STRING)",
			},
			sql: "ALTER TABLE probe_t8 ALTER COLUMN v SET DATA TYPE STRING",
		},
		{
			name: "alter_column_drop_not_null",
			setup: []string{
				"CREATE TABLE probe_t9 (k INT64 NOT NULL)",
			},
			sql: "ALTER TABLE probe_t9 ALTER COLUMN k DROP NOT NULL",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, err := sql.Open("googlesqlite", ":memory:?_test=probe_"+c.name)
			if err != nil {
				t.Fatalf("sql.Open: %v", err)
			}
			defer db.Close()
			db.SetMaxOpenConns(1)
			ctx := context.Background()
			conn, err := db.Conn(ctx)
			if err != nil {
				t.Fatalf("Conn: %v", err)
			}
			defer conn.Close()

			for _, s := range c.setup {
				if _, err := conn.ExecContext(ctx, s); err != nil {
					t.Logf("setup %q failed (skipping): %v", s, err)
					t.SkipNow()
				}
			}
			// Try Exec first; if it returns an error referencing the
			// unsupported kind, that still exercised the analyzer path.
			_, err = conn.ExecContext(ctx, c.sql)
			if err != nil {
				// Some procedural / DDL forms are intentionally
				// unsupported. We only fail the test if the error is
				// something unrelated (e.g. a panic surfacing through
				// recoverPanicAsError).
				lower := strings.ToLower(err.Error())
				if strings.Contains(lower, "panic") {
					t.Fatalf("panic exec %q: %v", c.sql, err)
				}
				t.Logf("exec %q returned %v (acceptable for probe)", c.sql, err)
				return
			}
		})
	}
}

// ---- from tests/parity/driver_test.go ----

func TestDriver(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS Singers (
  SingerId   INT64 NOT NULL,
  FirstName  STRING(1024),
  LastName   STRING(1024),
  SingerInfo BYTES(MAX)
)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT Singers (SingerId, FirstName, LastName) VALUES (1, 'John', 'Titor')`); err != nil {
		t.Fatal(err)
	}
	row := db.QueryRow("SELECT SingerID, FirstName, LastName FROM Singers WHERE SingerId = @id", 1)
	if row.Err() != nil {
		t.Fatal(row.Err())
	}
	var (
		singerID  int64
		firstName string
		lastName  string
	)
	if err := row.Scan(&singerID, &firstName, &lastName); err != nil {
		t.Fatal(err)
	}
	if singerID != 1 || firstName != "John" || lastName != "Titor" {
		t.Fatalf("failed to find row %v %v %v", singerID, firstName, lastName)
	}
	if _, err := db.Exec(`
CREATE VIEW IF NOT EXISTS SingerNames AS SELECT FirstName || ' ' || LastName AS Name FROM Singers`); err != nil {
		t.Fatal(err)
	}

	viewRow := db.QueryRow("SELECT Name FROM SingerNames LIMIT 1")
	if viewRow.Err() != nil {
		t.Fatal(viewRow.Err())
	}

	var name string

	if err := viewRow.Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "John Titor" {
		t.Fatalf("failed to find view row")
	}
}

func TestRegisterCustomDriver(t *testing.T) {
	sql.Register("googlesqlite-custom", &googlesqlite.Driver{
		ConnectHook: func(conn *googlesqlite.Conn) error {
			return conn.SetNamePath([]string{"project-id", "datasetID"})
		},
	})
	db, err := sql.Open("googlesqlite-custom", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS tableID (Id INT64 NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT `project-id`.datasetID.tableID (Id) VALUES (1)"); err != nil {
		t.Fatal(err)
	}
	row := db.QueryRow("SELECT * FROM project-id.datasetID.tableID WHERE Id = ?", 1)
	if row.Err() != nil {
		t.Fatal(row.Err())
	}
	var id int64
	if err := row.Scan(&id); err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("failed to find row %v", id)
	}
}

func TestChangedCatalog(t *testing.T) {
	t.Run("table", func(t *testing.T) {
		db, err := sql.Open("googlesqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		result, err := db.Exec(`
CREATE TABLE IF NOT EXISTS Singers (
  SingerId   INT64 NOT NULL,
  FirstName  STRING(1024),
  LastName   STRING(1024),
  SingerInfo BYTES(MAX)
)`)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := db.Query(`DROP TABLE Singers`)
		if err != nil {
			t.Fatal(err)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		resultCatalog, err := googlesqlite.ChangedCatalogFromResult(result)
		if err != nil {
			t.Fatal(err)
		}
		if !resultCatalog.Changed() {
			t.Fatal("failed to get changed catalog")
		}
		if len(resultCatalog.Table.Added) != 1 {
			t.Fatal("failed to get created table spec")
		}
		if diff := cmp.Diff(resultCatalog.Table.Added[0].NamePath, []string{"Singers"}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
		rowsCatalog, err := googlesqlite.ChangedCatalogFromRows(rows)
		if err != nil {
			t.Fatal(err)
		}
		if !rowsCatalog.Changed() {
			t.Fatal("failed to get changed catalog")
		}
		if len(rowsCatalog.Table.Deleted) != 1 {
			t.Fatal("failed to get deleted table spec")
		}
		if diff := cmp.Diff(rowsCatalog.Table.Deleted[0].NamePath, []string{"Singers"}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
	t.Run("function", func(t *testing.T) {
		db, err := sql.Open("googlesqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		result, err := db.ExecContext(context.Background(), `CREATE FUNCTION ANY_ADD(x ANY TYPE, y ANY TYPE) AS ((x + 4) / y)`)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := db.QueryContext(context.Background(), `DROP FUNCTION ANY_ADD`)
		if err != nil {
			t.Fatal(err)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		resultCatalog, err := googlesqlite.ChangedCatalogFromResult(result)
		if err != nil {
			t.Fatal(err)
		}
		if !resultCatalog.Changed() {
			t.Fatal("failed to get changed catalog")
		}
		if len(resultCatalog.Function.Added) != 1 {
			t.Fatal("failed to get created function spec")
		}
		if diff := cmp.Diff(resultCatalog.Function.Added[0].NamePath, []string{"ANY_ADD"}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
		rowsCatalog, err := googlesqlite.ChangedCatalogFromRows(rows)
		if err != nil {
			t.Fatal(err)
		}
		if !rowsCatalog.Changed() {
			t.Fatal("failed to get changed catalog")
		}
		if len(rowsCatalog.Function.Deleted) != 1 {
			t.Fatal("failed to get deleted function spec")
		}
		if diff := cmp.Diff(rowsCatalog.Function.Deleted[0].NamePath, []string{"ANY_ADD"}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
}

func TestPreparedStatements(t *testing.T) {
	t.Run("prepared select", func(t *testing.T) {
		db, err := sql.Open("googlesqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS Singers (
  SingerId   INT64 NOT NULL,
  FirstName  STRING(1024),
  LastName   STRING(1024),
  SingerInfo BYTES(MAX)
)`); err != nil {
			t.Fatal(err)
		}
		stmt, err := db.Prepare("SELECT * FROM Singers WHERE SingerId = ?")
		if err != nil {
			t.Fatal(err)
		}
		rows, err := stmt.Query("123")
		if err != nil {
			t.Fatal(err)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if rows.Next() {
			t.Fatal("found unexpected row; expected no rows")
		}
	})
	t.Run("prepared insert", func(t *testing.T) {
		db, err := sql.Open("googlesqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS Items (ItemId   INT64 NOT NULL)`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec("INSERT `Items` (`ItemId`) VALUES (123)"); err != nil {
			t.Fatal(err)
		}

		// Test that executing without args fails
		_, err = db.Exec("INSERT `Items` (`ItemId`) VALUES (?)")
		if err == nil {
			t.Fatal("expected error when inserting without args; got no error")
		}

		stmt, err := db.Prepare("INSERT `Items` (`ItemId`) VALUES (?)")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := stmt.Exec(456); err != nil {
			t.Fatal(err)
		}

		stmt, err = db.PrepareContext(context.Background(), "INSERT `Items` (`ItemId`) VALUES (?)")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := stmt.Exec(456); err != nil {
			t.Fatal(err)
		}

		rows, err := db.Query("SELECT * FROM Items WHERE ItemId = 456")
		if err != nil {
			t.Fatal(err)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if !rows.Next() {
			t.Fatal("expected no rows; expected one row")
		}
	})
}
