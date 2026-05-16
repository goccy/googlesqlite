package timestamp

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func PARSE_TIMESTAMP_WITH_TIMEZONE(format, date, zone string) (value.Value, error) {
	t, err := helper.ParseTimeFormat(format, date, helper.FormatTypeTimestamp)
	if err != nil {
		return nil, err
	}
	loc, err := value.ToLocation(zone)
	if err != nil {
		return nil, err
	}
	modified, err := value.ModifyTimeZone(*t, loc)
	if err != nil {
		return nil, err
	}
	return value.TimestampValue(modified), nil
}
