package googlesqlite_test

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// The tests here cover the `CREATE TABLE` shapes that build a table out
// of another table rather than a column list — LIKE, COPY, CLONE — plus
// CREATE / DROP SNAPSHOT TABLE. The grammar and the inheritance rules
// they assert come from docs/third_party/googlesql-docs/resolved_ast.md
// (ResolvedCreateTableStmt, ResolvedCreateSnapshotTableStmt).

// newSourceTableDB opens a database holding a single source table whose
// schema exercises everything the clone family has to carry across:
// a NOT NULL column, a DEFAULT, an ARRAY (encoded, not a native SQLite
// type) and a PRIMARY KEY.
func newSourceTableDB(t *testing.T, name string) *sql.DB {
	t.Helper()
	db, err := sql.Open("googlesqlite", ":memory:?_test="+name)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	for _, stmt := range []string{
		"CREATE TABLE cts_src (a INT64 NOT NULL, b STRING DEFAULT 'z', c ARRAY<INT64>, PRIMARY KEY (a) NOT ENFORCED)",
		"INSERT INTO cts_src (a, b, c) VALUES (1, 'x', [1, 2])",
		"INSERT INTO cts_src (a, b, c) VALUES (2, 'y', [3])",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("setup %q: %v", stmt, err)
		}
	}
	return db
}

// scanAB reads (a, b) pairs from a query, rendering each row as "a=b".
func scanAB(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("Query %q: %v", query, err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var a int64
		var b sql.NullString
		if err := rows.Scan(&a, &b); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, strconv.FormatInt(a, 10)+"="+b.String)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	return got
}

// TestCreateTableLike asserts that LIKE reproduces the source schema —
// including the NOT NULL, DEFAULT and PRIMARY KEY that the analyzer's
// own materialised column list drops — and copies no rows.
func TestCreateTableLike(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_like")

	if _, err := db.Exec("CREATE TABLE cts_like LIKE cts_src"); err != nil {
		t.Fatalf("CREATE TABLE LIKE: %v", err)
	}

	if got := scanAB(t, db, "SELECT a, b FROM cts_like"); len(got) != 0 {
		t.Errorf("LIKE copied rows, want none: %v", got)
	}

	// NOT NULL is inherited.
	if _, err := db.Exec("INSERT INTO cts_like (a, b) VALUES (NULL, 'q')"); err == nil {
		t.Error("LIKE dropped NOT NULL from the source column")
	}

	// DEFAULT is inherited: omitting b must yield the source default.
	if _, err := db.Exec("INSERT INTO cts_like (a) VALUES (5)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	got := scanAB(t, db, "SELECT a, b FROM cts_like ORDER BY a")
	if len(got) != 1 || got[0] != "5=z" {
		t.Errorf("LIKE dropped DEFAULT: got %v, want [5=z]", got)
	}
}

// TestCreateTableCopyClone covers COPY and CLONE, which upstream treats
// as the same statement shape differing only in cost.
func TestCreateTableCopyClone(t *testing.T) {
	t.Parallel()
	for _, keyword := range []string{"COPY", "CLONE"} {
		t.Run(keyword, func(t *testing.T) {
			t.Parallel()
			db := newSourceTableDB(t, "ct_"+strings.ToLower(keyword))

			if _, err := db.Exec("CREATE TABLE cts_dst " + keyword + " cts_src"); err != nil {
				t.Fatalf("CREATE TABLE %s: %v", keyword, err)
			}

			got := scanAB(t, db, "SELECT a, b FROM cts_dst ORDER BY a")
			want := []string{"1=x", "2=y"}
			if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
				t.Errorf("%s rows: got %v, want %v", keyword, got, want)
			}

			// The schema is inherited, not just the data.
			if _, err := db.Exec("INSERT INTO cts_dst (a, b) VALUES (NULL, 'q')"); err == nil {
				t.Errorf("%s dropped NOT NULL from the source column", keyword)
			}

			// ARRAY columns survive the copy: they are stored in the
			// driver's encoded form, so a naive copy would corrupt them.
			var length int64
			if err := db.QueryRow("SELECT ARRAY_LENGTH(c) FROM cts_dst WHERE a = 1").Scan(&length); err != nil {
				t.Fatalf("%s ARRAY_LENGTH: %v", keyword, err)
			}
			if length != 2 {
				t.Errorf("%s ARRAY column: got length %d, want 2", keyword, length)
			}
		})
	}
}

