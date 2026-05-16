package math

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
	"gonum.org/v1/gonum/floats/scalar"
)

func ROUND(x value.Value, precision int) (value.Value, error) {
	xv, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(scalar.Round(xv, precision)), nil
}

var BindRound = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("ROUND: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	var precision = 0
	if len(args) == 2 {
		i64, err := args[1].ToInt64()
		if err != nil {
			return nil, err
		}
		precision, err = helper.SafeInt(i64)
		if err != nil {
			return nil, err
		}
	}
	return ROUND(args[0], precision)
})
