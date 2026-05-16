package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func GT(a, b value.Value) (value.Value, error) {
	cond, err := a.GT(b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(cond), nil
}