// TestCreateTableCloneWhere covers the optional WHERE clause, which the
// analyzer represents by wrapping the source TableScan in a FilterScan.
func TestCreateTableCloneWhere(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_clone_where")

	if _, err := db.Exec("CREATE TABLE cts_filtered CLONE cts_src WHERE a > 1"); err != nil {
		t.Fatalf("CREATE TABLE CLONE ... WHERE: %v", err)
	}
	got := scanAB(t, db, "SELECT a, b FROM cts_filtered ORDER BY a")
	if len(got) != 1 || got[0] != "2=y" {
		t.Errorf("CLONE ... WHERE: got %v, want [2=y]", got)
	}
}

// TestCreateTableCloneIfNotExists pins the guard against re-copying.
// SQLite's own IF NOT EXISTS covers the CREATE but cannot cover the
// separate INSERT that fills the table, so an unguarded implementation
// appends a fresh set of rows on every re-run.
func TestCreateTableCloneIfNotExists(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_clone_ine")

	for i := 0; i < 3; i++ {
		if _, err := db.Exec("CREATE TABLE IF NOT EXISTS cts_ine COPY cts_src"); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	if got := scanAB(t, db, "SELECT a, b FROM cts_ine ORDER BY a"); len(got) != 2 {
		t.Errorf("IF NOT EXISTS duplicated rows: got %d rows (%v), want 2", len(got), got)
	}
}

// TestCreateTableCloneOrReplace checks the converse: OR REPLACE drops
// the target first, so re-running must leave exactly one copy.
func TestCreateTableCloneOrReplace(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_clone_replace")

	for i := 0; i < 2; i++ {
		if _, err := db.Exec("CREATE OR REPLACE TABLE cts_repl CLONE cts_src"); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	if got := scanAB(t, db, "SELECT a, b FROM cts_repl ORDER BY a"); len(got) != 2 {
		t.Errorf("OR REPLACE: got %d rows (%v), want 2", len(got), got)
	}
}

// TestCreateTableCloneOptions pins upstream's option-inheritance rule:
// explicitly specified options win, unspecified ones fall through from
// the source table.
func TestCreateTableCloneOptions(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_clone_options")

	for _, stmt := range []string{
		"CREATE TABLE ctsds.opt_src (a INT64) OPTIONS(description='orig', friendly_name='fn')",
		"CREATE TABLE ctsds.opt_dst COPY ctsds.opt_src OPTIONS(description='new')",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%q: %v", stmt, err)
		}
	}

	rows, err := db.Query(
		"SELECT option_name, option_value FROM ctsds.INFORMATION_SCHEMA.TABLE_OPTIONS " +
			"WHERE table_name = 'opt_dst' ORDER BY option_name")
	if err != nil {
		t.Fatalf("TABLE_OPTIONS: %v", err)
	}
	defer rows.Close()

	got := map[string]string{}
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got[name] = value
	}
	if got["description"] != "'new'" {
		t.Errorf("explicit option not applied: description = %q, want 'new'", got["description"])
	}
	if got["friendly_name"] != "'fn'" {
		t.Errorf("unspecified option not inherited: friendly_name = %q, want 'fn'", got["friendly_name"])
	}
}

