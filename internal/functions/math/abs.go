package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func ABS(a value.Value) (value.Value, error) {
	f64, err := a.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(math.Abs(f64)), nil
}
