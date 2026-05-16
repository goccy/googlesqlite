package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func CODE_POINTS_TO_STRING(v *value.ArrayValue) (value.Value, error) {
	runes := make([]rune, 0, len(v.Values))
	for _, vv := range v.Values {
		if vv == nil {
			return nil, nil
		}
		i64, err := vv.ToInt64()
		if err != nil {
			return nil, err
		}
		if i64 == 0 {
			continue
		}
		r, err := helper.SafeInt32(i64)
		if err != nil {
			return nil, err
		}
		runes = append(runes, rune(r))
	}
	return value.StringValue(string(runes)), nil
}

var BindCodePointsToString = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToArray()
	if err != nil {
		return nil, err
	}
	return CODE_POINTS_TO_STRING(v)
})