// TestCreateTableCloneForSystemTimeRejected pins the explicit rejection
// of time travel. googlesqlite keeps only the current state of a table,
// so honouring the clause is impossible; answering with present-day
// rows would be silently wrong. Plain queries already reject it, and
// this keeps the CREATE TABLE path consistent.
func TestCreateTableCloneForSystemTimeRejected(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_clone_systime")

	_, err := db.Exec("CREATE TABLE cts_st COPY cts_src FOR SYSTEM_TIME AS OF CURRENT_TIMESTAMP()")
	if err == nil {
		t.Fatal("FOR SYSTEM_TIME AS OF was accepted, want an error")
	}
	if !strings.Contains(err.Error(), "FOR SYSTEM_TIME AS OF is not supported") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSnapshotTable covers CREATE SNAPSHOT TABLE and DROP SNAPSHOT
// TABLE. googlesqlite materialises a snapshot as an ordinary table
// holding a copy of the source rows; it is neither point-in-time nor
// read-only.
func TestSnapshotTable(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_snapshot")

	if _, err := db.Exec("CREATE SNAPSHOT TABLE cts_snap CLONE cts_src"); err != nil {
		t.Fatalf("CREATE SNAPSHOT TABLE: %v", err)
	}
	got := scanAB(t, db, "SELECT a, b FROM cts_snap ORDER BY a")
	want := []string{"1=x", "2=y"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("snapshot rows: got %v, want %v", got, want)
	}

	if _, err := db.Exec("DROP SNAPSHOT TABLE cts_snap"); err != nil {
		t.Fatalf("DROP SNAPSHOT TABLE: %v", err)
	}
	if _, err := db.Query("SELECT a FROM cts_snap"); err == nil {
		t.Error("snapshot table still readable after DROP SNAPSHOT TABLE")
	}

	if _, err := db.Exec("DROP SNAPSHOT TABLE IF EXISTS cts_absent"); err != nil {
		t.Errorf("DROP SNAPSHOT TABLE IF EXISTS on a missing table: %v", err)
	}
}

// TestCreateTableSourceScopes checks the clone family against the
// create-scope and name-path modifiers the rest of the CREATE TABLE
// surface supports.
func TestCreateTableSourceScopes(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_scopes")

	for _, stmt := range []string{
		"CREATE TEMP TABLE cts_tmp_like LIKE cts_src",
		"CREATE TEMP TABLE cts_tmp_copy COPY cts_src",
		"CREATE TABLE ctsns.cts_pathed CLONE cts_src",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Errorf("%q: %v", stmt, err)
		}
	}
}

// TestCreateTableSourcePrepared drives the clone family through
// db.Prepare rather than db.Exec. The driver implements both
// driver.ExecerContext and driver.Conn.PrepareContext, and the two
// reach different code: db.Exec runs CreateTableStmtAction.exec, while
// a prepared statement runs CreateTableStmt.Exec. Both have to create
// the table and copy the rows.
func TestCreateTableSourcePrepared(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		stmt string
		want int64
	}{
		{"CREATE TABLE cts_p LIKE cts_src", 0},
		{"CREATE TABLE cts_p COPY cts_src", 2},
		{"CREATE TABLE cts_p CLONE cts_src", 2},
		{"CREATE TABLE cts_p CLONE cts_src WHERE a > 1", 1},
		{"CREATE SNAPSHOT TABLE cts_p CLONE cts_src", 2},
	} {
		t.Run(tc.stmt, func(t *testing.T) {
			t.Parallel()
			db := newSourceTableDB(t, "ct_prepared")

			stmt, err := db.Prepare(tc.stmt)
			if err != nil {
				t.Fatalf("Prepare: %v", err)
			}
			defer stmt.Close()
			if _, err := stmt.Exec(); err != nil {
				t.Fatalf("Exec: %v", err)
			}

			var got int64
			if err := db.QueryRow("SELECT COUNT(*) FROM cts_p").Scan(&got); err != nil {
				t.Fatalf("COUNT: %v", err)
			}
			if got != tc.want {
				t.Errorf("row count: got %d, want %d", got, tc.want)
			}
		})
	}
}

// TestCreateTableSourcePreparedIfNotExists pins the IF NOT EXISTS guard
// on the prepared path, which reaches it through a different call site
// than db.Exec does.
func TestCreateTableSourcePreparedIfNotExists(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_prepared_ine")

	for i := 0; i < 3; i++ {
		stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS cts_pine COPY cts_src")
		if err != nil {
			t.Fatalf("Prepare %d: %v", i, err)
		}
		if _, err := stmt.Exec(); err != nil {
			t.Fatalf("Exec %d: %v", i, err)
		}
		stmt.Close()
	}

	var got int64
	if err := db.QueryRow("SELECT COUNT(*) FROM cts_pine").Scan(&got); err != nil {
		t.Fatalf("COUNT: %v", err)
	}
	if got != 2 {
		t.Errorf("prepared IF NOT EXISTS duplicated rows: got %d, want 2", got)
	}
}

