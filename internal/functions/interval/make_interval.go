package interval

import (
	"cloud.google.com/go/bigquery"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func MAKE_INTERVAL(year, month, day, hour, minute, second int64) (value.Value, error) {
	year32, err := helper.SafeInt32(year)
	if err != nil {
		return nil, err
	}
	month32, err := helper.SafeInt32(month)
	if err != nil {
		return nil, err
	}
	day32, err := helper.SafeInt32(day)
	if err != nil {
		return nil, err
	}
	hour32, err := helper.SafeInt32(hour)
	if err != nil {
		return nil, err
	}
	minute32, err := helper.SafeInt32(minute)
	if err != nil {
		return nil, err
	}
	second32, err := helper.SafeInt32(second)
	if err != nil {
		return nil, err
	}
	return &value.IntervalValue{
		IntervalValue: &bigquery.IntervalValue{
			Years:   year32,
			Months:  month32,
			Days:    day32,
			Hours:   hour32,
			Minutes: minute32,
			Seconds: second32,
		},
	}, nil
}

func BindMakeInterval(args ...value.Value) (value.Value, error) {
	year, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	month, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	day, err := args[2].ToInt64()
	if err != nil {
		return nil, err
	}
	hour, err := args[3].ToInt64()
	if err != nil {
		return nil, err
	}
	minute, err := args[4].ToInt64()
	if err != nil {
		return nil, err
	}
	second, err := args[5].ToInt64()
	if err != nil {
		return nil, err
	}
	return MAKE_INTERVAL(year, month, day, hour, minute, second)
}
