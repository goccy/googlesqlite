package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func CHR(v int64) (value.Value, error) {
	if v == 0 {
		return value.StringValue(""), nil
	}
	r, err := helper.SafeInt32(v)
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(rune(r))), nil
}

var BindChr = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return CHR(v)
})
