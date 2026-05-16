package array

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func RANGE_BUCKET(point value.Value, array *value.ArrayValue) (value.Value, error) {
	if point == nil {
		return nil, nil
	}
	var idx int
	for _, v := range array.Values {
		if v == nil {
			return nil, fmt.Errorf("RANGE_BUCKET: NULL value found in array")
		}
		cond, err := point.GTE(v)
		if err != nil {
			return nil, err
		}
		if !cond {
			break
		}
		idx++
	}
	return value.IntValue(idx), nil
}

var BindRangeBucket = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	array, err := b.ToArray()
	if err != nil {
		return nil, err
	}
	return RANGE_BUCKET(a, array)
})
