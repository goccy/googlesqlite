package string

import (
	"fmt"
	"slices"

	"github.com/goccy/googlesqlite/internal/value"
)

func REVERSE(val value.Value) (value.Value, error) {
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		runes := []rune(v)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return value.StringValue(string(runes)), nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		ret := make([]byte, 0, len(v))
		for _, v0 := range slices.Backward(v) {
			ret = append(ret, v0)
		}
		return value.BytesValue(ret), nil
	}
	return nil, fmt.Errorf("REVERSE: val must be STRING or BYTES")
}
