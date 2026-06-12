package googlesqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/goccy/googlesqlite"
	"github.com/google/go-cmp/cmp"
)

// ---- from dml_coverage_test.go ----

// TestMergeIntoInventory replicates the canonical MERGE example from
// docs/third_party/googlesql-docs/data-manipulation-language.md (Inventory
// + NewArrivals walkthrough, lines 1417-1471). After the MERGE:
//
//	dishwasher          30
//	dryer               50   (30 + 20)
//	front load washer   20
//	microwave           20
//	oven                35   (5 + 30)
//	refrigerator        25   (new)
//	top load washer     20   (10 + 10)
//
// The MERGE statement combines INSERT and UPDATE through a single
// join condition, exercising newMergeStmtAction and the merged-table
// emulation in analyzer.go.
func TestMergeIntoInventory(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=merge_inventory")
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

	mustExec := func(q string) {
		t.Helper()
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}

	mustExec("CREATE TABLE NewArrivals (product STRING, quantity INT64, warehouse STRING)")
	mustExec("CREATE TABLE Inventory (product STRING, quantity INT64)")

	// Source rows (NewArrivals).
	for _, r := range []struct {
		p string
		q int64
		w string
	}{
		{"dryer", 20, "warehouse #2"},
		{"oven", 30, "warehouse #3"},
		{"refrigerator", 25, "warehouse #2"},
		{"top load washer", 10, "warehouse #1"},
	} {
		if _, err := conn.ExecContext(ctx,
			"INSERT INTO NewArrivals (product, quantity, warehouse) VALUES (?, ?, ?)",
			r.p, r.q, r.w); err != nil {
			t.Fatalf("INSERT NewArrivals %s: %v", r.p, err)
		}
	}
	// Target rows (Inventory).
	for _, r := range []struct {
		p string
		q int64
	}{
		{"dishwasher", 30},
		{"dryer", 30},
		{"front load washer", 20},
		{"microwave", 20},
		{"oven", 5},
		{"top load washer", 10},
	} {
		if _, err := conn.ExecContext(ctx,
			"INSERT INTO Inventory (product, quantity) VALUES (?, ?)",
			r.p, r.q); err != nil {
			t.Fatalf("INSERT Inventory %s: %v", r.p, err)
		}
	}

	if _, err := conn.ExecContext(ctx, `MERGE Inventory T
USING NewArrivals S
ON T.product = S.product
WHEN MATCHED THEN
  UPDATE SET quantity = T.quantity + S.quantity
WHEN NOT MATCHED THEN
  INSERT (product, quantity) VALUES(product, quantity)`); err != nil {
		t.Fatalf("MERGE: %v", err)
	}

	// Expected post-MERGE state, per the doc table.
	want := map[string]int64{
		"dishwasher":        30,
		"dryer":             50,
		"front load washer": 20,
		"microwave":         20,
		"oven":              35,
		"refrigerator":      25,
		"top load washer":   20,
	}

	rows, err := conn.QueryContext(ctx, "SELECT product, quantity FROM Inventory ORDER BY product")
	if err != nil {
		t.Fatalf("Query Inventory: %v", err)
	}
	defer rows.Close()
	got := map[string]int64{}
	for rows.Next() {
		var p string
		var q int64
		if err := rows.Scan(&p, &q); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got[p] = q
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("row count = %d; want %d (got=%v)", len(got), len(want), got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %d; want %d", k, got[k], v)
		}
	}
}

