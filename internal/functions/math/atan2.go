package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func ATAN2(x, y value.Value) (value.Value, error) {
	xv, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	yv, err := y.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(math.Atan2(xv, yv)), nil
}
