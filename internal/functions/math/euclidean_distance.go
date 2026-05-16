package math

import (
	"fmt"
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

// EUCLIDEAN_DISTANCE returns the L2 distance between two
// equal-length ARRAY<FLOAT64> vectors. NULL vectors propagate.
func EUCLIDEAN_DISTANCE(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("EUCLIDEAN_DISTANCE: invalid number of arguments: got %d, want 2", len(args))
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
		return nil, fmt.Errorf("EUCLIDEAN_DISTANCE: vector lengths differ (%d vs %d)", len(a), len(b))
	}
	var sum float64
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return value.FloatValue(math.Sqrt(sum)), nil
}

// vectorFromValue extracts an ARRAY<FLOAT64> as a Go slice. Used
// by every vector-distance function.
func vectorFromValue(v value.Value) ([]float64, error) {
	arr, err := v.ToArray()
	if err != nil {
		return nil, err
	}
	out := make([]float64, len(arr.Values))
	for i, e := range arr.Values {
		if e == nil {
			return nil, fmt.Errorf("vector element %d is NULL", i)
		}
		f, err := e.ToFloat64()
		if err != nil {
			return nil, err
		}
		out[i] = f
	}
	return out, nil
}
