package string

import (
	"unicode/utf8"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func SAFE_CONVERT_BYTES_TO_STRING(val []byte) (value.Value, error) {
	var ret []rune
	for len(val) > 0 {
		r, size := utf8.DecodeRune(val)
		ret = append(ret, r)
		val = val[size:]
	}
	return value.StringValue(string(ret)), nil
}

var BindSafeConvertBytesToString = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToBytes()
	if err != nil {
		return nil, err
	}
	return SAFE_CONVERT_BYTES_TO_STRING(v)
})
