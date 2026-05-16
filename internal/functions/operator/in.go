package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func IN(a value.Value, values ...value.Value) (value.Value, error) {
	if a == nil {
		return nil, nil
	}
	for _, v := range values {
		if v == nil {
			continue
		}
		cond, err := a.EQ(v)
		if err != nil {
			return nil, err
		}
		if cond {
			return value.BoolValue(true), nil
		}
	}
	return value.BoolValue(false), nil
}

func BindIn(args ...value.Value) (value.Value, error) {
	return IN(args[0], args[1:]...)
}
