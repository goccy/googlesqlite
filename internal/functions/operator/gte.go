package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func GTE(a, b value.Value) (value.Value, error) {
	cond, err := a.GTE(b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(cond), nil
}
