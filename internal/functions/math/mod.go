package math

import (
	"fmt"
	"math"

	"github.com/goccy/googlesqlite/internal/value"
)

func MOD(x, y value.Value) (value.Value, error) {
	xv, err := x.ToFloat64()
	if err != nil {
		return nil, err
	}
	yv, err := y.ToFloat64()
	if err != nil {
		return nil, err
	}
	if yv == 0 {
		return nil, fmt.Errorf("MOD: zero divided")
	}
	return value.FloatValue(math.Mod(xv, yv)), nil
}
