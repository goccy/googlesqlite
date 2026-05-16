package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func BYTE_LENGTH(v []byte) (value.Value, error) {
	return value.IntValue(len(v)), nil
}

var BindByteLength = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToBytes()
	if err != nil {
		return nil, err
	}
	return BYTE_LENGTH(v)
})
