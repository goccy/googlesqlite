package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func POW(x, y value.Value) (value.Value, error) {
	xf, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	yf, err := y.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(math.Pow(xf, yf)), nil
}
