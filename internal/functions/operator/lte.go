package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func LTE(a, b value.Value) (value.Value, error) {
	cond, err := a.LTE(b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(cond), nil
}