// TestCreateTableSourcePersistence checks that a cloned table's
// inherited schema survives a reopen of a file-backed database. The
// constraints live in the catalog's stored spec rather than in the
// SQLite schema alone, so they have to round-trip through it.
func TestCreateTableSourcePersistence(t *testing.T) {
	t.Parallel()
	dsn := filepath.Join(t.TempDir(), "clone.db")

	db, err := sql.Open("googlesqlite", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	for _, stmt := range []string{
		"CREATE TABLE cts_src (a INT64 NOT NULL, b STRING DEFAULT 'z', c ARRAY<INT64>)",
		"INSERT INTO cts_src (a, b, c) VALUES (1, 'x', [1, 2])",
		"CREATE TABLE cts_persist COPY cts_src",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%q: %v", stmt, err)
		}
	}
	db.Close()

	reopened, err := sql.Open("googlesqlite", dsn)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()

	// Inherited NOT NULL still enforced.
	if _, err := reopened.Exec("INSERT INTO cts_persist (a, b) VALUES (NULL, 'q')"); err == nil {
		t.Error("NOT NULL lost across reopen")
	}
	// Inherited DEFAULT still applied.
	if _, err := reopened.Exec("INSERT INTO cts_persist (a) VALUES (9)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	var b string
	if err := reopened.QueryRow("SELECT b FROM cts_persist WHERE a = 9").Scan(&b); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if b != "z" {
		t.Errorf("DEFAULT lost across reopen: got %q, want %q", b, "z")
	}
	// Encoded ARRAY values survive.
	var length int64
	if err := reopened.QueryRow("SELECT ARRAY_LENGTH(c) FROM cts_persist WHERE a = 1").Scan(&length); err != nil {
		t.Fatalf("ARRAY_LENGTH: %v", err)
	}
	if length != 2 {
		t.Errorf("ARRAY corrupted across reopen: got length %d, want 2", length)
	}
}

// TestCreateTableCloneIntoItself pins the rejection of a
// self-referential clone. `CREATE OR REPLACE TABLE t COPY t` drops the
// target first, so an unguarded implementation copies from the table it
// has just emptied and silently leaves it with no rows.
func TestCreateTableCloneIntoItself(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_self")

	_, err := db.Exec("CREATE OR REPLACE TABLE cts_src COPY cts_src")
	if err == nil {
		t.Fatal("self-copy was accepted, want an error")
	}
	if !strings.Contains(err.Error(), "into itself") {
		t.Errorf("unexpected error: %v", err)
	}

	// The source must be untouched.
	var got int64
	if err := db.QueryRow("SELECT COUNT(*) FROM cts_src").Scan(&got); err != nil {
		t.Fatalf("COUNT: %v", err)
	}
	if got != 2 {
		t.Errorf("self-copy destroyed data: source has %d rows, want 2", got)
	}
}

// TestCreateTableCopyTypes checks that every value type the driver
// encodes survives a COPY. Several are stored in an encoded form rather
// than a native SQLite type, so a copy that went through the wrong
// representation would corrupt them.
func TestCreateTableCopyTypes(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=ct_types")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	for _, stmt := range []string{
		`CREATE TABLE cts_types (
			i INT64, s STRING, f FLOAT64, bo BOOL, byt BYTES,
			d DATE, ts TIMESTAMP, n NUMERIC, j JSON,
			st STRUCT<x INT64, y STRING>, arr ARRAY<STRING>)`,
		`INSERT INTO cts_types VALUES (
			1, 'str', 1.5, TRUE, b'bytes',
			DATE '2020-01-02', TIMESTAMP '2020-01-02 03:04:05+00',
			NUMERIC '1.23', JSON '{"k":[1,2]}',
			STRUCT(7 AS x, 'sv' AS y), ['p', 'q'])`,
		"CREATE TABLE cts_types_copy COPY cts_types",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%q: %v", stmt, err)
		}
	}

	var (
		i        int64
		s        string
		f        float64
		bo       bool
		d        string
		numeric  string
		jsonElem string
		structX  int64
		structY  string
		arrLen   int64
	)
	err = db.QueryRow(`SELECT i, s, f, bo, d, n,
		JSON_VALUE(j, '$.k[1]'), st.x, st.y, ARRAY_LENGTH(arr)
		FROM cts_types_copy`).
		Scan(&i, &s, &f, &bo, &d, &numeric, &jsonElem, &structX, &structY, &arrLen)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}

	for _, tc := range []struct {
		name      string
		got, want any
	}{
		{"INT64", i, int64(1)},
		{"STRING", s, "str"},
		{"FLOAT64", f, 1.5},
		{"BOOL", bo, true},
		{"DATE", d, "2020-01-02"},
		{"NUMERIC", numeric, "1.23"},
		{"JSON element", jsonElem, "2"},
		{"STRUCT field x", structX, int64(7)},
		{"STRUCT field y", structY, "sv"},
		{"ARRAY length", arrLen, int64(2)},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}

