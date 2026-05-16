package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func TO_CODE_POINTS(v value.Value) (value.Value, error) {
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		ret := &value.ArrayValue{}
		for _, r := range s {
			ret.Values = append(ret.Values, value.IntValue(r))
		}
		return ret, nil
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		ret := &value.ArrayValue{}
		for _, bb := range b {
			ret.Values = append(ret.Values, value.IntValue(bb))
		}
		return ret, nil
	}
	return nil, fmt.Errorf("TO_CODE_POINTS: value type is must be STRING or BYTES type")
}

func BindToCodePoints(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TO_CODE_POINTS: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return &value.ArrayValue{}, nil
	}
	return TO_CODE_POINTS(args[0])
}
