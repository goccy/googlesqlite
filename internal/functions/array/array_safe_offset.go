package array

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_SAFE_OFFSET(v value.Value, idx int) (value.Value, error) {
	array, err := v.ToArray()
	if err != nil {
		return nil, err
	}
	if idx < 0 || len(array.Values) <= idx {
		return nil, nil
	}
	return array.Values[idx], nil
}

var BindSafeArrayAtOffset = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	i64, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return ARRAY_SAFE_OFFSET(a, int(i64))
})
