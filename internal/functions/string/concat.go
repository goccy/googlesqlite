package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func CONCAT(args ...value.Value) (value.Value, error) {
	var ret []byte
	for _, v := range args {
		if v == nil {
			continue
		}
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		ret = append(ret, b...)
	}
	switch args[0].(type) {
	case value.StringValue:
		return value.StringValue(string(ret)), nil
	case value.BytesValue:
		return value.BytesValue(ret), nil
	}
	return nil, fmt.Errorf("CONCAT: argument type must be STRING or BYTES")
}

var BindConcat = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("CONCAT: invalid number of arguments: got %d, want at least 1", len(args))
	}
	return CONCAT(args...)
})
