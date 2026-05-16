package timestamp

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func PARSE_TIMESTAMP(format, date string) (value.Value, error) {
	t, err := helper.ParseTimeFormat(format, date, helper.FormatTypeTimestamp)
	if err != nil {
		return nil, err
	}
	return value.TimestampValue(*t), nil
}

func BindParseTimestamp(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	format, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	target, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	if len(args) == 2 {
		return PARSE_TIMESTAMP(format, target)
	}
	timeZone, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	return PARSE_TIMESTAMP_WITH_TIMEZONE(format, target, timeZone)
}
