package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LEFT(v value.Value, length int64) (value.Value, error) {
	if length < 0 {
		return nil, fmt.Errorf("LEFT: unexpected length value. length must be positive number")
	}
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		runes := []rune(s)
		if len(runes) <= int(length) {
			return v, nil
		}
		return value.StringValue(string(runes[:length])), nil
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		if len(b) <= int(length) {
			return v, nil
		}
		return value.BytesValue(b[:length]), nil
	}
	return nil, fmt.Errorf("LEFT: value type is must be STRING or BYTES type")
}

var BindLeft = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	length, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return LEFT(a, length)
})
