package math

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func SAFE_DIVIDE(x, y value.Value) (value.Value, error) {
	xv, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	yv, err := y.ToFloat64()
	if err != nil {
		return nil, err
	}
	if yv == 0 {
		return nil, nil
	}
	return value.FloatValue(xv / yv), nil
}