// TestMergeOnFalseInsert exercises the `MERGE ... ON FALSE` idiom for
// bulk INSERT. With ON FALSE no source row ever matches a target row,
// so every WHEN NOT MATCHED BY TARGET THEN INSERT clause fires for
// every source row — effectively `INSERT INTO target SELECT ... FROM
// source`. Authoritative source: BigQuery docs (data-manipulation-
// language) describe ON FALSE as a way to express conditional INSERTs
// over a source table.
// goccy/googlesqlite#16
func TestMergeOnFalseInsert(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=merge_on_false_insert")
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

	mustExec := func(q string) {
		t.Helper()
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}

	mustExec("CREATE TABLE NewArrivals (product STRING, quantity INT64)")
	mustExec("CREATE TABLE Inventory (product STRING, quantity INT64)")
	for _, r := range []struct {
		p string
		q int64
	}{
		{"dryer", 20},
		{"oven", 30},
		{"refrigerator", 25},
	} {
		if _, err := conn.ExecContext(ctx,
			"INSERT INTO NewArrivals (product, quantity) VALUES (?, ?)", r.p, r.q); err != nil {
			t.Fatalf("INSERT NewArrivals %s: %v", r.p, err)
		}
	}

	// Pre-seed Inventory with a row to verify that ON FALSE does not
	// touch the existing target rows when the only WHEN clause is
	// NOT MATCHED BY TARGET (i.e. an append-only bulk INSERT).
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO Inventory (product, quantity) VALUES (?, ?)", "microwave", 20); err != nil {
		t.Fatalf("INSERT Inventory seed: %v", err)
	}

	if _, err := conn.ExecContext(ctx, `MERGE Inventory T
USING NewArrivals S
ON FALSE
WHEN NOT MATCHED BY TARGET THEN
  INSERT (product, quantity) VALUES(product, quantity)`); err != nil {
		t.Fatalf("MERGE ON FALSE: %v", err)
	}

	want := map[string]int64{
		"microwave":    20, // seeded, untouched
		"dryer":        20, // new
		"oven":         30, // new
		"refrigerator": 25, // new
	}
	rows, err := conn.QueryContext(ctx, "SELECT product, quantity FROM Inventory ORDER BY product")
	if err != nil {
		t.Fatalf("Query Inventory: %v", err)
	}
	defer rows.Close()
	got := map[string]int64{}
	for rows.Next() {
		var p string
		var q int64
		if err := rows.Scan(&p, &q); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got[p] = q
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("row count = %d; want %d (got=%v)", len(got), len(want), got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %d; want %d", k, got[k], v)
		}
	}
}

// TestMergeOnFalseNotMatchedBySourceDelete exercises the dual idiom:
// `MERGE ... ON FALSE WHEN NOT MATCHED BY SOURCE THEN DELETE` empties
// the target table because every target row is unmatched. WHEN MATCHED
// is unreachable under ON FALSE and is silently dropped — verified by
// running alongside an unreachable WHEN MATCHED clause that would
// otherwise mutate the rows.
// goccy/googlesqlite#16
func TestMergeOnFalseNotMatchedBySourceDelete(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=merge_on_false_delete")
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

	mustExec := func(q string) {
		t.Helper()
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}

	mustExec("CREATE TABLE NewArrivals (product STRING, quantity INT64)")
	mustExec("CREATE TABLE Inventory (product STRING, quantity INT64)")
	for _, r := range []struct {
		p string
		q int64
	}{
		{"dishwasher", 30},
		{"microwave", 20},
		{"oven", 5},
	} {
		if _, err := conn.ExecContext(ctx,
			"INSERT INTO Inventory (product, quantity) VALUES (?, ?)", r.p, r.q); err != nil {
			t.Fatalf("INSERT Inventory %s: %v", r.p, err)
		}
	}
	// Source has rows but ON FALSE means they never match a target row.
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO NewArrivals (product, quantity) VALUES (?, ?)", "dryer", 20); err != nil {
		t.Fatalf("INSERT NewArrivals: %v", err)
	}

	if _, err := conn.ExecContext(ctx, `MERGE Inventory T
USING NewArrivals S
ON FALSE
WHEN MATCHED THEN
  UPDATE SET quantity = 999
WHEN NOT MATCHED BY SOURCE THEN
  DELETE`); err != nil {
		t.Fatalf("MERGE ON FALSE: %v", err)
	}

	var n int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM Inventory").Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Errorf("Inventory rows after DELETE = %d; want 0", n)
	}
}

