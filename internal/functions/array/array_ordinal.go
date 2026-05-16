package array

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_ORDINAL(v value.Value, idx int) (value.Value, error) {
	array, err := v.ToArray()
	if err != nil {
		return nil, err
	}
	if idx < 1 || len(array.Values) < idx {
		return nil, fmt.Errorf("ORDINAL(%d) is out of range", idx)
	}
	return array.Values[idx-1], nil
}

var BindArrayAtOrdinal = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	i64, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return ARRAY_ORDINAL(a, int(i64))
})
