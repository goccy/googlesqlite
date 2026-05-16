package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func UNICODE(v string) (value.Value, error) {
	runes := []rune(v)
	if len(runes) == 0 {
		return value.IntValue(0), nil
	}
	return value.IntValue(runes[0]), nil
}

var BindUnicode = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return UNICODE(v)
})