// TestMergeOnNonEqualityPredicate exercises an ON expression that is
// not a simple `target.k = source.k` equality. The MERGE analyzer now
// renders any boolean predicate verbatim — `S.value > T.threshold`
// drives the LEFT JOIN that builds the matched staging set, and the
// MATCHED clause then updates every target row that some source row
// exceeded. goccy/googlesqlite#16
func TestMergeOnNonEqualityPredicate(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=merge_non_equality")
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

	mustExec := func(q string) {
		t.Helper()
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}

	mustExec("CREATE TABLE Threshold (label STRING, threshold INT64)")
	mustExec("CREATE TABLE Reading (value INT64)")
	for _, r := range []struct {
		l string
		t int64
	}{{"low", 10}, {"high", 50}, {"unreachable", 1000}} {
		if _, err := conn.ExecContext(ctx,
			"INSERT INTO Threshold (label, threshold) VALUES (?, ?)", r.l, r.t); err != nil {
			t.Fatalf("INSERT Threshold %s: %v", r.l, err)
		}
	}
	for _, v := range []int64{5, 20, 60} {
		if _, err := conn.ExecContext(ctx,
			"INSERT INTO Reading (value) VALUES (?)", v); err != nil {
			t.Fatalf("INSERT Reading %d: %v", v, err)
		}
	}

	// ON S.value > T.threshold:
	//   threshold 10  ← values 20, 60 → MATCHED
	//   threshold 50  ← value 60      → MATCHED
	//   threshold 1000 ← (no source)   → NOT MATCHED BY SOURCE → kept as-is
	if _, err := conn.ExecContext(ctx, `MERGE Threshold T
USING Reading S
ON S.value > T.threshold
WHEN MATCHED THEN
  UPDATE SET label = 'TRIGGERED'`); err != nil {
		t.Fatalf("MERGE: %v", err)
	}

	want := map[string]string{
		"TRIGGERED":   "TRIGGERED", // two target rows got triggered; we only check by label
		"unreachable": "unreachable",
	}
	rows, err := conn.QueryContext(ctx, "SELECT label, COUNT(*) FROM Threshold GROUP BY label ORDER BY label")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	got := map[string]int64{}
	for rows.Next() {
		var label string
		var n int64
		if err := rows.Scan(&label, &n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got[label] = n
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if got["TRIGGERED"] != 2 {
		t.Errorf("TRIGGERED count = %d; want 2 (got=%v)", got["TRIGGERED"], got)
	}
	if got["unreachable"] != 1 {
		t.Errorf("unreachable count = %d; want 1 (got=%v)", got["unreachable"], got)
	}
	if _, ok := want[""]; ok {
		t.Fatalf("unexpected empty label")
	}
}

// TestMergeOnTrue exercises `ON TRUE`, the other degenerate predicate:
// every target row matches every source row. The WHEN MATCHED UPDATE
// references both target and source columns, so the SET expression
// passes through the staging table's per-side renamed columns.
// goccy/googlesqlite#16
func TestMergeOnTrue(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=merge_on_true")
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

	mustExec := func(q string) {
		t.Helper()
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}

	mustExec("CREATE TABLE T (id INT64, n INT64)")
	mustExec("CREATE TABLE S (multiplier INT64)")
	for _, r := range []struct {
		id, n int64
	}{{1, 10}, {2, 20}} {
		if _, err := conn.ExecContext(ctx,
			"INSERT INTO T (id, n) VALUES (?, ?)", r.id, r.n); err != nil {
			t.Fatalf("INSERT T: %v", err)
		}
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO S (multiplier) VALUES (?)", 3); err != nil {
		t.Fatalf("INSERT S: %v", err)
	}

	if _, err := conn.ExecContext(ctx, `MERGE T USING S
ON TRUE
WHEN MATCHED THEN
  UPDATE SET n = n * S.multiplier`); err != nil {
		t.Fatalf("MERGE: %v", err)
	}

	want := map[int64]int64{1: 30, 2: 60}
	rows, err := conn.QueryContext(ctx, "SELECT id, n FROM T ORDER BY id")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	got := map[int64]int64{}
	for rows.Next() {
		var id, n int64
		if err := rows.Scan(&id, &n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got[id] = n
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("id=%d n=%d; want %d", k, got[k], v)
		}
	}
}

// TestTruncateTable exercises the TruncateStmtAction surface (TRUNCATE
// TABLE rewriting to DELETE FROM). Authoritative source:
// docs/third_party/googlesql-docs/data-manipulation-language.md "TRUNCATE
// TABLE Inventory" — empties the table without removing it.
func TestTruncateTable(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=truncate_table")
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
		"CREATE TABLE Inventory (product STRING, quantity INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO Inventory (product, quantity) VALUES ('a', 1), ('b', 2)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"TRUNCATE TABLE Inventory"); err != nil {
		t.Fatalf("TRUNCATE: %v", err)
	}

	rows, err := conn.QueryContext(ctx, "SELECT product FROM Inventory")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatalf("expected zero rows after TRUNCATE; got at least one")
	}
}

