package googlesqlite_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// wideCaseChainQuery builds the shape of the dbt snowplow channel
// models: CTE a computes m with an n-branch CASE, and CTE b reads m
// twice in each of its own n branches. Each CTE is referenced once.
func wideCaseChainQuery(n int) string {
	var b strings.Builder
	b.WriteString("WITH a AS (SELECT c, CASE")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, " WHEN LOWER(c) LIKE '%%a%d%%' THEN 'v%d'", i, i)
	}
	b.WriteString(" END AS m FROM t), b AS (SELECT c, CASE")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, " WHEN m = 'v%d' OR m LIKE '%%w%d' THEN 'x%d'", i, i, i)
	}
	b.WriteString(" ELSE 'none' END AS n FROM a) SELECT c, n FROM b ORDER BY c")
	return b.String()
}

// TestWideCaseCTEChainMemory covers single-reference CTEs whose columns
// are large expressions read many times by the next CTE. SQLite
// flattens such a CTE into its consumer by copying the column's
// expression into every reference, so 400 branches meant 800 copies of
// a 400-branch CASE, and preparing the 34 KB statement failed with
// "sqlite3: out of memory" on the 256 MiB wasm heap. The dbt
// flipto_web_sessions_this_run and flipto_web_page_views_this_run
// models failed the same way on empty tables.
func TestWideCaseCTEChainMemory(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, q := range []string{
		"CREATE TABLE t (c STRING)",
		"INSERT INTO t VALUES ('xa7y'), ('zz'), ('a399')",
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	rows, err := db.Query(wideCaseChainQuery(400))
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var c, n string
		if err := rows.Scan(&c, &n); err != nil {
			t.Fatal(err)
		}
		got = append(got, c+"="+n)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	want := []string{"a399=x3", "xa7y=x7", "zz=none"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}
}
