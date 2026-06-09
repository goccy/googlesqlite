package googlesqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
)

// ---- from tests/parity/regression_test.go ----

// TestRegression_DefaultTimezoneIsUTC asserts that naive TIMESTAMP literals
// (no offset) are parsed in UTC, matching BigQuery's documented default.
//
// Origin: upstream the upstream analyzer defaulted to
// America/Los_Angeles, so EXTRACT(HOUR FROM TIMESTAMP '2024-01-01 12:00:00')
// returned 20 instead of 12. This regression test pins the fix.
func TestRegression_DefaultTimezoneIsUTC(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	row := db.QueryRowContext(ctx, "SELECT EXTRACT(HOUR FROM TIMESTAMP '2024-01-01 12:00:00')")
	var got int64
	if err := row.Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 12 {
		t.Fatalf("default TZ leaks: EXTRACT(HOUR) returned %d, want 12 (UTC)", got)
	}
}

// TestRegression_IntegerTypeAlias asserts that INTEGER and INT are accepted
// as aliases for INT64 in DDL.
//
// Regression: BigQuery documents INTEGER and INT as aliases for INT64.
func TestRegression_IntegerTypeAlias(t *testing.T) {
	t.Parallel()
	for _, alias := range []string{"INTEGER", "INT", "SMALLINT", "BIGINT", "TINYINT"} {
		t.Run(alias, func(t *testing.T) {
			db, err := sql.Open("googlesqlite", "file:alias_"+alias+"?mode=memory&cache=private")
			if err != nil {
				t.Fatal(err)
			}
			db.SetMaxOpenConns(1)
			defer db.Close()
			ctx := context.Background()
			if _, err := db.ExecContext(ctx, "CREATE TABLE t (id "+alias+")"); err != nil {
				t.Fatalf("CREATE TABLE with %s alias rejected: %v", alias, err)
			}
		})
	}
}

// =====================================================================
// F5 / F6 / F7 / F8 / F9 / F13 triage tests
// =====================================================================

// =====================================================================
// Group A — newly added functions (post-F1-F13 design questions, A1-A5)
// =====================================================================

