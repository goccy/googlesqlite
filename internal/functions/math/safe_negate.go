package math

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func SAFE_NEGATE(x value.Value) (value.Value, error) {
	xv, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(-xv), nil
}
