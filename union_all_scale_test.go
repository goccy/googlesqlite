package googlesqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/goccy/googlesqlite"
)

// unionAllLiteralQuery builds the shape of a dbt fixture comparison:
// two CTEs of n literal UNION ALL arms each, compared both ways with
// EXCEPT DISTINCT. expected's values are shifted by off, so off == 0
// returns no rows and off != 0 returns every row of both sides.
func unionAllLiteralQuery(n, off int) string {
	arms := func(off int) string {
		var b strings.Builder
		for i := 0; i < n; i++ {
			if i == 0 {
				fmt.Fprintf(&b, "SELECT 'c%d' AS check_name, '%d' AS value", i, i+off)
			} else {
				fmt.Fprintf(&b, " UNION ALL SELECT 'c%d', '%d'", i, i+off)
			}
		}
		return b.String()
	}
	return "WITH result AS (" + arms(0) + "), expected AS (" + arms(off) + ") " +
		"SELECT 'extra', d.* FROM (SELECT * FROM result EXCEPT DISTINCT SELECT * FROM expected) d " +
		"UNION ALL SELECT 'missing', d.* FROM (SELECT * FROM expected EXCEPT DISTINCT SELECT * FROM result) d"
}

// TestLongUnionAllChain covers compound SELECTs with hundreds of arms.
// SQLite plans a compound with one stack frame per arm, and from about
// 210 arms on the wasm stack overflowed: the query failed with "slice
// bounds out of range" and the connection hung on close.
func TestLongUnionAllChain(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, tc := range []struct{ n, off, want int }{
		{n: 500, off: 0, want: 0},
		{n: 1000, off: 0, want: 0},
		{n: 1000, off: 1, want: 2000},
	} {
		t.Run(fmt.Sprintf("n=%d/off=%d", tc.n, tc.off), func(t *testing.T) {
			start := time.Now()
			rows, err := db.QueryContext(context.Background(), unionAllLiteralQuery(tc.n, tc.off))
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			got := 0
			for rows.Next() {
				var kind, name, value string
				if err := rows.Scan(&kind, &name, &value); err != nil {
					t.Fatal(err)
				}
				got++
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %d rows, want %d", got, tc.want)
			}
			t.Logf("%d arms per side: %v", tc.n, time.Since(start))
		})
	}
}

// TestLongUnionDistinctChain checks that UNION DISTINCT still removes
// duplicates across the groups a long chain is split into.
func TestLongUnionDistinctChain(t *testing.T) {
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	arms := make([]string, 300)
	for i := range arms {
		arms[i] = fmt.Sprintf("SELECT %d AS x", i%7)
	}
	var got int
	q := "SELECT COUNT(*) FROM (" + strings.Join(arms, " UNION DISTINCT ") + ")"
	if err := db.QueryRowContext(context.Background(), q).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Fatalf("got %d distinct rows, want 7", got)
	}
}