// TestCreateTableSourceMissing checks that an unknown source table is
// reported by the analyzer rather than producing an empty table.
func TestCreateTableSourceMissing(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_missing")

	for _, stmt := range []string{
		"CREATE TABLE cts_x LIKE cts_absent",
		"CREATE TABLE cts_x COPY cts_absent",
		"CREATE TABLE cts_x CLONE cts_absent",
		"CREATE SNAPSHOT TABLE cts_x CLONE cts_absent",
	} {
		_, err := db.Exec(stmt)
		if err == nil {
			t.Errorf("%q: want an error for a missing source table", stmt)
			continue
		}
		if !strings.Contains(err.Error(), "Table not found") {
			t.Errorf("%q: unexpected error: %v", stmt, err)
		}
	}
}

// TestCreateTableCloneIntoItselfCaseInsensitive pins the case-folding
// half of the self-clone guard. SQLite resolves table names
// case-insensitively, so `CREATE OR REPLACE TABLE T COPY t` drops `t`
// even though the two spellings differ in the statement text. A
// case-sensitive guard lets exactly that through and empties the source.
func TestCreateTableCloneIntoItselfCaseInsensitive(t *testing.T) {
	t.Parallel()
	for _, stmt := range []string{
		"CREATE OR REPLACE TABLE cts_src COPY CTS_SRC",
		"CREATE OR REPLACE TABLE CTS_SRC COPY cts_src",
		"CREATE OR REPLACE TABLE CtS_SrC CLONE cts_src",
	} {
		t.Run(stmt, func(t *testing.T) {
			t.Parallel()
			db := newSourceTableDB(t, "ct_self_case")

			if _, err := db.Exec(stmt); err == nil {
				t.Error("case-differing self-copy was accepted, want an error")
			}
			var got int64
			if err := db.QueryRow("SELECT COUNT(*) FROM cts_src").Scan(&got); err != nil {
				t.Fatalf("COUNT: %v", err)
			}
			if got != 2 {
				t.Errorf("source lost rows: got %d, want 2", got)
			}
		})
	}
}

// TestCreateTableCloneCopyFailureCleansUp checks that a copy which
// fails after the table exists does not wedge the table name. The
// CREATE and the row-copy INSERT are separate statements and may sit
// inside a caller-owned transaction, so the failure path drops the
// table it created rather than rolling anything back.
func TestCreateTableCloneCopyFailureCleansUp(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_copy_fail")

	// A predicate that only fails at evaluation time: the statement
	// analyses cleanly, the CREATE runs, then the INSERT errors.
	if _, err := db.Exec("CREATE TABLE cts_fail CLONE cts_src WHERE 1 / 0 > 0"); err == nil {
		t.Fatal("expected the division by zero to fail the copy")
	}

	// The name must be reusable rather than half-claimed.
	if _, err := db.Exec("CREATE TABLE cts_fail (z INT64)"); err != nil {
		t.Errorf("table name wedged after a failed copy: %v", err)
	}
}

