package datetime

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/date"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATETIME_ADD(t time.Time, v int64, part string) (value.Value, error) {
	switch part {
	case "MICROSECOND":
		return value.DatetimeValue(t.Add(time.Duration(v) * time.Microsecond)), nil
	case "MILLISECOND":
		return value.DatetimeValue(t.Add(time.Duration(v) * time.Millisecond)), nil
	case "SECOND":
		return value.DatetimeValue(t.Add(time.Duration(v) * time.Second)), nil
	case "MINUTE":
		return value.DatetimeValue(t.Add(time.Duration(v) * time.Minute)), nil
	case "HOUR":
		return value.DatetimeValue(t.Add(time.Duration(v) * time.Hour)), nil
	default:
		date, err := date.DATE_ADD(t, v, part)
		if err != nil {
			return nil, fmt.Errorf("DATETIME_ADD: %w", err)
		}
		datetime, err := date.ToTime()
		if err != nil {
			return nil, fmt.Errorf("DATETIME_ADD: %w", err)
		}
		return value.DatetimeValue(
			time.Date(
				datetime.Year(),
				datetime.Month(),
				datetime.Day(),
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				t.Location(),
			),
		), nil
	}
}

var BindDatetimeAdd = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	num, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	part, err := c.ToString()
	if err != nil {
		return nil, err
	}
	return DATETIME_ADD(t, num, part)
})
