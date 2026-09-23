package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestNumericDivisionScale covers NUMERIC and BIGNUMERIC division. Per
// https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types#decimal_types
// NUMERIC has scale 9 and BIGNUMERIC scale 38, so a quotient is rounded
// to that many decimal places, half away from zero. The emulator used to
// keep the exact rational, so an equality against the rounded literal
// was false and sums of quotients drifted from BigQuery.
func TestNumericDivisionScale(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=numeric_division_scale")
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
		{"numeric divide rounds to scale 9", "SELECT CAST(NUMERIC '100' / 3 = NUMERIC '33.333333333' AS STRING)", "true"},
		{"numeric divide rounds half away from zero", "SELECT CAST(NUMERIC '2' / 3 = NUMERIC '0.666666667' AS STRING)", "true"},
		{"negative numeric divide rounds away from zero", "SELECT CAST(NUMERIC '-2' / 3 = NUMERIC '-0.666666667' AS STRING)", "true"},
		{"sum of rounded quotients", "SELECT CAST(NUMERIC '1' / 3 + NUMERIC '1' / 3 + NUMERIC '1' / 3 AS STRING)", "0.999999999"},
		{"safe_divide on numeric", "SELECT CAST(SAFE_DIVIDE(NUMERIC '100', 3) = NUMERIC '33.333333333' AS STRING)", "true"},
		{"safe_divide on numeric keeps full scale", "SELECT CAST(SAFE_DIVIDE(NUMERIC '1', 7) AS STRING)", "0.142857143"},
		{"safe_divide on numeric by zero", "SELECT CAST(SAFE_DIVIDE(NUMERIC '100', 0) IS NULL AS STRING)", "true"},
		{"bignumeric divide rounds to scale 38", "SELECT CAST(BIGNUMERIC '2' / 3 = BIGNUMERIC '0.66666666666666666666666666666666666667' AS STRING)", "true"},
		{"safe_divide on bignumeric", "SELECT CAST(SAFE_DIVIDE(BIGNUMERIC '1', 3) AS STRING)", "0.33333333333333333333333333333333333333"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			if err := db.QueryRowContext(ctx, tc.query).Scan(&got); err != nil {
				t.Fatalf("%s: %v", tc.query, err)
			}
			if got != tc.want {
				t.Errorf("%s = %q, want %q", tc.query, got, tc.want)
			}
		})
	}
}
