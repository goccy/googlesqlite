package interval

import (
	"cloud.google.com/go/bigquery"
	"github.com/goccy/googlesqlite/internal/value"
)

func MAKE_INTERVAL(year, month, day, hour, minute, second int64) (value.Value, error) {
	return &value.IntervalValue{
		IntervalValue: &bigquery.IntervalValue{
			Years:   int32(year),
			Months:  int32(month),
			Days:    int32(day),
			Hours:   int32(hour),
			Minutes: int32(minute),
			Seconds: int32(second),
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
