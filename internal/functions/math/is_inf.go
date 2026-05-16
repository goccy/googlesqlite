package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func IS_INF(a value.Value) (value.Value, error) {
	f64, err := a.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.BoolValue(math.IsInf(f64, 0) || math.IsInf(f64, -1)), nil
}
