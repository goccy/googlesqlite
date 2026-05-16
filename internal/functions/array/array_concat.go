package array

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_CONCAT(args ...value.Value) (value.Value, error) {
	arr := &value.ArrayValue{}
	for _, arg := range args {
		subarr, err := arg.ToArray()
		if err != nil {
			return nil, err
		}
		arr.Values = append(arr.Values, subarr.Values...)
	}
	return arr, nil
}

func BindArrayConcat(args ...value.Value) (value.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("ARRAY_CONCAT: required arguments")
	}
	return ARRAY_CONCAT(args...)
}
