package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func NOT(a value.Value) (value.Value, error) {
	v, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return value.BoolValue(v == 0), nil
}
