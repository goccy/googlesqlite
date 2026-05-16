package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func RIGHT(val value.Value, length int64) (value.Value, error) {
	if length < 0 {
		return nil, fmt.Errorf("RIGHT: unexpected length val. length must be positive number")
	}
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		runes := []rune(v)
		if len(runes) <= int(length) {
			return val, nil
		}
		return value.StringValue(string(runes[len(runes)-int(length):])), nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		if len(v) <= int(length) {
			return val, nil
		}
		return value.BytesValue(v[len(v)-int(length):]), nil
	}
	return nil, fmt.Errorf("RIGHT: val must be STRING or BYTES")
}

var BindRight = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	length, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return RIGHT(a, length)
})
