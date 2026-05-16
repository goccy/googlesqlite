package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func EXP(x value.Value) (value.Value, error) {
	f, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(math.Exp(f)), nil
}
