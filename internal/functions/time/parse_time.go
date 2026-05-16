package time

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func PARSE_TIME(format, date string) (value.Value, error) {
	t, err := helper.ParseTimeFormat(format, date, helper.FormatTypeTime)
	if err != nil {
		return nil, err
	}
	return value.TimeValue(*t), nil
}

var BindParseTime = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	format, err := a.ToString()
	if err != nil {
		return nil, err
	}
	target, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return PARSE_TIME(format, target)
})
