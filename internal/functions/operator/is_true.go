package operator

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func IS_TRUE(a value.Value) (value.Value, error) {
	b, err := a.ToBool()
	if err != nil {
		return nil, err
	}
	return value.BoolValue(b), nil
}
