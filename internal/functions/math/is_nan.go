package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func IS_NAN(a value.Value) (value.Value, error) {
	f64, err := a.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.BoolValue(math.IsNaN(f64)), nil
}
