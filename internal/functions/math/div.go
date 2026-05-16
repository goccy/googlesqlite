package math

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func DIV(x, y value.Value) (value.Value, error) {
	xv, err := x.ToInt64()
	if err != nil {
		return nil, err
	}
	yv, err := y.ToInt64()
	if err != nil {
		return nil, err
	}
	if yv == 0 {
		return nil, fmt.Errorf("DIV: zero divided")
	}
	return value.IntValue(xv / yv), nil
}
