package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func SIGN(a value.Value) (value.Value, error) {
	f64, err := a.ToFloat64()
	if err != nil {
		return nil, err
	}
	if math.Signbit(f64) {
		return value.IntValue(-1), nil
	} else if f64 == 0 {
		return value.IntValue(0), nil
	}
	return value.IntValue(1), nil
}
