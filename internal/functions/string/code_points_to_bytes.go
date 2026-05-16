package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func CODE_POINTS_TO_BYTES(v *value.ArrayValue) (value.Value, error) {
	b := make([]byte, 0, len(v.Values))
	for _, vv := range v.Values {
		i64, err := vv.ToInt64()
		if err != nil {
			return nil, err
		}
		b = append(b, byte(i64))
	}
	return value.BytesValue(b), nil
}

var BindCodePointsToBytes = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToArray()
	if err != nil {
		return nil, err
	}
	return CODE_POINTS_TO_BYTES(v)
})
