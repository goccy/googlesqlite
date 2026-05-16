package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func SUBSTR(val value.Value, pos int64, length *int64) (value.Value, error) {
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		runes := []rune(v)
		runesLen := int64(len(runes))
		actualPos := substrPos(pos, runesLen)
		actualLen, err := substrLen(length, runesLen)
		if err != nil {
			return nil, err
		}
		startIdx := actualPos
		endIdx := min(actualPos+actualLen, runesLen)
		return value.StringValue(v[startIdx:endIdx]), nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		vLen := int64(len(v))
		actualPos := substrPos(pos, vLen)
		actualLen, err := substrLen(length, vLen)
		if err != nil {
			return nil, err
		}
		startIdx := actualPos
		endIdx := min(actualPos+actualLen, vLen)
		return value.BytesValue(v[startIdx:endIdx]), nil
	}
	return nil, fmt.Errorf("STRPOS: argument type must be STRING or BYTES")
}

var BindSubstr = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, fmt.Errorf("SUBSTR: invalid number of arguments: got %d, want 2 or 3", len(args))
	}
	pos, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	var length *int64
	if len(args) == 3 {
		v, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		length = &v
	}
	return SUBSTR(args[0], pos, length)
})
