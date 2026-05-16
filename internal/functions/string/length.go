package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LENGTH(v value.Value) (value.Value, error) {
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		runes := []rune(s)
		return value.IntValue(len(runes)), nil
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.IntValue(len(b)), nil
	}
	return nil, fmt.Errorf("LENGTH: value type is must be STRING or BYTES type")
}

var BindLength = helper.Scalar1(LENGTH)
