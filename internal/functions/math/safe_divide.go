package math

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func SAFE_DIVIDE(x, y value.Value) (value.Value, error) {
	if isBigNumeric(x) || isBigNumeric(y) || isNumeric(x) || isNumeric(y) {
		yr, err := y.ToRat()
		if err != nil {
			return nil, err
		}
		if yr.Sign() == 0 {
			return nil, nil
		}
		xr, err := x.ToRat()
		if err != nil {
			return nil, err
		}
		q := new(value.NumericValue)
		q.Rat = xr.Quo(xr, yr)
		q.IsBigNumeric = isBigNumeric(x) || isBigNumeric(y)
		return q, nil
	}
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

func isNumeric(v value.Value) bool {
	_, ok := v.(*value.NumericValue)
	return ok
}
