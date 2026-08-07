package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func CHR(v int64) (value.Value, error) {
	if v == 0 {
		return value.StringValue(""), nil
	}
	if !validCodePoint(v) {
		return nil, fmt.Errorf("CHR: Invalid codepoint %d", v)
	}
	return value.StringValue(string(rune(v))), nil
}

var BindChr = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return CHR(v)
})
