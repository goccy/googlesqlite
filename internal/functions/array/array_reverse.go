package array

import (
	"fmt"
	"slices"

	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_REVERSE(v *value.ArrayValue) (value.Value, error) {
	ret := &value.ArrayValue{}
	for _, v0 := range slices.Backward(v.Values) {
		ret.Values = append(ret.Values, v0)
	}
	return ret, nil
}

func BindArrayReverse(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ARRAY_REVERSE: invalid number of arguments: got %d, want 1", len(args))
	}
	arr, err := args[0].ToArray()
	if err != nil {
		return nil, err
	}
	return ARRAY_REVERSE(arr)
}
