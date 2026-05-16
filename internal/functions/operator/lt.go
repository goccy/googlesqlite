package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func LT(a, b value.Value) (value.Value, error) {
	cond, err := a.LT(b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(cond), nil
}
