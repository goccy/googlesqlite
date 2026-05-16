package string

import (
	"regexp"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func REGEXP_CONTAINS(val, expr string) (value.Value, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(re.MatchString(val)), nil
}

var BindRegexpContains = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	expr, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return REGEXP_CONTAINS(v, expr)
})
