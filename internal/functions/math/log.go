package math

import (
	"math"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LOG(x, y value.Value) (value.Value, error) {
	xf, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	yi, err := x.ToInt64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(math.Ldexp(xf, int(yi))), nil
}

var BindLog = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) == 1 {
		return LN(args[0])
	}
	return LOG(args[0], args[1])
})
