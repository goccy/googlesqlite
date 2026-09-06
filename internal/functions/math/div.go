package math

import (
	"fmt"
	"math/big"

	"github.com/goccy/googlesqlite/internal/value"
)

func DIV(x, y value.Value) (value.Value, error) {
	if xi, ok := x.(value.IntValue); ok {
		if yi, ok := y.(value.IntValue); ok {
			if yi == 0 {
				return nil, fmt.Errorf("DIV: division by zero")
			}
			return value.IntValue(int64(xi) / int64(yi)), nil
		}
	}
	xr, err := x.ToRat()
	if err != nil {
		return nil, err
	}
	yr, err := y.ToRat()
	if err != nil {
		return nil, err
	}
	if yr.Sign() == 0 {
		return nil, fmt.Errorf("DIV: division by zero")
	}
	q := new(big.Rat).Quo(xr, yr)
	n := new(big.Int).Quo(q.Num(), q.Denom())
	return &value.NumericValue{Rat: new(big.Rat).SetInt(n), IsBigNumeric: isBigNumeric(x) || isBigNumeric(y)}, nil
}