// TestA1_ContainsSubstr
func TestA1_ContainsSubstr(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	for _, tc := range []struct {
		query string
		want  bool
	}{
		{"SELECT CONTAINS_SUBSTR('hello world', 'world')", true},
		{"SELECT CONTAINS_SUBSTR('Hello World', 'world')", true}, // case-insensitive
		{"SELECT CONTAINS_SUBSTR('Hello World', 'WORLD')", true}, // case-insensitive
		{"SELECT CONTAINS_SUBSTR('hello world', 'goodbye')", false},
		{"SELECT CONTAINS_SUBSTR('hello', '')", true}, // empty needle
	} {
		var got bool
		if err := db.QueryRowContext(context.Background(), tc.query).Scan(&got); err != nil {
			t.Errorf("%s: %v", tc.query, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %v want %v", tc.query, got, tc.want)
		}
	}
	// NULL operand returns NULL.
	var n sql.NullBool
	if err := db.QueryRowContext(context.Background(),
		`SELECT CONTAINS_SUBSTR(CAST(NULL AS STRING), 'foo')`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n.Valid {
		t.Errorf("CONTAINS_SUBSTR(NULL, 'foo') should be NULL, got %v", n.Bool)
	}
}

// TestA2_EditDistance — zlite#212
func TestA2_EditDistance(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	for _, tc := range []struct {
		query string
		want  int64
	}{
		{`SELECT EDIT_DISTANCE('kitten', 'sitting')`, 3},
		{`SELECT EDIT_DISTANCE('flaw', 'lawn')`, 2},
		{`SELECT EDIT_DISTANCE('abc', 'abc')`, 0},
		{`SELECT EDIT_DISTANCE('', 'abc')`, 3},
		{`SELECT EDIT_DISTANCE('abc', '')`, 3},
	} {
		var got int64
		if err := db.QueryRowContext(context.Background(), tc.query).Scan(&got); err != nil {
			t.Errorf("%s: %v", tc.query, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %d want %d", tc.query, got, tc.want)
		}
	}
}

// TestA3_MaxByMinBy — bqe#388 / bqe#356
func TestA3_MaxByMinBy(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	row := db.QueryRowContext(context.Background(), `
SELECT MAX_BY(name, score), MIN_BY(name, score)
FROM UNNEST([
  STRUCT('alice' AS name, 10 AS score),
  STRUCT('bob',   20),
  STRUCT('carol', 15)
])
`)
	var maxName, minName string
	if err := row.Scan(&maxName, &minName); err != nil {
		t.Fatal(err)
	}
	if maxName != "bob" {
		t.Errorf("MAX_BY got %q want bob", maxName)
	}
	if minName != "alice" {
		t.Errorf("MIN_BY got %q want alice", minName)
	}
}

// TestA4_Lax — bqe#243
func TestA4_Lax(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	for _, tc := range []struct {
		query string
		check func(t *testing.T, row *sql.Row)
	}{
		{
			`SELECT LAX_INT64(JSON '42')`,
			func(t *testing.T, row *sql.Row) {
				var v int64
				if err := row.Scan(&v); err != nil {
					t.Fatal(err)
				}
				if v != 42 {
					t.Errorf("got %d want 42", v)
				}
			},
		},
		{
			`SELECT LAX_FLOAT64(JSON '3.5')`,
			func(t *testing.T, row *sql.Row) {
				var v float64
				if err := row.Scan(&v); err != nil {
					t.Fatal(err)
				}
				if v != 3.5 {
					t.Errorf("got %g want 3.5", v)
				}
			},
		},
		{
			`SELECT LAX_BOOL(JSON 'true')`,
			func(t *testing.T, row *sql.Row) {
				var v bool
				if err := row.Scan(&v); err != nil {
					t.Fatal(err)
				}
				if !v {
					t.Errorf("got %v want true", v)
				}
			},
		},
		{
			`SELECT LAX_STRING(JSON '"hello"')`,
			func(t *testing.T, row *sql.Row) {
				var v string
				if err := row.Scan(&v); err != nil {
					t.Fatal(err)
				}
				if v != "hello" {
					t.Errorf("got %q want hello", v)
				}
			},
		},
		// Forgiving: incompatible value returns NULL, not an error.
		{
			`SELECT LAX_INT64(JSON '"not a number"')`,
			func(t *testing.T, row *sql.Row) {
				var v sql.NullInt64
				if err := row.Scan(&v); err != nil {
					t.Fatal(err)
				}
				if v.Valid {
					t.Errorf("LAX_INT64 of non-numeric JSON should be NULL, got %d", v.Int64)
				}
			},
		},
	} {
		t.Run(tc.query, func(t *testing.T) {
			tc.check(t, db.QueryRowContext(context.Background(), tc.query))
		})
	}
}

// TestA5_LogicalOrAndWindow — zlite#200
func TestA5_LogicalOrAndWindow(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
SELECT
  v,
  LOGICAL_OR(v)  OVER (ORDER BY i) AS or_so_far,
  LOGICAL_AND(v) OVER (ORDER BY i) AS and_so_far
FROM UNNEST([
  STRUCT(1 AS i, FALSE AS v),
  STRUCT(2,      TRUE),
  STRUCT(3,      TRUE)
])
ORDER BY i
`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	type row struct{ v, or_, and_ bool }
	got := []row{}
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.v, &r.or_, &r.and_); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	want := []row{
		{false, false, false},
		{true, true, false},
		{true, true, false},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("row %d: got %+v want %+v", i, got[i], want[i])
		}
	}
}

// =====================================================================
// Group W — driver-side wiring after design-questions audit
// (C3 EXPORT DATA, G2 RANGE, G3 TVF, G4 WITH RECURSIVE, G1 ST_*)
// =====================================================================

// TestW_ExportData asserts the driver treats
// `EXPORT DATA OPTIONS(...) AS <query>` as a write: the inner query's
// rows are materialized to the OPTIONS URI through the registered
// URIWriter for the scheme, and no rows are returned to the caller. The
// actual bytes landing at a gs:// destination (against fake-gcs-server)
// are covered by TestExportDataStatementGCS; this test runs the same
// shape through a per-test scoped URIWriter so the assertion stays at the
// driver layer with no process-wide state.
func TestW_ExportData(t *testing.T) {
	scheme, capture := registerMemScheme(t)
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), fmt.Sprintf(`
EXPORT DATA OPTIONS(uri='%s://regression/wexport.csv', format='CSV') AS
SELECT 1 AS id, 'alice' AS name
UNION ALL SELECT 2, 'bob'`, scheme))
	if err != nil {
		t.Fatalf("EXPORT DATA failed: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("EXPORT DATA returned a row (%d, %q); want none", id, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	got, ok := capture.get("regression/wexport.csv")
	if !ok {
		t.Fatalf("URI writer did not capture the export (keys: %v)", capture.keys())
	}
	if want := "id,name\n1,alice\n2,bob\n"; string(got) != want {
		t.Errorf("captured body = %q; want %q", got, want)
	}
}

// TestW_GeographyExtra covers the G1 ST_* additions.
func TestW_GeographyExtra(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	for _, tc := range []struct {
		query string
		check func(t *testing.T, row *sql.Row)
	}{
		{
			`SELECT ST_X(ST_GEOGFROMTEXT('POINT(1.5 2.5)'))`,
			func(t *testing.T, row *sql.Row) {
				var got float64
				if err := row.Scan(&got); err != nil {
					t.Fatal(err)
				}
				if got != 1.5 {
					t.Errorf("got %g want 1.5", got)
				}
			},
		},
		{
			`SELECT ST_Y(ST_GEOGFROMTEXT('POINT(1.5 2.5)'))`,
			func(t *testing.T, row *sql.Row) {
				var got float64
				if err := row.Scan(&got); err != nil {
					t.Fatal(err)
				}
				if got != 2.5 {
					t.Errorf("got %g want 2.5", got)
				}
			},
		},
		{
			`SELECT ST_ASTEXT(ST_GEOGFROMTEXT('POINT(1 2)'))`,
			func(t *testing.T, row *sql.Row) {
				var got string
				if err := row.Scan(&got); err != nil {
					t.Fatal(err)
				}
				if got != "POINT (1 2)" {
					t.Errorf("got %q want %q", got, "POINT (1 2)")
				}
			},
		},
		{
			`SELECT ST_EQUALS(ST_GEOGFROMTEXT('POINT(1 2)'), ST_GEOGFROMTEXT('POINT(1 2)'))`,
			func(t *testing.T, row *sql.Row) {
				var got bool
				if err := row.Scan(&got); err != nil {
					t.Fatal(err)
				}
				if !got {
					t.Errorf("expected ST_EQUALS to be TRUE for identical points")
				}
			},
		},
		{
			`SELECT ST_INTERSECTS(ST_GEOGFROMTEXT('POINT(0 0)'), ST_GEOGFROMTEXT('POINT(0 0)'))`,
			func(t *testing.T, row *sql.Row) {
				var got bool
				if err := row.Scan(&got); err != nil {
					t.Fatal(err)
				}
				if !got {
					t.Errorf("expected ST_INTERSECTS at same point to be TRUE")
				}
			},
		},
		{
			`SELECT ST_DWITHIN(ST_GEOGFROMTEXT('POINT(0 0)'), ST_GEOGFROMTEXT('POINT(0.001 0.001)'), 200)`,
			func(t *testing.T, row *sql.Row) {
				var got bool
				if err := row.Scan(&got); err != nil {
					t.Fatal(err)
				}
				// Two points 0.001° apart at the equator are ~157m apart, within 200m.
				if !got {
					t.Errorf("expected ST_DWITHIN within 200m to be TRUE")
				}
			},
		},
	} {
		t.Run(tc.query, func(t *testing.T) {
			tc.check(t, db.QueryRowContext(ctx, tc.query))
		})
	}
}

// TestW_TVF_DDL covers the first leg of G3 driver-side wiring: the
// CREATE TABLE FUNCTION statement is accepted as a DDL no-op. The
// call site (ResolvedTvfscan) is the follow-up — see TestW_TVF_Call.
func TestW_TVF_DDL(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(),
		`CREATE TABLE FUNCTION my_tvf(n INT64) AS (SELECT n AS x)`); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION rejected: %v", err)
	}
}

// TestW_TVF_Call exercises the call-site path: a CREATE TABLE
// FUNCTION definition is registered, then `SELECT * FROM tvf(...)`
// inlines the body with arg substitution. Covers single-arg, multi-
// arg, and multi-column-output variations.
func TestW_TVF_Call(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`CREATE TABLE FUNCTION add_one(x INT64) AS (SELECT x + 1 AS y)`); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION add_one: %v", err)
	}
	row := db.QueryRowContext(ctx, `SELECT y FROM add_one(41)`)
	var single int64
	if err := row.Scan(&single); err != nil {
		t.Fatalf("scan add_one: %v", err)
	}
	if single != 42 {
		t.Errorf("add_one result got %d want 42", single)
	}

	if _, err := db.ExecContext(ctx,
		`CREATE TABLE FUNCTION pair(a INT64, b INT64) AS (SELECT a AS lo, b AS hi)`); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION pair: %v", err)
	}
	row = db.QueryRowContext(ctx, `SELECT lo, hi FROM pair(3, 7)`)
	var lo, hi int64
	if err := row.Scan(&lo, &hi); err != nil {
		t.Fatalf("scan pair: %v", err)
	}
	if lo != 3 || hi != 7 {
		t.Errorf("pair result got (%d,%d) want (3,7)", lo, hi)
	}
}

// TestW_DeclareSetScript covers C3: DECLARE / SET produce
// parse-level statements that the upstream analyzer rejects as
// "Statement not supported". The driver intercepts them in a SQL
// pre-rewrite pass — DECLARE / SET update a per-Conn variable map
// and are removed from the SQL handed to the analyzer; bare
// references to declared variables in subsequent statements are
// substituted with the current expression.
//
// Origin: bqe#344.
func TestW_DeclareSetScript(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// One statement at a time on a single connection.
	if _, err := conn.ExecContext(ctx, `DECLARE x INT64 DEFAULT 10`); err != nil {
		t.Fatalf("DECLARE x: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `DECLARE name STRING DEFAULT 'alice'`); err != nil {
		t.Fatalf("DECLARE name: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `SET x = 42`); err != nil {
		t.Fatalf("SET x: %v", err)
	}
	row := conn.QueryRowContext(ctx, `SELECT x, name`)
	var n int64
	var s string
	if err := row.Scan(&n, &s); err != nil {
		t.Fatalf("SELECT x, name: %v", err)
	}
	if n != 42 || s != "alice" {
		t.Errorf("got x=%d name=%q want 42 / alice", n, s)
	}

	// Multi-statement script in a single ExecContext.
	row = conn.QueryRowContext(ctx,
		`DECLARE m INT64 DEFAULT 7; SET m = m + 100; SELECT m`)
	var m int64
	if err := row.Scan(&m); err != nil {
		t.Fatalf("multi-stmt script: %v", err)
	}
	if m != 107 {
		t.Errorf("got m=%d want 107", m)
	}
}

// TestW_PartitionTimePseudoColumn covers C5: ingestion-time
// partitioned tables expose `_PARTITIONTIME` and `_PARTITIONDATE`
// pseudo-columns. The analyzer rejected the syntax outright before
// the fix; the SQL pre-rewriter now strips PARTITION BY clauses
// (metadata-only for googlesqlite) and substitutes the pseudo-column
// references with CURRENT_TIMESTAMP() / CURRENT_DATE() at every
// other call site.
//
// Origin: bqe#317.
func TestW_PartitionTimePseudoColumn(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	defer func() { _, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS events_part`) }()

	if _, err := db.ExecContext(ctx,
		`CREATE TABLE events_part (id INT64) PARTITION BY DATE(_PARTITIONTIME)`); err != nil {
		t.Fatalf("CREATE TABLE with PARTITION BY DATE(_PARTITIONTIME): %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO events_part (id) VALUES (1), (2), (3)`); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	row := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events_part WHERE _PARTITIONDATE = CURRENT_DATE()`)
	var n int64
	if err := row.Scan(&n); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if n != 3 {
		t.Errorf("got %d rows want 3 (every row's _PARTITIONDATE should be today)", n)
	}
}

// TestW_UpdateStructFieldRewrite covers C2: `UPDATE t SET col.field
// = value` previously emitted invalid SQL because SQL forbids
// assigning to a function call (the formatter rendered the LHS as
// `googlesqlite_get_struct_field(col, 0)`). Fix rewrites the
// assignment into a whole-column replacement that runs the new
// `googlesqlite_struct_with_field_set` UDF, which preserves all
// other fields of the struct.
//
// Origin: bqe#299, bqe#128.
func TestW_UpdateStructFieldRewrite(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	defer func() { _, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS people`) }()

	if _, err := db.ExecContext(ctx,
		`CREATE TABLE people (id INT64, info STRUCT<name STRING, age INT64>)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO people VALUES (1, STRUCT('alice' AS name, 30 AS age)), (2, STRUCT('bob' AS name, 40 AS age))`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE people SET info.name = 'ALICE' WHERE id = 1`); err != nil {
		t.Fatalf("UPDATE struct field: %v", err)
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, info.name, info.age FROM people ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	type row struct {
		id   int64
		name string
		age  int64
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.name, &r.age); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	want := []row{
		{1, "ALICE", 30}, // name updated, age preserved
		{2, "bob", 40},   // untouched
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row[%d]: got %+v want %+v", i, got[i], want[i])
		}
	}
}

// TestW_DefaultClauseSubstitution covers C1: a column declared with
// DEFAULT must receive the default value at INSERT time when the
// column is omitted from the INSERT column list. Pre-fix the row
// stored NULL; the fix splices the default expression into each
// row's VALUES tuple at format time.
//
// Origin: bqe#211, bqe#73.
func TestW_DefaultClauseSubstitution(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	defer func() { _, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS t_default`) }()
	if _, err := db.ExecContext(ctx,
		`CREATE TABLE t_default (id INT64, status STRING DEFAULT 'pending', count INT64 DEFAULT 0)`); err != nil {
		t.Fatal(err)
	}

	// Omit both DEFAULT columns — both should get their defaults.
	if _, err := db.ExecContext(ctx, `INSERT INTO t_default (id) VALUES (1)`); err != nil {
		t.Fatalf("INSERT row 1: %v", err)
	}
	// Override one default explicitly.
	if _, err := db.ExecContext(ctx, `INSERT INTO t_default (id, status) VALUES (2, 'active')`); err != nil {
		t.Fatalf("INSERT row 2: %v", err)
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, status, count FROM t_default ORDER BY id`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	type row struct {
		id     int64
		status string
		count  int64
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.status, &r.count); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	want := []row{
		{1, "pending", 0},
		{2, "active", 0},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row[%d]: got %+v want %+v", i, got[i], want[i])
		}
	}
}

// TestW_GeographyNonPoint covers GEOGRAPHY values beyond Point: the
// driver round-trips LINESTRING / POLYGON / MULTIPOINT /
// MULTILINESTRING / MULTIPOLYGON / GEOMETRYCOLLECTION through
// ST_GEOGFROMTEXT / ST_AS_TEXT and treats them as first-class
// values (storage, retrieval, equality). Spherical computations
// (ST_DISTANCE on non-Point pairs) intentionally surface a clear
// "POINT only" error rather than returning wrong numbers.
//
// Origin: B1 driver-side partial in upstream-asks.md.
func TestW_GeographyNonPoint(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	cases := []struct {
		name string
		wkt  string
	}{
		{"linestring", "LINESTRING (0 0, 1 1, 2 2)"},
		{"polygon", "POLYGON ((0 0, 4 0, 4 4, 0 4, 0 0), (1 1, 2 1, 2 2, 1 2, 1 1))"},
		{"multipoint", "MULTIPOINT (0 0, 1 1, 2 2)"},
		{"multilinestring", "MULTILINESTRING ((0 0, 1 1), (2 2, 3 3))"},
		{"multipolygon", "MULTIPOLYGON (((0 0, 1 0, 1 1, 0 1, 0 0)), ((2 2, 3 2, 3 3, 2 3, 2 2)))"},
	}
	for _, tc := range cases {
		t.Run("round-trip "+tc.name, func(t *testing.T) {
			row := db.QueryRowContext(ctx,
				`SELECT ST_ASTEXT(ST_GEOGFROMTEXT(?))`, tc.wkt)
			var got string
			if err := row.Scan(&got); err != nil {
				t.Fatalf("scan: %v", err)
			}
			if got != tc.wkt {
				t.Errorf("got %q want %q", got, tc.wkt)
			}
		})
	}

	// Equality between two equivalent LINESTRINGs.
	t.Run("equality linestring", func(t *testing.T) {
		row := db.QueryRowContext(ctx,
			`SELECT ST_EQUALS(ST_GEOGFROMTEXT('LINESTRING (0 0, 1 1)'), ST_GEOGFROMTEXT('LINESTRING (0 0, 1 1)'))`)
		var got bool
		if err := row.Scan(&got); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if !got {
			t.Errorf("identical LINESTRINGs should be equal")
		}
	})
}

// TestW_StructInSubqueryAnonymousFields covers the documented invariant:
// when a named STRUCT is compared against a subquery yielding
// anonymous-field STRUCTs, the IN expression returned the wrong
// answer because the runtime CastValue / IN comparison path was
// keying on field names. The fix has two parts: (a) CastValue
// falls back to a positional zip when the source struct has no
// matching named keys; (b) STRUCT-typed IN expressions are
// emitted with `COLLATE googlesqlite_collate` so SQLite routes
// equality through the value-aware collation that already
// understands StructValue positionally.
func TestW_StructInSubqueryAnonymousFields(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	cases := []struct {
		name string
		sql  string
		want bool
	}{
		{
			"named-IN-anonymous-multi-row-match",
			`SELECT STRUCT(1 AS x, 'a' AS y) IN (SELECT (1, 'a') UNION ALL SELECT (2, 'b'))`,
			true,
		},
		{
			"named-IN-anonymous-multi-row-no-match",
			`SELECT STRUCT(7 AS x, 'z' AS y) IN (SELECT (1, 'a') UNION ALL SELECT (2, 'b'))`,
			false,
		},
		{
			"named-IN-named-multi-row-match",
			`SELECT STRUCT(1 AS x, 'a' AS y) IN (SELECT STRUCT(1 AS x, 'a' AS y) UNION ALL SELECT STRUCT(2 AS x, 'b' AS y))`,
			true,
		},
		{
			"anonymous-IN-anonymous-multi-row-match",
			`SELECT (1, 'a') IN (SELECT (1, 'a') UNION ALL SELECT (2, 'b'))`,
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got bool
			if err := db.QueryRowContext(ctx, tc.sql).Scan(&got); err != nil {
				t.Fatalf("scan: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

// TestW_InformationSchema verifies the INFORMATION_SCHEMA virtual
// catalog: tables / columns / schemata are auto-derived from the
// live catalog state, file-backed catalogs survive a process
// restart, and the WHERE __schema = '<dataset>' filter pushdown
// reaches the vtab cleanly.
//
// Origin: bqe#48.
func TestW_InformationSchema(t *testing.T) {
	t.Parallel()
	tmp, err := os.CreateTemp("", "info-*.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	// First connection: populate the catalog. We register two
	// tables under <project>.<dataset>; INFORMATION_SCHEMA must
	// pick them up.
	func() {
		db, err := sql.Open("googlesqlite", tmp.Name())
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		ctx := context.Background()
		for _, q := range []string{
			"CREATE TABLE `proj.ds.users` (id INT64, name STRING)",
			"CREATE TABLE `proj.ds.orders` (oid INT64, user_id INT64)",
		} {
			if _, err := db.ExecContext(ctx, q); err != nil {
				t.Fatalf("setup %q: %v", q, err)
			}
		}
	}()

	// Reopen — INFORMATION_SCHEMA must reflect the persisted
	// catalog. This is the persistence path the user explicitly
	// asked about during plan review.
	db, err := sql.Open("googlesqlite", tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	// TABLES view, ordered by name so the assertion is stable.
	rows, err := db.QueryContext(ctx,
		`SELECT table_name, table_type FROM ds.INFORMATION_SCHEMA.TABLES ORDER BY table_name`)
	if err != nil {
		t.Fatalf("TABLES query: %v", err)
	}
	type tableRow struct{ name, typ string }
	var got []tableRow
	for rows.Next() {
		var r tableRow
		if err := rows.Scan(&r.name, &r.typ); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	rows.Close()
	want := []tableRow{
		{"orders", "BASE TABLE"},
		{"users", "BASE TABLE"},
	}
	if len(got) != len(want) {
		t.Fatalf("TABLES rows: got %v want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("TABLES[%d]: got %v want %v", i, got[i], w)
		}
	}

	// COLUMNS for the users table — verify column ordering and types.
	rows, err = db.QueryContext(ctx,
		`SELECT column_name, data_type, ordinal_position
           FROM ds.INFORMATION_SCHEMA.COLUMNS
          WHERE table_name = 'users'
          ORDER BY ordinal_position`)
	if err != nil {
		t.Fatalf("COLUMNS query: %v", err)
	}
	type colRow struct {
		name string
		typ  string
		ord  int64
	}
	var cols []colRow
	for rows.Next() {
		var c colRow
		if err := rows.Scan(&c.name, &c.typ, &c.ord); err != nil {
			t.Fatal(err)
		}
		cols = append(cols, c)
	}
	rows.Close()
	if len(cols) != 2 {
		t.Fatalf("COLUMNS got %d rows want 2: %v", len(cols), cols)
	}
	if cols[0] != (colRow{"id", "INT64", 1}) {
		t.Errorf("COLUMNS[0]: got %v want {id INT64 1}", cols[0])
	}
	if cols[1] != (colRow{"name", "STRING", 2}) {
		t.Errorf("COLUMNS[1]: got %v want {name STRING 2}", cols[1])
	}

	// SCHEMATA: just `ds` should be present (we registered tables
	// only under proj.ds).
	row := db.QueryRowContext(ctx,
		`SELECT schema_name FROM ds.INFORMATION_SCHEMA.SCHEMATA`)
	var schema string
	if err := row.Scan(&schema); err != nil {
		t.Fatalf("SCHEMATA scan: %v", err)
	}
	if schema != "ds" {
		t.Errorf("SCHEMATA: got %q want ds", schema)
	}
}

// TestW_CteMaterializeMultiRef exercises a multi-reference CTE.
// F12 makes the formatter add the SQLite `MATERIALIZED` hint to
// CTEs referenced more than once so the body runs once instead of
// once per reference. This test verifies correctness across both
// the multi-ref and single-ref paths plus the SetMaterializeCTE
// toggle (no observable behavioural difference; the toggle gates
// the hint emission).
func TestW_CteMaterializeMultiRef(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	defer func() { _, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS events`) }()

	if _, err := db.ExecContext(ctx, `CREATE TABLE events (uid INT64, kind STRING)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO events VALUES (1, 'a'), (1, 'b'), (2, 'a'), (3, 'c')`); err != nil {
		t.Fatal(err)
	}

	// Multi-ref CTE: two scalar subqueries each reference per_user.
	multiRef := `WITH per_user AS (
    SELECT uid, COUNTIF(kind = 'a') AS a, COUNTIF(kind = 'b') AS b
      FROM events GROUP BY uid
)
SELECT (SELECT COUNT(*) FROM per_user WHERE a > 0) AS users_with_a,
       (SELECT COUNT(*) FROM per_user WHERE b > 0) AS users_with_b`
	row := db.QueryRowContext(ctx, multiRef)
	var a, b int64
	if err := row.Scan(&a, &b); err != nil {
		t.Fatalf("multi-ref scan: %v", err)
	}
	if a != 2 || b != 1 {
		t.Errorf("multi-ref got a=%d b=%d want 2/1", a, b)
	}

	// Single-ref CTE: hint should NOT be added but result identical.
	row = db.QueryRowContext(ctx,
		`WITH per_user AS (SELECT uid FROM events) SELECT COUNT(*) FROM per_user`)
	var n int64
	if err := row.Scan(&n); err != nil {
		t.Fatalf("single-ref scan: %v", err)
	}
	if n != 4 {
		t.Errorf("single-ref got %d want 4", n)
	}

	// Recursive CTE: skipped from materialise hint, must still
	// produce 1..5.
	row = db.QueryRowContext(ctx, `
WITH RECURSIVE seq AS (
    SELECT 1 AS n UNION ALL SELECT n + 1 FROM seq WHERE n < 5
)
SELECT SUM(n) FROM seq`)
	var s int64
	if err := row.Scan(&s); err != nil {
		t.Fatalf("recursive scan: %v", err)
	}
	if s != 15 {
		t.Errorf("recursive got %d want 15", s)
	}
}

// TestW_NativeWindowAggregators verifies that LOGICAL_OR /
// LOGICAL_AND / BIT_AND / BIT_OR / BIT_XOR / ARRAY_CONCAT_AGG run
// through native ncruces window UDFs (Step / Inverse / Done) instead
// of the predecessor's correlated-subquery emulation. F11 hardens
// this so all built-in window aggregators dispatch natively.
func TestW_NativeWindowAggregators(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	defer func() {
		_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS t`)
		_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS arrs`)
	}()
	if _, err := db.ExecContext(ctx, `CREATE TABLE t (k INT64, b BOOL, n INT64)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO t (k, b, n) VALUES (1, true, 1), (1, false, 2), (1, NULL, 4), (2, true, 8), (2, true, 16)`); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		sql  string
		want any
	}{
		{
			"LOGICAL_OR over partition with mixed values",
			`SELECT LOGICAL_OR(b) OVER (PARTITION BY k) FROM t WHERE k = 1 LIMIT 1`,
			true,
		},
		{
			"LOGICAL_AND over partition with mixed values",
			`SELECT LOGICAL_AND(b) OVER (PARTITION BY k) FROM t WHERE k = 1 LIMIT 1`,
			false,
		},
		{
			"BIT_OR partition fold",
			`SELECT BIT_OR(n) OVER (PARTITION BY k) FROM t WHERE k = 1 LIMIT 1`,
			int64(7),
		},
		{
			"BIT_AND partition fold",
			`SELECT BIT_AND(n) OVER (PARTITION BY k) FROM t WHERE k = 2 LIMIT 1`,
			int64(0),
		},
		{
			"BIT_XOR partition fold",
			`SELECT BIT_XOR(n) OVER (PARTITION BY k) FROM t WHERE k = 1 LIMIT 1`,
			int64(7),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row := db.QueryRowContext(ctx, tc.sql)
			switch want := tc.want.(type) {
			case bool:
				var got bool
				if err := row.Scan(&got); err != nil {
					t.Fatalf("scan: %v", err)
				}
				if got != want {
					t.Errorf("got %v want %v", got, want)
				}
			case int64:
				var got int64
				if err := row.Scan(&got); err != nil {
					t.Fatalf("scan: %v", err)
				}
				if got != want {
					t.Errorf("got %d want %d", got, want)
				}
			}
		})
	}

	// ARRAY_CONCAT_AGG flattens partitions of arrays.
	if _, err := db.ExecContext(ctx, `CREATE TABLE arrs (k INT64, n INT64, a ARRAY<STRING>)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO arrs (k, n, a) VALUES (1, 1, ['10', '11']), (1, 2, ['20']), (2, 3, ['30', '31', '32'])`); err != nil {
		t.Fatal(err)
	}
	rows, err := db.QueryContext(ctx,
		`SELECT k, ARRAY_TO_STRING(ARRAY_CONCAT_AGG(a) OVER (PARTITION BY k ORDER BY n
                ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING), ',') AS s
         FROM arrs ORDER BY k, n`)
	if err != nil {
		t.Fatalf("ARRAY_CONCAT_AGG window: %v", err)
	}
	defer rows.Close()
	type pair struct {
		k int64
		s string
	}
	var got []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.k, &p.s); err != nil {
			t.Fatal(err)
		}
		got = append(got, p)
	}
	if len(got) != 3 {
		t.Fatalf("got %d rows want 3", len(got))
	}
	if got[0].s != "10,11,20" || got[1].s != "10,11,20" {
		t.Errorf("k=1 partitions: %+v", got[:2])
	}
	if got[2].s != "30,31,32" {
		t.Errorf("k=2 partition: %+v", got[2])
	}
}

// TestW_TypeAliasNotInColumnName verifies that the type-alias
// rewriter (INT/INTEGER/SMALLINT/BIGINT/TINYINT -> INT64) does NOT
// touch identifiers in column-name positions. Surfaced when a
// consumer generates `CREATE TABLE x (Int INT64)` for a Go
// struct field literally named `Int`; the column name was being
// rewritten to `INT64`, producing duplicate-name SQL and downstream
// "no such column" errors.
func TestW_TypeAliasNotInColumnName(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CREATE TABLE t (Int INT64)"); err != nil {
		t.Fatalf("CREATE TABLE with column named Int: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO t (Int) VALUES (-1)"); err != nil {
		t.Fatalf("INSERT into column named Int: %v", err)
	}
	var got int
	if err := db.QueryRowContext(ctx, "SELECT ABS(Int) FROM t").Scan(&got); err != nil {
		t.Fatalf("SELECT ABS(Int): %v", err)
	}
	if got != 1 {
		t.Errorf("ABS(-1) over column Int: got %d want 1", got)
	}

	// Verify aliases still rewrite where they ARE types.
	row := db.QueryRowContext(ctx, "SELECT CAST(2 AS INT)")
	var n int64
	if err := row.Scan(&n); err != nil {
		t.Fatalf("CAST(2 AS INT): %v", err)
	}
	if n != 2 {
		t.Errorf("CAST(2 AS INT) got %d want 2", n)
	}
}

// TestW_BqutilNamepathAlias verifies that a UDF registered under a
// single-segment name is also callable via the `bqutil.fn.<name>`
// dataset prefix used by BigQuery community UDFs.
//
// Origin: bqe#318.
func TestW_BqutilNamepathAlias(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`CREATE FUNCTION cw_url_decode(s STRING) RETURNS STRING AS (s)`); err != nil {
		t.Fatalf("CREATE FUNCTION: %v", err)
	}
	for _, q := range []string{
		`SELECT cw_url_decode("hi")`,
		`SELECT bqutil.fn.cw_url_decode("hi")`,
	} {
		var got string
		if err := db.QueryRowContext(ctx, q).Scan(&got); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		if got != "hi" {
			t.Errorf("%s got %q want \"hi\"", q, got)
		}
	}
}

// TestW_SystemVariableAssignAndRead exercises `SET @@var = expr`
// followed by `SELECT @@var`. The driver registers a small set of
// session-scope @@system_variable names on AnalyzerOptions and
// stores their values per-Conn so the read path can inline them as
// literals.
//
// Origin: bqe#147.
func TestW_SystemVariableAssignAndRead(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx := context.Background()

	if _, err := conn.ExecContext(ctx, `SET @@time_zone = "America/Los_Angeles"`); err != nil {
		t.Fatalf("SET @@time_zone: %v", err)
	}
	var got string
	if err := conn.QueryRowContext(ctx, `SELECT @@time_zone`).Scan(&got); err != nil {
		t.Fatalf("SELECT @@time_zone: %v", err)
	}
	if got != "America/Los_Angeles" {
		t.Errorf("@@time_zone got %q want America/Los_Angeles", got)
	}

	// Updating the value must overwrite, not append.
	if _, err := conn.ExecContext(ctx, `SET @@time_zone = "UTC"`); err != nil {
		t.Fatalf("SET @@time_zone (overwrite): %v", err)
	}
	if err := conn.QueryRowContext(ctx, `SELECT @@time_zone`).Scan(&got); err != nil {
		t.Fatalf("SELECT @@time_zone (after overwrite): %v", err)
	}
	if got != "UTC" {
		t.Errorf("@@time_zone after overwrite got %q want UTC", got)
	}
}

// TestW_JsonMutators verifies driver-side runtime UDFs for the JSON
// mutators. SQLite's builtin `json_remove` / `json_set` operate on
// raw JSON text but the driver stores JSON values envelope-encoded,
// so a direct passthrough trips on "malformed JSON". `JSON_STRIP_NULLS`
// has no SQLite analogue at all. All three are implemented as
// driver-side UDFs that decode the envelope before mutating.
//
// Origin: upstream feature.
func TestW_JsonMutators(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	cases := []struct {
		name string
		sql  string
		want string
	}{
		{
			"strip_nulls drops null members recursively",
			`SELECT TO_JSON_STRING(JSON_STRIP_NULLS(JSON '{"a":1,"b":null,"c":{"d":null,"e":2}}'))`,
			`{"a":1,"c":{"e":2}}`,
		},
		{
			"remove single path",
			`SELECT TO_JSON_STRING(JSON_REMOVE(JSON '{"a":1,"b":2}', '$.a'))`,
			`{"b":2}`,
		},
		{
			"remove multiple paths",
			`SELECT TO_JSON_STRING(JSON_REMOVE(JSON '{"a":1,"b":2,"c":3}', '$.a', '$.c'))`,
			`{"b":2}`,
		},
		{
			"set top-level path",
			`SELECT TO_JSON_STRING(JSON_SET(JSON '{"a":1}', '$.b', JSON '2'))`,
			`{"a":1,"b":2}`,
		},
		{
			"set nested path creates missing parent",
			`SELECT TO_JSON_STRING(JSON_SET(JSON '{"a":1}', '$.b.c', JSON '5'))`,
			`{"a":1,"b":{"c":5}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			if err := db.QueryRowContext(ctx, tc.sql).Scan(&got); err != nil {
				t.Fatalf("query: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

// TestW_JsonValueNonLiteralPath verifies that
// `LanguageFeatureFeatureEnableConstantExpressionInJsonPath` is
// enabled — the analyzer must accept a `ResolvedArgumentRef` as the
// JSON path argument inside a CREATE FUNCTION body.
//
// Origin: bqe#357.
func TestW_JsonValueNonLiteralPath(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`CREATE FUNCTION pluck(j JSON, p STRING) RETURNS STRING AS (JSON_VALUE(j, p))`); err != nil {
		t.Fatalf("CREATE FUNCTION with non-literal JSON path: %v", err)
	}
	row := db.QueryRowContext(ctx,
		`SELECT pluck(JSON '{"a": {"b": 7}}', '$.a.b')`)
	var got string
	if err := row.Scan(&got); err != nil {
		t.Fatalf("call pluck: %v", err)
	}
	if got != "7" {
		t.Errorf("pluck got %q want \"7\"", got)
	}
}

// TestW_RangeConstructor covers G2 driver-side wiring. RANGE is a
// first-party GoogleSQL type (TYPE_RANGE = 29 in
// googlesql/public/type.proto, gated by FEATURE_RANGE_TYPE).
//
// Origin: upstream regression.
func TestW_RangeConstructor(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	// Constructor returns a RANGE value; CAST to STRING produces
	// `[start, end)`.
	row := db.QueryRowContext(ctx,
		`SELECT CAST(RANGE(DATE '2024-01-01', DATE '2024-12-31') AS STRING)`)
	var got string
	if err := row.Scan(&got); err != nil {
		t.Fatalf("RANGE scalar: %v", err)
	}
	if got != "[2024-01-01, 2024-12-31)" {
		t.Errorf("RANGE got %q want \"[2024-01-01, 2024-12-31)\"", got)
	}

	// RANGE_START / RANGE_END accessors round-trip through encoded layout.
	row = db.QueryRowContext(ctx, `
SELECT
  CAST(RANGE_START(RANGE(DATE '2024-01-01', DATE '2024-12-31')) AS STRING),
  CAST(RANGE_END(RANGE(DATE '2024-01-01', DATE '2024-12-31')) AS STRING)`)
	var start, end string
	if err := row.Scan(&start, &end); err != nil {
		t.Fatalf("RANGE_START/END: %v", err)
	}
	if start != "2024-01-01" {
		t.Errorf("RANGE_START got %q want 2024-01-01", start)
	}
	if end != "2024-12-31" {
		t.Errorf("RANGE_END got %q want 2024-12-31", end)
	}

	// Unbounded markers.
	row = db.QueryRowContext(ctx, `
SELECT
  RANGE_IS_START_UNBOUNDED(RANGE(CAST(NULL AS DATE), DATE '2024-12-31')),
  RANGE_IS_END_UNBOUNDED(RANGE(DATE '2024-01-01', CAST(NULL AS DATE)))`)
	var sUnb, eUnb bool
	if err := row.Scan(&sUnb, &eUnb); err != nil {
		t.Fatalf("RANGE unbounded check: %v", err)
	}
	if !sUnb {
		t.Error("RANGE_IS_START_UNBOUNDED should be TRUE for NULL start")
	}
	if !eUnb {
		t.Error("RANGE_IS_END_UNBOUNDED should be TRUE for NULL end")
	}
}

// TestW_WithRecursive covers G4 driver-side wiring. SQLite supports
// `WITH RECURSIVE` natively but rejects the recursive-cte reference
// when it is wrapped in a subquery. Our formatter pre-existing path
// already emits ProjectScan as `SELECT ... FROM (...)`; the
// RecursiveScanNode formatter passes that through unchanged so
// `FROM t` stays at the top level of the recursive branch.
//
// Origin: predecessor regression.
func TestW_WithRecursive(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
WITH RECURSIVE t AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM t WHERE n < 5
)
SELECT n FROM t ORDER BY n`)
	if err != nil {
		t.Fatalf("WITH RECURSIVE failed: %v", err)
	}
	defer rows.Close()
	got := []int64{}
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		got = append(got, n)
	}
	want := []int64{1, 2, 3, 4, 5}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("row %d: got %d want %d", i, got[i], want[i])
		}
	}
}

// TestRegression_CTENameWithHyphen: a backtick-quoted CTE name containing a
// hyphen (valid in BigQuery) was emitted unquoted by WithEntryNode, producing
// `near "-": syntax error`. Covers plain and RECURSIVE CTEs. goccy/googlesqlite#17
func TestRegression_CTENameWithHyphen(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	var x int64
	if err := db.QueryRowContext(ctx,
		"WITH `with-hyphen` AS (SELECT 1 AS x) SELECT x FROM `with-hyphen`",
	).Scan(&x); err != nil {
		t.Fatalf("non-recursive hyphenated CTE failed: %v", err)
	}
	if x != 1 {
		t.Errorf("non-recursive: got %d want 1", x)
	}

	var total int64
	if err := db.QueryRowContext(ctx,
		"WITH RECURSIVE `rec-hyphen` AS (SELECT 1 AS n UNION ALL SELECT n + 1 FROM `rec-hyphen` WHERE n < 3) SELECT SUM(n) AS total FROM `rec-hyphen`",
	).Scan(&total); err != nil {
		t.Fatalf("recursive hyphenated CTE failed: %v", err)
	}
	if total != 6 {
		t.Errorf("recursive: got %d want 6", total)
	}
}

// TestRegression_DDLPartitionByAccepted: bqe#152
func TestRegression_DDLPartitionByAccepted(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f5_152?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE t152 (id INT64, ts TIMESTAMP) PARTITION BY DATE(ts)`); err != nil {
		t.Fatalf("PARTITION BY rejected: %v", err)
	}
}

// TestRegression_DDLClusterByAccepted: bqe#373
func TestRegression_DDLClusterByAccepted(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f5_373?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE t373 (id INT64, name STRING) CLUSTER BY id`); err != nil {
		t.Fatalf("CLUSTER BY rejected: %v", err)
	}
}

// TestRegression_DDLColumnOptionsAccepted: bqe#212
func TestRegression_DDLColumnOptionsAccepted(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f5_212?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE t212 (id INT64 OPTIONS(description="primary key"))`); err != nil {
		t.Fatalf("column OPTIONS rejected: %v", err)
	}
}

// TestRegression_DDLDefaultClauseInsertOmitted: bqe#211 / bqe#73
//
// DEFAULT clause persistence and substitution-on-INSERT requires
// catalog-side metadata + an INSERT-time expression evaluator.
// Tracked as a design question for the user, not a one-shot bug fix.
func TestRegression_DDLDefaultClauseInsertOmitted(t *testing.T) {
	t.Parallel()
	t.Skip("DEFAULT clause persistence is a design change; deferred for user discussion")
}

// TestRegression_DDLNotNull: bqe#210
func TestRegression_DDLNotNull(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f5_210?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE t210 (id INT64, name STRING NOT NULL)`); err != nil {
		t.Fatalf("NOT NULL rejected: %v", err)
	}
	// Inserting NULL into a NOT NULL column must error.
	if _, err := db.ExecContext(ctx, `INSERT INTO t210 (id, name) VALUES (1, NULL)`); err == nil {
		t.Fatalf("expected NOT NULL constraint violation, got nil")
	}
}

// TestRegression_CTASWithColumnList: bqe#306
func TestRegression_CTASWithColumnList(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f5_306?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`CREATE TABLE t306 (id INT64, name STRING) AS SELECT 1 AS id, 'a' AS name`); err != nil {
		t.Fatalf("CTAS with column list rejected: %v", err)
	}
}

// TestRegression_UpdateFromSelect: bqe#310
func TestRegression_UpdateFromSelect(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f6_310?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE u310_t (id INT64, v STRING)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE u310_s (id INT64, v STRING)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO u310_t (id, v) VALUES (1, 'old')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO u310_s (id, v) VALUES (1, 'new')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE u310_t SET v = (SELECT v FROM u310_s WHERE u310_s.id = u310_t.id) WHERE TRUE`); err != nil {
		t.Fatalf("UPDATE FROM rejected: %v", err)
	}
}

// TestRegression_ReservedFieldName_End: bqe#402
func TestRegression_ReservedFieldName_End(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f7_402?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE t402 (`end` TIMESTAMP)"); err != nil {
		t.Fatalf("reserved-word identifier rejected: %v", err)
	}
}

// TestRegression_TruncateTable: bqe#378
func TestRegression_TruncateTable(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f7_378?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE t378 (id INT64)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE TABLE t378`); err != nil {
		t.Fatalf("TRUNCATE TABLE rejected: %v", err)
	}
}

// TestRegression_LikeNullLHS: zlite#185
func TestRegression_LikeNullLHS(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	row := db.QueryRowContext(context.Background(),
		`SELECT CAST(NULL AS STRING) LIKE 'foo', CAST(NULL AS INT64) BETWEEN 1 AND 10`)
	var like, between sql.NullBool
	if err := row.Scan(&like, &between); err != nil {
		t.Fatal(err)
	}
	if like.Valid {
		t.Errorf("NULL LIKE 'foo' should be NULL, got %v", like.Bool)
	}
	if between.Valid {
		t.Errorf("NULL BETWEEN 1 AND 10 should be NULL, got %v", between.Bool)
	}
}

// TestRegression_TableSuffixLongestPrefix: bqe#272
func TestRegression_TableSuffixLongestPrefix(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", "file:f9_272?mode=memory&cache=private")
	db.SetMaxOpenConns(1)
	defer db.Close()
	ctx := context.Background()
	for _, stmt := range []string{
		"CREATE TABLE `proj.ds.event_2024_01_01` (id INT64)",
		"CREATE TABLE `proj.ds.event_fresh_2024_01_01` (id INT64)",
		"INSERT INTO `proj.ds.event_2024_01_01` (id) VALUES (1)",
		"INSERT INTO `proj.ds.event_fresh_2024_01_01` (id) VALUES (2)",
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	rows, err := db.QueryContext(ctx,
		"SELECT _TABLE_SUFFIX, id FROM `proj.ds.event_*` ORDER BY 1")
	if err != nil {
		t.Fatalf("wildcard query: %v", err)
	}
	defer rows.Close()
	got := []string{}
	for rows.Next() {
		var s string
		var id int64
		if err := rows.Scan(&s, &id); err != nil {
			t.Fatal(err)
		}
		got = append(got, s)
	}
	// Must include both event_2024_01_01 (suffix "2024_01_01")
	// and event_fresh_2024_01_01 (suffix "fresh_2024_01_01") — i.e.,
	// neither table is matched against the other prefix.
	if len(got) != 2 {
		t.Fatalf("got %d rows, want 2: %v", len(got), got)
	}
}

// TestRegression_TimestampFloatToInt: bqe#390
func TestRegression_TimestampFloatToInt(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	row := db.QueryRowContext(context.Background(),
		`SELECT UNIX_SECONDS(TIMESTAMP '2024-01-01 00:00:00 UTC')`)
	var got int64
	if err := row.Scan(&got); err != nil {
		t.Fatalf("UNIX_SECONDS not scannable as INT64: %v", err)
	}
	if got != 1704067200 {
		t.Fatalf("got %d want 1704067200", got)
	}
}

// TestRegression_DatetimeAsStringSpaceSeparator: bqe#175
func TestRegression_DatetimeAsStringSpaceSeparator(t *testing.T) {
	t.Parallel()
	db, _ := sql.Open("googlesqlite", ":memory:")
	defer db.Close()
	row := db.QueryRowContext(context.Background(),
		`SELECT CAST(DATETIME '2024-01-01 12:34:56' AS STRING)`)
	var got string
	if err := row.Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != "2024-01-01 12:34:56" {
		t.Errorf("CAST DATETIME AS STRING got %q want %q", got, "2024-01-01 12:34:56")
	}
}

// TestRegression_JsonQueryExtractsField asserts that JSON_QUERY and
// JSON_EXTRACT_SCALAR extract the field at the path, not the entire
// JSON document.
func TestRegression_JsonQueryExtractsField(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row := db.QueryRowContext(context.Background(),
		`SELECT JSON_QUERY('{"data": {"test": "hello"}}', '$.data.test'),
		        JSON_EXTRACT_SCALAR('{"data": {"test": "hello"}}', '$.data.test')`)
	var query, scalar sql.NullString
	if err := row.Scan(&query, &scalar); err != nil {
		t.Fatal(err)
	}
	if !query.Valid || query.String != `"hello"` {
		t.Errorf("JSON_QUERY got %#v want \"hello\"", query)
	}
	if !scalar.Valid || scalar.String != "hello" {
		t.Errorf("JSON_EXTRACT_SCALAR got %#v want hello", scalar)
	}
}

// TestRegression_JsonExtractQuotedSegment asserts that JSON_EXTRACT
// accepts a double-quoted path segment containing dots.
func TestRegression_JsonExtractQuotedSegment(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row := db.QueryRowContext(context.Background(),
		`SELECT JSON_EXTRACT('{"relation.parent.name": "name"}', '$."relation.parent.name"')`)
	var got sql.NullString
	if err := row.Scan(&got); err != nil {
		t.Fatal(err)
	}
	want := `"name"`
	if !got.Valid || got.String != want {
		t.Fatalf("got %#v want %q", got, want)
	}
}

// TestRegression_JsonValueOnTypedJsonColumn asserts that JSON_VALUE
// works against a column of type JSON populated via PARSE_JSON.
func TestRegression_JsonValueOnTypedJsonColumn(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", "file:json_typed?mode=memory&cache=private")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "CREATE TABLE x_y (data JSON)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO x_y (data) VALUES (PARSE_JSON('{"something": "1234"}'))`); err != nil {
		t.Fatal(err)
	}
	row := db.QueryRowContext(ctx, `SELECT JSON_VALUE(data, '$.something') FROM x_y`)
	var got sql.NullString
	if err := row.Scan(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Valid || got.String != "1234" {
		t.Fatalf("JSON_VALUE on typed JSON column got %#v want \"1234\"", got)
	}
}

// TestRegression_PercentileContValues asserts that PERCENTILE_CONT
// over [20, 30, 40] returns the documented BigQuery values.
func TestRegression_PercentileContValues(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row := db.QueryRowContext(context.Background(), `
WITH cte AS
    (SELECT 20 AS age UNION ALL SELECT 30 UNION ALL SELECT 40)
SELECT
    CAST(PERCENTILE_CONT(age, 0)    OVER() AS FLOAT64),
    CAST(PERCENTILE_CONT(age, 0.01) OVER() AS FLOAT64),
    CAST(PERCENTILE_CONT(age, 0.5)  OVER() AS FLOAT64),
    CAST(PERCENTILE_CONT(age, 0.99) OVER() AS FLOAT64),
    CAST(PERCENTILE_CONT(age, 1)    OVER() AS FLOAT64)
FROM cte LIMIT 1`)
	var p0, p01, p50, p99, p100 float64
	if err := row.Scan(&p0, &p01, &p50, &p99, &p100); err != nil {
		t.Fatal(err)
	}
	check := func(name string, got, want float64) {
		if abs := got - want; abs < -0.01 || abs > 0.01 {
			t.Errorf("%s: got %g want %g", name, got, want)
		}
	}
	check("p0", p0, 20.0)
	check("p01", p01, 20.2)
	check("p50", p50, 30.0)
	check("p99", p99, 39.8)
	check("p100", p100, 40.0)
}

// TestRegression_ArrayLengthCompareInJoin asserts ARRAY_LENGTH on a
// nullable JOIN result does not panic.
func TestRegression_ArrayLengthCompareInJoin(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
SELECT t1.idx, ARRAY_LENGTH(t1.arr_col), ARRAY_LENGTH(t2.arr_col)
FROM (
    SELECT 1 as idx, [1, 2, 3] as arr_col
    UNION ALL SELECT 2 as idx, [1, 2, 3]
    UNION ALL SELECT 3 as idx, [1, 2, 3, 4]
) as t1
LEFT JOIN (
    SELECT 1 as idx, [1, 2, 3] as arr_col
    UNION ALL SELECT 3 as idx, [1, 2, 3]
    UNION ALL SELECT 4 as idx, [1, 2, 3]
) as t2
ON t1.idx = t2.idx
WHERE ARRAY_LENGTH(t1.arr_col) <> ARRAY_LENGTH(t2.arr_col)
ORDER BY t1.idx`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var idx int64
		var l1, l2 sql.NullInt64
		if err := rows.Scan(&idx, &l1, &l2); err != nil {
			t.Fatal(err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestRegression_RankOverNullColumn asserts that DENSE_RANK / RANK /
// PERCENT_RANK over a column containing nulls do not panic.
//
// Regression: the predecessor panicked in WINDOW_DENSE_RANK.Done with a
// nil pointer dereference.
func TestRegression_RankOverNullColumn(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
WITH t AS (
  SELECT 1 AS x UNION ALL SELECT 2 UNION ALL SELECT NULL
   UNION ALL SELECT 3 UNION ALL SELECT 1 UNION ALL SELECT NULL
)
SELECT
  CAST(DENSE_RANK() OVER (ORDER BY x) AS INT64),
  CAST(RANK() OVER (ORDER BY x) AS INT64),
  CAST(PERCENT_RANK() OVER (ORDER BY x) AS FLOAT64)
FROM t`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var dr, r int64
		var pr float64
		if err := rows.Scan(&dr, &r, &pr); err != nil {
			t.Fatal(err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestRegression_WindowNullPartition asserts that LEAD over a null
// partition value does not panic.
//
// Origin: predecessor regression.
func TestRegression_WindowNullPartition(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
WITH finishers AS
 (SELECT 'Sophia Liu' as name,
  TIMESTAMP '2016-10-18 2:51:45+00' as finish_time,
  'F30-34' as division
  UNION ALL SELECT 'Nilly Nada', TIMESTAMP '2016-10-18 3:10:14+00', NULL)
SELECT
  name,
  COALESCE(division, 'NONE') AS division,
  LEAD(name, 2, 'Nobody')
    OVER (PARTITION BY division ORDER BY finish_time ASC) AS two_runners_back
FROM finishers
ORDER BY finish_time
`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, division, two string
		if err := rows.Scan(&name, &division, &two); err != nil {
			t.Fatal(err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("scan/rows.Err: %v", err)
	}
}

// TestRegression_CountStarOverEmpty asserts COUNT(*) OVER () returns
// the row count for every row.
//
// Origin: the previous driver returned
// "mismatch rowid 1 != 2" because the rewriter assumed an ORDER BY.
func TestRegression_CountStarOverEmpty(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
WITH Produce AS
 (SELECT 'kale' AS item
  UNION ALL SELECT 'banana'
  UNION ALL SELECT 'cabbage'
  UNION ALL SELECT 'apple'
  UNION ALL SELECT 'leek'
  UNION ALL SELECT 'lettuce')
SELECT item, COUNT(*) OVER () AS total
FROM Produce
ORDER BY item
`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var item string
		var total int64
		if err := rows.Scan(&item, &total); err != nil {
			t.Fatal(err)
		}
		if total != 6 {
			t.Fatalf("row %d: total=%d, want 6", count, total)
		}
		count++
	}
	if count != 6 {
		t.Fatalf("got %d rows, want 6", count)
	}
}

// TestRegression_DateTruncIsoweek asserts that DATE_TRUNC(... ISOWEEK)
// returns the Monday of the ISO week containing the given date.
func TestRegression_DateTruncIsoweek(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		desc, query, want string
	}{
		// 2024-01-01 is a Monday (ISO week 1 start).
		{"Monday already", `SELECT DATE_TRUNC(DATE '2024-01-01', ISOWEEK)`, "2024-01-01"},
		// 2024-01-04 is a Thursday → Monday is 2024-01-01.
		{"midweek to Monday", `SELECT DATE_TRUNC(DATE '2024-01-04', ISOWEEK)`, "2024-01-01"},
		// 2024-01-07 is a Sunday → Monday is 2024-01-01.
		{"Sunday to Monday", `SELECT DATE_TRUNC(DATE '2024-01-07', ISOWEEK)`, "2024-01-01"},
		// 2023-01-01 is Sunday, ISO week 52 of 2022; Monday is 2022-12-26.
		{"Sunday rollback to prior year", `SELECT DATE_TRUNC(DATE '2023-01-01', ISOWEEK)`, "2022-12-26"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			db, err := sql.Open("googlesqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			row := db.QueryRowContext(context.Background(), tc.query)
			var got string
			if err := row.Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("query %q: got %q want %q", tc.query, got, tc.want)
			}
		})
	}
}

// TestRegression_GenerateUuidPerRow asserts that GENERATE_UUID()
// returns a different value for every row, instead of being folded
// to a single value across the result set.
//
// Regression: SQLite was constant-folding the call because the function
// was registered with the deterministic flag.
func TestRegression_GenerateUuidPerRow(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
SELECT GENERATE_UUID() FROM UNNEST([1, 2, 3, 4, 5]) AS x`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		seen[v] = true
	}
	if len(seen) < 2 {
		t.Fatalf("GENERATE_UUID() folded to one value across 5 rows: %v", seen)
	}
}

// TestRegression_ParseDateOverflow asserts that PARSE_DATE rejects
// inputs whose components overflow per-component bounds.
//
// Origin: `PARSE_DATE("%D", "99/01/24")`
// returned "2032-03-01" by overflowing the month into the year part.
func TestRegression_ParseDateOverflow(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row := db.QueryRowContext(context.Background(),
		`SELECT PARSE_DATE("%D", "99/01/24")`)
	var got string
	err = row.Scan(&got)
	if err == nil {
		t.Fatalf("expected an error for overflowing month component, got result %q", got)
	}
}

// TestRegression_TrimUnicodeWhitespace asserts that TRIM/LTRIM/RTRIM
// without a cutset argument strip the full Unicode whitespace class,
// not only ASCII space.
//
// Origin: predecessor regression.
func TestRegression_TrimUnicodeWhitespace(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		desc, query, want string
	}{
		{"TRIM tabs", `SELECT TRIM('\tapple\t')`, "apple"},
		{"TRIM mixed whitespace", `SELECT TRIM('\t\n  apple \r\n')`, "apple"},
		{"LTRIM tabs only on left", `SELECT LTRIM('\tapple\t')`, "apple\t"},
		{"RTRIM tabs only on right", `SELECT RTRIM('\tapple\t')`, "\tapple"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			db, err := sql.Open("googlesqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			row := db.QueryRowContext(context.Background(), tc.query)
			var got string
			if err := row.Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("query %q: got %q want %q", tc.query, got, tc.want)
			}
		})
	}
}

// TestRegression_WindowOrderByNulls asserts NULLS FIRST/LAST is honoured
// inside a window's ORDER BY clause.
//
// Origin: predecessor regression.
func TestRegression_WindowOrderByNulls(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `
WITH toks AS (
  SELECT DATETIME "2024-01-07 00:00:00" as dt
  UNION ALL SELECT DATETIME "2024-01-01 00:00:00"
  UNION ALL SELECT CAST(NULL AS DATETIME)
)
SELECT
  CAST(ROW_NUMBER() OVER (ORDER BY dt DESC NULLS FIRST) AS INT64),
  CAST(ROW_NUMBER() OVER (ORDER BY dt DESC NULLS LAST) AS INT64),
  COALESCE(CAST(dt AS STRING), 'null')
FROM toks
ORDER BY 3
`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	type triple struct {
		rn1, rn2 int64
		v        string
	}
	got := []triple{}
	for rows.Next() {
		var v triple
		if err := rows.Scan(&v.rn1, &v.rn2, &v.v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	// After ORDER BY 3 (string ascending): the dates come first, then
	// the literal 'null' string.
	want := []triple{
		{rn1: 3, rn2: 2, v: "2024-01-01T00:00:00"},
		{rn1: 2, rn2: 1, v: "2024-01-07T00:00:00"},
		{rn1: 1, rn2: 3, v: "null"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("row %d: got %+v want %+v", i, got[i], want[i])
		}
	}
}

// TestRegression_FormatDate_PercentU asserts that FORMAT_DATE('%u', ...)
// returns the ISO weekday (Mon=1..Sun=7), matching BigQuery, not the
// ISO week number which the the previous driver returned.
//
// Origin: predecessor regression.
func TestRegression_FormatDate_PercentU(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		desc string
		date string
		want string
	}{
		{"Monday 2024-01-01", "2024-01-01", "1"},
		{"Tuesday 2024-01-02", "2024-01-02", "2"},
		{"Wednesday 2024-01-03", "2024-01-03", "3"},
		{"Thursday 2024-01-04", "2024-01-04", "4"},
		{"Friday 2024-01-05", "2024-01-05", "5"},
		{"Saturday 2024-01-06", "2024-01-06", "6"},
		{"Sunday 2024-01-07", "2024-01-07", "7"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			db, err := sql.Open("googlesqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			row := db.QueryRowContext(context.Background(),
				"SELECT FORMAT_DATE('%u', DATE '"+tc.date+"')")
			var got string
			if err := row.Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("FORMAT_DATE('%%u', %s) got %q want %q", tc.date, got, tc.want)
			}
		})
	}
}

// TestRegression_CovarPop asserts that COVAR_POP divides by n, not n-1.
// Origin: implementation called gonum's
// `stat.Covariance` which returns sample covariance.
func TestRegression_CovarPop(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// x = [1,2,3,4,5], y = [2,4,6,8,10]
	// mean_x = 3, mean_y = 6
	// Sum (xi - 3)*(yi - 6) = 4 + 1 + 0 + 1 + 4 = ... actually:
	// (1-3)*(2-6) + (2-3)*(4-6) + (3-3)*(6-6) + (4-3)*(8-6) + (5-3)*(10-6)
	// = 8 + 2 + 0 + 2 + 8 = 20
	// COVAR_POP = 20 / 5 = 4.0
	// COVAR_SAMP = 20 / 4 = 5.0
	row := db.QueryRowContext(context.Background(), `
SELECT COVAR_POP(y, x) FROM UNNEST([
  STRUCT(1 AS x, 2 AS y),
  STRUCT(2, 4),
  STRUCT(3, 6),
  STRUCT(4, 8),
  STRUCT(5, 10)
])`)
	var got float64
	if err := row.Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 4.0 {
		t.Fatalf("COVAR_POP got %g want 4.0 (sample covariance would be 5.0)", got)
	}
}

// TestRegression_DatetimeDiffSameDay asserts that DATETIME_DIFF(... DAY)
// returns 0 when both arguments are on the same calendar day, even when
// the time components differ.
//
// Origin: the DAY branch rounded any non-zero
// sub-day fractional difference up to 1, returning 1 for a 30-minute
// gap on the same day.
func TestRegression_DatetimeDiffSameDay(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		desc  string
		query string
		want  int64
	}{
		{
			desc:  "same day, 30-minute gap (should be 0)",
			query: `SELECT DATETIME_DIFF(DATETIME "2024-02-01 16:00:00", DATETIME "2024-02-01 15:30:00", DAY)`,
			want:  0,
		},
		{
			desc:  "adjacent dates, 2-minute gap across midnight (should be 1)",
			query: `SELECT DATETIME_DIFF(DATETIME "2024-02-02 00:01:00", DATETIME "2024-02-01 23:59:00", DAY)`,
			want:  1,
		},
		{
			desc:  "calendar dates differ even with reverse time-of-day",
			query: `SELECT DATETIME_DIFF(DATETIME "2024-02-02 01:00:00", DATETIME "2024-02-01 23:00:00", DAY)`,
			want:  1,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			db, err := sql.Open("googlesqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			row := db.QueryRowContext(context.Background(), tc.query)
			var got int64
			if err := row.Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}

// TestRegression_InUnnestArrayParam asserts that an array passed via a
// named parameter survives `SELECT * FROM UNNEST(@arr)` analysis.
func TestRegression_InUnnestArrayParam(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	rows, err := db.QueryContext(ctx,
		"SELECT x FROM UNNEST(@states) AS x ORDER BY x",
		sql.Named("states", []string{"WA", "WI", "WV", "WY"}))
	if err != nil {
		t.Fatalf("UNNEST(@arr) failed analysis: %v", err)
	}
	defer rows.Close()
	got := []string{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	want := []string{"WA", "WI", "WV", "WY"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
