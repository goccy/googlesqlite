package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func SECH(x value.Value) (value.Value, error) {
	xv, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(1 / math.Cosh(xv)), nil
}
