package datetime

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATETIME(args ...value.Value) (value.Value, error) {
	if len(args) == 6 {
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
		location, err := value.ToLocation("")
		if err != nil {
			return nil, err
		}
		yearInt, err := helper.SafeInt(year)
		if err != nil {
			return nil, err
		}
		monthInt, err := helper.SafeInt(month)
		if err != nil {
			return nil, err
		}
		dayInt, err := helper.SafeInt(day)
		if err != nil {
			return nil, err
		}
		hourInt, err := helper.SafeInt(hour)
		if err != nil {
			return nil, err
		}
		minuteInt, err := helper.SafeInt(minute)
		if err != nil {
			return nil, err
		}
		secondInt, err := helper.SafeInt(second)
		if err != nil {
			return nil, err
		}
		return value.DatetimeValue(time.Date(
			yearInt,
			time.Month(monthInt),
			dayInt,
			hourInt,
			minuteInt,
			secondInt,
			0,
			location,
		)), nil
	}
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("DATETIME: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	switch v := args[0].(type) {
	case value.DateValue:
		t, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		if len(args) == 2 {
			t2, err := args[1].ToTime()
			if err != nil {
				return nil, fmt.Errorf("DATETIME: second argument must be time type: %w", err)
			}
			return value.DatetimeValue(time.Date(
				t.Year(),
				t.Month(),
				t.Day(),
				t2.Hour(),
				t2.Minute(),
				t2.Second(),
				t2.Nanosecond(),
				t2.Location(),
			)), nil
		}
		return value.DatetimeValue(t), nil
	case value.TimestampValue:
		t, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		if len(args) == 2 {
			zone, err := args[1].ToString()
			if err != nil {
				return nil, fmt.Errorf("DATETIME: second argument must be string type: %w", err)
			}
			loc, err := value.ToLocation(zone)
			if err != nil {
				return nil, err
			}
			return value.DatetimeValue(t.In(loc)), nil
		}
		return value.DatetimeValue(t), nil
	}
	return nil, fmt.Errorf("DATETIME: first argument must be DATE or TIMESTAMP type")
}

// BindDatetime short-circuits to NULL when any argument is NULL;
// DATETIME itself performs the arity check.
var BindDatetime = helper.ScalarN(DATETIME)
