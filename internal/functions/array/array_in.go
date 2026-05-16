package array

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_IN(a, b value.Value) (value.Value, error) {
	array, err := b.ToArray()
	if err != nil {
		return nil, err
	}
	cond, err := array.Has(a)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(cond), nil
}

func BindInArray(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return value.BoolValue(false), nil
	}
	return ARRAY_IN(args[0], args[1])
}
