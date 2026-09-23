package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestNumericMultiplicationScale covers NUMERIC and BIGNUMERIC
// multiplication. A product of two scale-9 (or scale-38) values can carry
// twice the scale; BigQuery rounds it back to the type's scale, half away
// from zero. The emulator used to keep the exact product.
func TestNumericMultiplicationScale(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=numeric_multiplication_scale")
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
		{"numeric product rounds half away from zero", "SELECT CAST(NUMERIC '0.000000001' * NUMERIC '0.5' AS STRING)", "0.000000001"},
		{"negative numeric product rounds away from zero", "SELECT CAST(NUMERIC '-0.000000001' * NUMERIC '0.5' AS STRING)", "-0.000000001"},
		{"numeric product rounds down below half", "SELECT CAST(NUMERIC '0.000000001' * NUMERIC '0.4' AS STRING)", "0"},
		{"sum of rounded numeric products", "SELECT CAST(NUMERIC '0.000000001' * NUMERIC '0.5' + NUMERIC '0.000000001' * NUMERIC '0.5' AS STRING)", "0.000000002"},
		{"sum of rounded negative numeric products", "SELECT CAST(NUMERIC '-0.000000001' * NUMERIC '0.5' + NUMERIC '-0.000000001' * NUMERIC '0.5' AS STRING)", "-0.000000002"},
		{"numeric product equals rounded literal", "SELECT CAST(NUMERIC '1.000000001' * NUMERIC '1.000000001' = NUMERIC '1.000000002' AS STRING)", "true"},
		{"bignumeric product rounds to scale 38", "SELECT CAST(BIGNUMERIC '0.00000000000000000000000000000000000001' * BIGNUMERIC '0.5' AS STRING)", "0.00000000000000000000000000000000000001"},
		{"negative bignumeric product rounds away from zero", "SELECT CAST(BIGNUMERIC '-0.00000000000000000000000000000000000001' * BIGNUMERIC '0.5' AS STRING)", "-0.00000000000000000000000000000000000001"},
		{"sum of rounded bignumeric products", "SELECT CAST(BIGNUMERIC '0.00000000000000000000000000000000000001' * BIGNUMERIC '0.5' + BIGNUMERIC '0.00000000000000000000000000000000000001' * BIGNUMERIC '0.5' AS STRING)", "0.00000000000000000000000000000000000002"},
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
