package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestApproxQuantilesSortsAndIgnoresNulls covers APPROX_QUANTILES per
// https://cloud.google.com/bigquery/docs/reference/standard-sql/approximate_aggregate_functions#approx_quantiles
// ("NULLs are ignored" unless RESPECT NULLS is given; element 0 is the
// minimum and the last element the maximum). On inputs this small
// BigQuery's answer is exact, so each expected array is BigQuery's.
// The emulator used to pick quantiles in input order and keep NULLs.
func TestApproxQuantilesSortsAndIgnoresNulls(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=approx_quantiles_nulls")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		query string
		want  string
	}{
		{"unsorted input", "SELECT FORMAT('%T', APPROX_QUANTILES(x, 4)) FROM UNNEST([0, 22, 10]) x", "[0, 0, 10, 22, 22]"},
		{"nulls ignored by default", "SELECT FORMAT('%T', APPROX_QUANTILES(x, 4)) FROM UNNEST([6, NULL, NULL]) x", "[6, 6, 6, 6, 6]"},
		{"explicit ignore nulls", "SELECT FORMAT('%T', APPROX_QUANTILES(x, 4 IGNORE NULLS)) FROM UNNEST([6, NULL, NULL]) x", "[6, 6, 6, 6, 6]"},
		{"respect nulls sorts nulls first", "SELECT FORMAT('%T', APPROX_QUANTILES(x, 2 RESPECT NULLS)) FROM UNNEST([6, NULL, NULL]) x", "[NULL, NULL, 6]"},
		{"min and max at the ends", "SELECT FORMAT('%T', APPROX_QUANTILES(x, 1)) FROM UNNEST([5, NULL, 9, 1]) x", "[1, 9]"},
		{"only nulls", "SELECT FORMAT('%T', APPROX_QUANTILES(x, 2)) FROM UNNEST([CAST(NULL AS INT64), NULL]) x", "NULL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got sql.NullString
			if err := db.QueryRowContext(ctx, tc.query).Scan(&got); err != nil {
				t.Fatalf("%s: %v", tc.query, err)
			}
			s := "NULL"
			if got.Valid {
				s = got.String
			}
			if s != tc.want {
				t.Errorf("%s = %q, want %q", tc.query, s, tc.want)
			}
		})
	}
}
