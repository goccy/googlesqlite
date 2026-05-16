package math

import (
	"fmt"
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

// COSINE_DISTANCE returns 1 - cosine_similarity(a, b) between two
// ARRAY<FLOAT64> vectors. NULL inputs propagate. Vectors must
// have the same length; either being all-zero is an error
// (matches BigQuery behaviour).
func COSINE_DISTANCE(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("COSINE_DISTANCE: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	a, err := vectorFromValue(args[0])
	if err != nil {
		return nil, err
	}
	b, err := vectorFromValue(args[1])
	if err != nil {
		return nil, err
	}
	if len(a) != len(b) {
		return nil, fmt.Errorf("COSINE_DISTANCE: vector lengths differ (%d vs %d)", len(a), len(b))
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return nil, fmt.Errorf("COSINE_DISTANCE: zero-magnitude vector")
	}
	return value.FloatValue(1 - dot/(math.Sqrt(na)*math.Sqrt(nb))), nil
}
