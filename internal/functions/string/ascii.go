package string

import (
	"fmt"
	"unicode/utf8"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ASCII(v value.Value) (value.Value, error) {
	if b, ok := v.(value.BytesValue); ok {
		if len(b) == 0 {
			return value.IntValue(0), nil
		}
		return value.IntValue(int64(b[0])), nil
	}
	s, err := v.ToString()
	if err != nil {
		return nil, err
	}
	if s == "" {
		return value.IntValue(0), nil
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r >= utf8.RuneSelf {
		return nil, fmt.Errorf("ASCII: argument is not a structurally valid ASCII string: '%s'", s)
	}
	return value.IntValue(int64(r)), nil
}

var BindAscii = helper.Scalar1(func(a value.Value) (value.Value, error) {
	return ASCII(a)
})