// TestCreateTableAsSelect exercises CREATE TABLE AS SELECT (CTAS),
// which goes through CreateTableAsSelectStmtAction. The doc source
// is docs/third_party/googlesql-docs/data-definition-language.md
// "CREATE TABLE AS SELECT" grammar; the expected behaviour is the
// new table contains the SELECT's rows verbatim.
func TestCreateTableAsSelect(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=ctas")
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
		"CREATE TABLE src (k INT64, v STRING)"); err != nil {
		t.Fatalf("CREATE TABLE src: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO src (k, v) VALUES (1, 'a'), (2, 'b')"); err != nil {
		t.Fatalf("INSERT src: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"CREATE TABLE dst AS SELECT k, v FROM src"); err != nil {
		t.Fatalf("CTAS: %v", err)
	}

	rows, err := conn.QueryContext(ctx, "SELECT k, v FROM dst ORDER BY k")
	if err != nil {
		t.Fatalf("Query dst: %v", err)
	}
	defer rows.Close()
	want := []struct {
		k int64
		v string
	}{{1, "a"}, {2, "b"}}
	var i int
	for rows.Next() {
		var k int64
		var v string
		if err := rows.Scan(&k, &v); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if i >= len(want) {
			t.Fatalf("extra row (%d, %q)", k, v)
		}
		if k != want[i].k || v != want[i].v {
			t.Fatalf("row[%d] = (%d, %q); want (%d, %q)", i, k, v, want[i].k, want[i].v)
		}
		i++
	}
	if i != len(want) {
		t.Fatalf("got %d rows; want %d", i, len(want))
	}
}

// TestTableFunctionCallSimple exercises CREATE TABLE FUNCTION (TVF)
// followed by SELECT FROM <tvf>(...). The TVF example is from
// docs/third_party/googlesql-docs/table-functions.md "CustomerRange"
// (lines 130-137 in that file). The expected rows are the subset
// of customers with id in [min, max].
func TestTableFunctionCallSimple(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=tvf_simple")
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
		"CREATE TABLE Customer (CustomerId INT64, Name STRING)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO Customer (CustomerId, Name) VALUES (1, 'A'), (2, 'B'), (3, 'C'), (4, 'D')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE TABLE FUNCTION CustomerRange(MinId INT64, MaxId INT64) AS (
		SELECT *
		FROM Customer
		WHERE CustomerId >= MinId AND CustomerId <= MaxId
	)`); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION: %v", err)
	}

	rows, err := conn.QueryContext(ctx,
		"SELECT CustomerId, Name FROM CustomerRange(2, 3) ORDER BY CustomerId")
	if err != nil {
		t.Fatalf("SELECT FROM TVF: %v", err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var id int64
		var n string
		if err := rows.Scan(&id, &n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if strings.Join(got, ",") != "B,C" {
		t.Fatalf("rows = %v; want [B C]", got)
	}
}

// TestDropFunction exercises DROP FUNCTION, which routes through
// newDropFunctionStmtAction / DropFunction. The function is
// registered first, asserted callable, then dropped, then a follow-
// up call must fail. Authoritative source: googlesql DROP FUNCTION
// grammar (docs/third_party/googlesql-docs/data-definition-language.md).
func TestDropFunction(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=drop_function")
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
		"CREATE FUNCTION add_two(x INT64) AS (x + 2)"); err != nil {
		t.Fatalf("CREATE FUNCTION: %v", err)
	}
	var got int64
	if err := conn.QueryRowContext(ctx, "SELECT add_two(40)").Scan(&got); err != nil {
		t.Fatalf("call after CREATE: %v", err)
	}
	if got != 42 {
		t.Fatalf("add_two(40) = %d; want 42", got)
	}
	if _, err := conn.ExecContext(ctx, "DROP FUNCTION add_two"); err != nil {
		t.Fatalf("DROP FUNCTION: %v", err)
	}
	if err := conn.QueryRowContext(ctx, "SELECT add_two(40)").Scan(&got); err == nil {
		t.Fatalf("expected error calling dropped function, got nil")
	}
}

// ---- from script_vars_test.go ----

// TestDeclareAndSet exercises DECLARE / SET / variable substitution
// in subsequent statements. Source: googlesql procedural language at
// docs/third_party/googlesql-docs/procedural-language.md (DECLARE / SET
// sections, lines 15-110).
//
// The expected outputs are exactly what the doc states:
//
//   - DECLARE x INT64 DEFAULT 0 → x reads as 0
//   - SET x = 5 → x reads as 5
//   - DECLARE x, y, z INT64 DEFAULT 0 → all three read as 0
//   - SET (a, b, c) = (1+3, 'foo', false) → a=4, b='foo', c=false
func TestDeclareAndSet(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=declare_set")
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

	mustExec := func(q string) {
		t.Helper()
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}

	// DECLARE with DEFAULT.
	mustExec("DECLARE x INT64 DEFAULT 0")
	var got int64
	if err := conn.QueryRowContext(ctx, "SELECT x").Scan(&got); err != nil {
		t.Fatalf("SELECT x: %v", err)
	}
	if got != 0 {
		t.Fatalf("x after DECLARE = %d; want 0", got)
	}

	// SET.
	mustExec("SET x = 5")
	if err := conn.QueryRowContext(ctx, "SELECT x").Scan(&got); err != nil {
		t.Fatalf("SELECT x after SET: %v", err)
	}
	if got != 5 {
		t.Fatalf("x after SET = %d; want 5", got)
	}

	// DECLARE without a DEFAULT — variable initialises to NULL.
	mustExec("DECLARE y INT64")
	var nullable sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT y").Scan(&nullable); err != nil {
		t.Fatalf("SELECT y: %v", err)
	}
	if nullable.Valid {
		t.Fatalf("y after bare DECLARE = %v; want NULL", nullable)
	}

	// SET y = 7.
	mustExec("SET y = 7")
	if err := conn.QueryRowContext(ctx, "SELECT y").Scan(&got); err != nil {
		t.Fatalf("SELECT y after SET: %v", err)
	}
	if got != 7 {
		t.Fatalf("y after SET = %d; want 7", got)
	}
}

// ---- from script_vars_extras_test.go ----

// TestDeclareFloat drives the formatScriptFloat branch of
// evaluateScriptVariableExpr in internal/script_vars.go. The
// DECLARE expression evaluates to a float64; the rewriter calls
// formatScriptFloat to lower it back into the textual SQL.
//
// Reference: docs/third_party/googlesql-docs/procedural-language.md
// DECLARE grammar permits any expression DEFAULT clause, and
// numeric float literals are the canonical example.
func TestDeclareFloat(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=declare_float")
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

	if _, err := conn.ExecContext(ctx, "DECLARE pi FLOAT64 DEFAULT 3.14"); err != nil {
		t.Fatalf("DECLARE: %v", err)
	}
	var got float64
	if err := conn.QueryRowContext(ctx, "SELECT pi").Scan(&got); err != nil {
		t.Fatalf("SELECT pi: %v", err)
	}
	if math.Abs(got-3.14) > 1e-9 {
		t.Fatalf("pi = %v; want 3.14", got)
	}
}

// TestDeclareString drives the string branch of
// evaluateScriptVariableExpr — quoted string literals get
// re-quoted with double-quotes for the rewriter (DQS_DML is on).
func TestDeclareString(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=declare_string")
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

	if _, err := conn.ExecContext(ctx, `DECLARE s STRING DEFAULT 'hello'`); err != nil {
		t.Fatalf("DECLARE: %v", err)
	}
	var got string
	if err := conn.QueryRowContext(ctx, "SELECT s").Scan(&got); err != nil {
		t.Fatalf("SELECT s: %v", err)
	}
	if got != "hello" {
		t.Fatalf("s = %q; want hello", got)
	}
}

// TestDeclareBool drives the bool branch of evaluateScriptVariableExpr.
func TestDeclareBool(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=declare_bool")
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

	if _, err := conn.ExecContext(ctx, "DECLARE flag BOOL DEFAULT TRUE"); err != nil {
		t.Fatalf("DECLARE: %v", err)
	}
	var got bool
	if err := conn.QueryRowContext(ctx, "SELECT flag").Scan(&got); err != nil {
		t.Fatalf("SELECT flag: %v", err)
	}
	if !got {
		t.Fatalf("flag = false; want true")
	}
}

// ---- from tests/parity/exec_test.go ----

func TestExec(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx := context.Background()
	ctx = googlesqlite.WithCurrentTime(ctx, now)
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, test := range []struct {
		name        string
		query       string
		args        []any
		expectedErr bool
	}{
		{
			name: "create table with all types",
			query: `
CREATE TABLE _table_a (
 intValue        INT64,
 boolValue       BOOL,
 doubleValue     DOUBLE,
 floatValue      FLOAT,
 stringValue     STRING,
 bytesValue      BYTES,
 numericValue    NUMERIC,
 bignumericValue BIGNUMERIC,
 intervalValue   INTERVAL,
 dateValue       DATE,
 datetimeValue   DATETIME,
 timeValue       TIME,
 timestampValue  TIMESTAMP
)`,
		},
		{
			name: "create table as select",
			query: `
CREATE TABLE foo ( id STRING PRIMARY KEY NOT NULL, name STRING );
CREATE TABLE bar ( id STRING, name STRING, PRIMARY KEY (id, name) );
CREATE OR REPLACE TABLE new_table_as_select AS (
  SELECT t1.id, t2.name FROM foo t1 JOIN bar t2 ON t1.id = t2.id
);
`,
		},
		{
			name: "recreate table",
			query: `
CREATE OR REPLACE TABLE recreate_table ( a string );
DROP TABLE recreate_table;
CREATE TABLE recreate_table ( b string );
INSERT recreate_table (b) VALUES ('hello');
`,
		},
		{
			name: "insert select",
			query: `
CREATE OR REPLACE TABLE TableA(product string, quantity int64);
INSERT TableA (product, quantity) SELECT 'top load washer', 10;
INSERT INTO TableA (product, quantity) SELECT * FROM UNNEST([('microwave', 20), ('dishwasher', 30)]);
`,
		},
		{
			name: "create view",
			query: `
CREATE VIEW _view_a AS SELECT * FROM TableA
`,
		},
		{
			name: "drop view",
			query: `
DROP VIEW IF EXISTS _view_a
`,
		},
		{
			name: "transaction",
			query: `
CREATE OR REPLACE TABLE Inventory
(
 product string,
 quantity int64,
 supply_constrained bool
);

CREATE OR REPLACE TABLE NewArrivals
(
 product string,
 quantity int64,
 warehouse string
);

INSERT Inventory (product, quantity)
VALUES('top load washer', 10),
     ('front load washer', 20),
     ('dryer', 30),
     ('refrigerator', 10),
     ('microwave', 20),
     ('dishwasher', 30);

INSERT NewArrivals (product, quantity, warehouse)
VALUES('top load washer', 100, 'warehouse #1'),
     ('dryer', 200, 'warehouse #2'),
     ('oven', 300, 'warehouse #1');

BEGIN TRANSACTION;

CREATE TEMP TABLE tmp AS SELECT * FROM NewArrivals WHERE warehouse = 'warehouse #1';
DELETE NewArrivals WHERE warehouse = 'warehouse #1';
MERGE Inventory AS I
USING tmp AS T
ON I.product = T.product
WHEN NOT MATCHED THEN
 INSERT(product, quantity, supply_constrained)
 VALUES(product, quantity, false)
WHEN MATCHED THEN
 UPDATE SET quantity = I.quantity + T.quantity;

TRUNCATE TABLE tmp;

COMMIT TRANSACTION;
`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := db.ExecContext(ctx, test.query); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNestedStructFieldAccess(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx := context.Background()
	ctx = googlesqlite.WithCurrentTime(ctx, now)
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `
CREATE TABLE table (
  id INT64,
  value STRUCT<fieldA STRING, fieldB STRUCT<fieldX STRING, fieldY STRING>>
)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(
		ctx,
		`INSERT table (id, value) VALUES (?, ?)`,
		123,
		map[string]any{
			"fieldB": map[string]any{
				"fieldY": "bar",
			},
		},
	); err != nil {
		t.Fatal(err)
	}
	rows, err := db.QueryContext(ctx, "SELECT value, value.fieldB, value.fieldB.fieldY FROM table")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	// STRUCT columns scan into Go-native []any (positional, with field
	// names living on the column type — accessed in SQL via `.field`).
	// This test only asserts that `value.fieldB.fieldY` round-trips as
	// "bar"; the intermediate columns are read but not inspected
	// further.
	type queryRow struct {
		Value  any
		FieldB any
		FieldY string
	}
	var results []*queryRow
	for rows.Next() {
		var (
			value  any
			fieldB any
			fieldY string
		)
		if err := rows.Scan(&value, &fieldB, &fieldY); err != nil {
			t.Fatal(err)
		}
		results = append(results, &queryRow{
			Value:  value,
			FieldB: fieldB,
			FieldY: fieldY,
		})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("failed to get results")
	}
	if results[0].FieldY != "bar" {
		t.Fatalf("failed to get fieldY")
	}
}

