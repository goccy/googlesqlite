package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func ACOSH(x value.Value) (value.Value, error) {
	xv, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(math.Acosh(xv)), nil
}
