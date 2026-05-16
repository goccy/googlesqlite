package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func EQ(a, b value.Value) (value.Value, error) {
	cond, err := a.EQ(b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(cond), nil
}
