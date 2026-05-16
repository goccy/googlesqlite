package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func CHAR_LENGTH(v []byte) (value.Value, error) {
	return value.IntValue(len([]rune(string(v)))), nil
}

var BindCharLength = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToBytes()
	if err != nil {
		return nil, err
	}
	return CHAR_LENGTH(v)
})