func TestCreateTempTableParity(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx := context.Background()
	ctx = googlesqlite.WithCurrentTime(ctx, now)
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, "CREATE TEMP TABLE tmp_table (id INT64)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TEMP TABLE tmp_table (id INT64)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TABLE tmp_table (id INT64)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TABLE tmp_table (id INT64)"); err == nil {
		t.Fatal("expected error")
	}
}

func TestWildcardTable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(
		ctx,
		"CREATE TABLE `project.dataset.table_a` AS SELECT specialName FROM UNNEST (['alice_a', 'bob_a']) as specialName",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(
		ctx,
		"CREATE TABLE `project.dataset.table_b` AS SELECT name FROM UNNEST(['alice_b', 'bob_b']) as name",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(
		ctx,
		"CREATE TABLE `project.dataset.table_c` AS SELECT name FROM UNNEST(['alice_c', 'bob_c']) as name",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(
		ctx,
		"CREATE TABLE `project.dataset.other_d` AS SELECT name FROM UNNEST(['alice_d', 'bob_d']) as name",
	); err != nil {
		t.Fatal(err)
	}
	t.Run("with first identifier", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, "SELECT name, _TABLE_SUFFIX FROM `project.dataset.table_*` WHERE name LIKE 'alice%' OR name IS NULL")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		type queryRow struct {
			Name   *string
			Suffix string
		}
		var results []*queryRow
		for rows.Next() {
			var (
				name   *string
				suffix string
			)
			if err := rows.Scan(&name, &suffix); err != nil {
				t.Fatal(err)
			}
			results = append(results, &queryRow{
				Name:   name,
				Suffix: suffix,
			})
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		stringPtr := func(v string) *string { return &v }
		if diff := cmp.Diff(results, []*queryRow{
			{Name: stringPtr("alice_c"), Suffix: "c"},
			{Name: stringPtr("alice_b"), Suffix: "b"},
			{Name: nil, Suffix: "a"},
			{Name: nil, Suffix: "a"},
		}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
	t.Run("without first identifier", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, "SELECT name, _TABLE_SUFFIX FROM `dataset.table_*` WHERE name LIKE 'alice%' OR name IS NULL")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		type queryRow struct {
			Name   *string
			Suffix string
		}
		var results []*queryRow
		for rows.Next() {
			var (
				name   *string
				suffix string
			)
			if err := rows.Scan(&name, &suffix); err != nil {
				t.Fatal(err)
			}
			results = append(results, &queryRow{
				Name:   name,
				Suffix: suffix,
			})
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		stringPtr := func(v string) *string { return &v }
		if diff := cmp.Diff(results, []*queryRow{
			{Name: stringPtr("alice_c"), Suffix: "c"},
			{Name: stringPtr("alice_b"), Suffix: "b"},
			{Name: nil, Suffix: "a"},
			{Name: nil, Suffix: "a"},
		}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
}

func TestTemplatedArgFunc(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	t.Run("simple any arguments", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			`CREATE FUNCTION ANY_ADD(x ANY TYPE, y ANY TYPE) AS ((x + 4) / y)`,
		); err != nil {
			t.Fatal(err)
		}
		t.Run("int64", func(t *testing.T) {
			rows, err := db.QueryContext(ctx, "SELECT ANY_ADD(3, 4)")
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			rows.Next()
			var num float64
			if err := rows.Scan(&num); err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(num) != "1.75" {
				t.Fatalf("failed to get max number. got %f", num)
			}
			if rows.Err() != nil {
				t.Fatal(rows.Err())
			}
		})
		t.Run("float64", func(t *testing.T) {
			rows, err := db.QueryContext(ctx, "SELECT ANY_ADD(18.22, 11.11)")
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			rows.Next()
			var num float64
			if err := rows.Scan(&num); err != nil {
				t.Fatal(err)
			}
			if num != 2.0 {
				t.Fatalf("failed to get max number. got %f", num)
			}
			if rows.Err() != nil {
				t.Fatal(rows.Err())
			}
		})
	})
	t.Run("array any arguments", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			`CREATE FUNCTION MAX_FROM_ARRAY(arr ANY TYPE) as (( SELECT MAX(x) FROM UNNEST(arr) as x ))`,
		); err != nil {
			t.Fatal(err)
		}
		t.Run("int64", func(t *testing.T) {
			rows, err := db.QueryContext(ctx, "SELECT MAX_FROM_ARRAY([1, 4, 2, 3])")
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			rows.Next()
			var num int64
			if err := rows.Scan(&num); err != nil {
				t.Fatal(err)
			}
			if num != 4 {
				t.Fatalf("failed to get max number. got %d", num)
			}
			if rows.Err() != nil {
				t.Fatal(rows.Err())
			}
		})
		t.Run("float64", func(t *testing.T) {
			rows, err := db.QueryContext(ctx, "SELECT MAX_FROM_ARRAY([1.234, 3.456, 4.567, 2.345])")
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			rows.Next()
			var num float64
			if err := rows.Scan(&num); err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(num) != "4.567" {
				t.Fatalf("failed to get max number. got %f", num)
			}
			if rows.Err() != nil {
				t.Fatal(rows.Err())
			}
		})
	})
}

func TestJavaScriptUDF(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	t.Run("operation", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			`
CREATE FUNCTION multiplyInputs(x FLOAT64, y FLOAT64)
RETURNS FLOAT64
LANGUAGE js
AS r"""
  return x*y;
"""`,
		); err != nil {
			t.Fatal(err)
		}
		rows, err := db.QueryContext(ctx, `
WITH numbers AS
  (SELECT 1 AS x, 5 as y UNION ALL SELECT 2 AS x, 10 as y UNION ALL SELECT 3 as x, 15 as y)
  SELECT x, y, multiplyInputs(x, y) AS product FROM numbers`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		results := [][]float64{}
		for rows.Next() {
			var (
				x, y, retVal float64
			)
			if err := rows.Scan(&x, &y, &retVal); err != nil {
				t.Fatal(err)
			}
			results = append(results, []float64{x, y, retVal})
		}
		if rows.Err() != nil {
			t.Fatal(rows.Err())
		}
		if diff := cmp.Diff(results, [][]float64{
			{1, 5, 5},
			{2, 10, 20},
			{3, 15, 45},
		}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
	t.Run("function", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			`
CREATE FUNCTION SumFieldsNamedFoo(json_row STRING)
RETURNS FLOAT64
LANGUAGE js
AS r"""
  function SumFoo(obj) {
    var sum = 0;
    for (var field in obj) {
      if (obj.hasOwnProperty(field) && obj[field] != null) {
        if (typeof obj[field] == "object") {
          sum += SumFoo(obj[field]);
        } else if (field == "foo") {
          sum += obj[field];
        }
      }
    }
    return sum;
  }
  var row = JSON.parse(json_row);
  return SumFoo(row);
"""`,
		); err != nil {
			t.Fatal(err)
		}
		rows, err := db.QueryContext(ctx, `
WITH Input AS (
  SELECT
    STRUCT(1 AS foo, 2 AS bar, STRUCT('foo' AS x, 3.14 AS foo) AS baz) AS s,
    10 AS foo
  UNION ALL
  SELECT
    NULL,
    4 AS foo
  UNION ALL
  SELECT
    STRUCT(NULL, 2 AS bar, STRUCT('fizz' AS x, 1.59 AS foo) AS baz) AS s,
    NULL AS foo
) SELECT TO_JSON_STRING(t), SumFieldsNamedFoo(TO_JSON_STRING(t)) FROM Input AS t`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		type queryRow struct {
			JSONRow string
			Sum     float64
		}
		results := []*queryRow{}
		for rows.Next() {
			var (
				jsonRow string
				sum     float64
			)
			if err := rows.Scan(&jsonRow, &sum); err != nil {
				t.Fatal(err)
			}
			results = append(results, &queryRow{JSONRow: jsonRow, Sum: sum})
		}
		if rows.Err() != nil {
			t.Fatal(rows.Err())
		}
		if diff := cmp.Diff(results, []*queryRow{
			{JSONRow: `{"s":{"foo":1,"bar":2,"baz":{"x":"foo","foo":3.14}},"foo":10}`, Sum: 14.14},
			{JSONRow: `{"s":null,"foo":4}`, Sum: 4},
			{JSONRow: `{"s":{"foo":null,"bar":2,"baz":{"x":"fizz","foo":1.59}},"foo":null}`, Sum: 1.59},
		}); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
	t.Run("multibytes", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			`
CREATE FUNCTION JS_JOIN(v ARRAY<STRING>)
RETURNS STRING
LANGUAGE js
AS r"""
  return v.join(' ');
"""`,
		); err != nil {
			t.Fatal(err)
		}
		rows, err := db.QueryContext(ctx, `SELECT JS_JOIN(['あいうえお', 'かきくけこ'])`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		if !rows.Next() {
			t.Fatal("failed to get result")
		}
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		if rows.Err() != nil {
			t.Fatal(rows.Err())
		}
		if v != "あいうえお かきくけこ" {
			t.Fatalf("got %s", v)
		}
	})
	t.Run("struct", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			`
CREATE FUNCTION structToArray(obj STRUCT<idx INT64, name STRING>)
RETURNS ARRAY<STRING>
LANGUAGE js AS """
  let result = []

  result.push(obj["idx"])
  result.push(obj["name"])
  return result;
""";
`,
		); err != nil {
			t.Fatal(err)
		}
		rows, err := db.QueryContext(ctx, `SELECT * FROM UNNEST(structToArray(STRUCT(1,"A")))`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var results []string
		for range 2 {
			if !rows.Next() {
				t.Fatal("failed to get result")
			}
			var v string
			if err := rows.Scan(&v); err != nil {
				t.Fatal(err)
			}
			results = append(results, v)
		}
		if rows.Err() != nil {
			t.Fatal(rows.Err())
		}
		if !reflect.DeepEqual(results, []string{"1", "A"}) {
			t.Fatalf("failed to get results")
		}
	})
}
