package math

import (
	"math/big"

	"github.com/goccy/googlesqlite/internal/value"
)

func SAFE_DIVIDE(x, y value.Value) (value.Value, error) {
	// NUMERIC and BIGNUMERIC must divide exactly and round to their
	// scale; going through float64 loses digits.
	xn, xok := x.(*value.NumericValue)
	yn, yok := y.(*value.NumericValue)
	if xok || yok {
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
		isBig := (xok && xn.IsBigNumeric) || (yok && yn.IsBigNumeric)
		return (&value.NumericValue{Rat: new(big.Rat).Set(xr), IsBigNumeric: isBig}).Div(y)
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
