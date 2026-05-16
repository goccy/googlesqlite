package array

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_LENGTH(v *value.ArrayValue) (value.Value, error) {
	return value.IntValue(len(v.Values)), nil
}

// BindArrayLength returns the element count of an array. Per BigQuery,
// ARRAY_LENGTH(NULL) returns NULL (handled by the Scalar1 wrapper).
var BindArrayLength = helper.Scalar1(func(v value.Value) (value.Value, error) {
	arr, err := v.ToArray()
	if err != nil {
		return nil, err
	}
	return ARRAY_LENGTH(arr)
})
