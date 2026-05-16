package array

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_OFFSET(v value.Value, idx int) (value.Value, error) {
	array, err := v.ToArray()
	if err != nil {
		return nil, err
	}
	if idx < 0 || len(array.Values) <= idx {
		return nil, fmt.Errorf("OFFSET(%d) is out of range", idx)
	}
	return array.Values[idx], nil
}

var BindArrayAtOffset = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	i64, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	idx, err := helper.SafeInt(i64)
	if err != nil {
		return nil, err
	}
	return ARRAY_OFFSET(a, idx)
})
