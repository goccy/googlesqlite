package datetime

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/date"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATETIME_TRUNC(t time.Time, part string) (value.Value, error) {
	switch part {
	case "MICROSECOND":
		return value.DatetimeValue(t), nil
	case "MILLISECOND":
		sec := time.Duration(t.Second()) - time.Duration(t.Second())/time.Microsecond
		return value.DatetimeValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			int(sec),
			0,
			t.Location(),
		)), nil
	case "SECOND":
		sec := time.Duration(t.Second()) / time.Second
		return value.DatetimeValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			int(sec),
			0,
			t.Location(),
		)), nil
	case "MINUTE":
		return value.DatetimeValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			0,
			0,
			t.Location(),
		)), nil
	case "HOUR":
		return value.DatetimeValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			0,
			0,
			0,
			t.Location(),
		)), nil
	default:
		date, err := date.DATE_TRUNC(t, part)
		if err != nil {
			return nil, fmt.Errorf("DATETIME_TRUNC: %w", err)
		}
		datetime, err := date.ToTime()
		if err != nil {
			return nil, fmt.Errorf("DATETIME_TRUNC: %w", err)
		}
		return value.DatetimeValue(
			time.Date(
				datetime.Year(),
				datetime.Month(),
				datetime.Day(),
				datetime.Hour(),
				datetime.Minute(),
				datetime.Second(),
				datetime.Nanosecond(),
				datetime.Location(),
			),
		), nil
	}
}

var BindDatetimeTrunc = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	part, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return DATETIME_TRUNC(t, part)
})
