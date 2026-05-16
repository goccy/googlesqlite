package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ASCII(v string) (value.Value, error) {
	return value.IntValue(v[0]), nil
}

var BindAscii = helper.Scalar1(func(a value.Value) (value.Value, error) {
	ascii, err := a.ToString()
	if err != nil {
		return nil, err
	}
	if ascii == "" {
		return value.IntValue(0), nil
	}
	return ASCII(ascii)
})
