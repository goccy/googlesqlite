package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func IS_FALSE(a value.Value) (value.Value, error) {
	if a == nil {
		return value.BoolValue(false), nil
	}
	b, err := a.ToBool()
	if err != nil {
		return nil, err
	}
	return value.BoolValue(!b), nil
}