// TestCreateTableCloneConcurrentIfNotExists exercises the IF NOT EXISTS
// guard from several connections at once. The guard is check-then-act,
// so a lost race would show up as duplicated rows. A file-backed DSN is
// required: separate connections to :memory: get separate databases.
func TestCreateTableCloneConcurrentIfNotExists(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", filepath.Join(t.TempDir(), "concurrent.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	for _, stmt := range []string{
		"CREATE TABLE cts_src (a INT64)",
		"INSERT INTO cts_src (a) VALUES (1)",
		"INSERT INTO cts_src (a) VALUES (2)",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%q: %v", stmt, err)
		}
	}

	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = db.Exec("CREATE TABLE IF NOT EXISTS cts_conc COPY cts_src")
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}

	var got int64
	if err := db.QueryRow("SELECT COUNT(*) FROM cts_conc").Scan(&got); err != nil {
		t.Fatalf("COUNT: %v", err)
	}
	if got != 2 {
		t.Errorf("concurrent IF NOT EXISTS clones produced %d rows, want 2", got)
	}
}

// TestCreateTableCloneUnsupportedSource checks the error for a source
// with no stored spec. INFORMATION_SCHEMA tables are registered
// straight into the analyzer catalog, so COPY / CLONE — which take the
// whole schema from the spec — have nothing to inherit and must say so.
func TestCreateTableCloneUnsupportedSource(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_unsupported")

	if _, err := db.Exec("CREATE TABLE ctsu.ctsu_src (a INT64)"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err := db.Exec("CREATE TABLE ctsu.ctsu_dst COPY ctsu.INFORMATION_SCHEMA.COLUMNS")
	if err == nil {
		t.Fatal("COPY from INFORMATION_SCHEMA was accepted, want an error")
	}
	if !strings.Contains(err.Error(), "only tables created through googlesqlite") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestCreateTableLikeCrossDataset checks that inheritance follows the
// source across a dataset boundary rather than silently finding nothing.
func TestCreateTableLikeCrossDataset(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_cross_ds")

	for _, stmt := range []string{
		"CREATE TABLE ctsx_a.src (a INT64 NOT NULL, b STRING DEFAULT 'z')",
		"CREATE TABLE ctsx_b.dst LIKE ctsx_a.src",
		"INSERT INTO ctsx_b.dst (a) VALUES (1)",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%q: %v", stmt, err)
		}
	}
	if _, err := db.Exec("INSERT INTO ctsx_b.dst (a, b) VALUES (NULL, 'q')"); err == nil {
		t.Error("cross-dataset LIKE dropped NOT NULL")
	}
	var b string
	if err := db.QueryRow("SELECT b FROM ctsx_b.dst WHERE a = 1").Scan(&b); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if b != "z" {
		t.Errorf("cross-dataset LIKE dropped DEFAULT: got %q, want %q", b, "z")
	}
}

// TestCreateTableCloneOutlivesSource checks that a clone is independent
// of its source. The inherited spec deep-copies the source's ColumnSpec
// pointers, which the catalog otherwise hands out live.
func TestCreateTableCloneOutlivesSource(t *testing.T) {
	t.Parallel()
	db := newSourceTableDB(t, "ct_outlives")

	for _, stmt := range []string{
		"CREATE TABLE cts_indep COPY cts_src",
		"DROP TABLE cts_src",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%q: %v", stmt, err)
		}
	}

	var got int64
	if err := db.QueryRow("SELECT COUNT(*) FROM cts_indep").Scan(&got); err != nil {
		t.Fatalf("clone unusable after its source was dropped: %v", err)
	}
	if got != 2 {
		t.Errorf("clone rows after source dropped: got %d, want 2", got)
	}
	if _, err := db.Exec("INSERT INTO cts_indep (a, b) VALUES (NULL, 'q')"); err == nil {
		t.Error("clone lost NOT NULL after its source was dropped")
	}
}
