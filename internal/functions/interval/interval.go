package interval

import (
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/goccy/googlesqlite/internal/value"
)

func INTERVAL(v int64, part string) (value.Value, error) {
	switch part {
	case "YEAR":
		return &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Years: int32(v)}}, nil
	case "MONTH":
		return &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Months: int32(v)}}, nil
	case "DAY":
		return &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Days: int32(v)}}, nil
	case "HOUR":
		return &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Hours: int32(v)}}, nil
	case "MINUTE":
		return &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Minutes: int32(v)}}, nil
	case "SECOND":
		return &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Seconds: int32(v)}}, nil
	case "NANOSECOND":
		return &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{SubSecondNanos: int32(v)}}, nil
	}
	return nil, fmt.Errorf("unexpected interval part: %s", part)
}

func BindInterval(args ...value.Value) (value.Value, error) {
	v, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	part, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	return INTERVAL(v, part)
}
