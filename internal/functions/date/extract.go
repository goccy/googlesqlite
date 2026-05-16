package date

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func EXTRACT(v value.Value, part, zone string) (value.Value, error) {
	switch vv := v.(type) {
	case *value.IntervalValue:
		switch part {
		case "YEAR":
			return value.IntValue(vv.Years), nil
		case "MONTH":
			return value.IntValue(vv.Months), nil
		case "DAY":
			return value.IntValue(vv.Days), nil
		case "HOUR":
			return value.IntValue(vv.Hours), nil
		case "MINUTE":
			return value.IntValue(vv.Minutes), nil
		case "SECOND":
			return value.IntValue(vv.Seconds), nil
		case "MILLISECOND":
			return value.IntValue(vv.SubSecondNanos / int32(time.Millisecond)), nil
		case "MICROSECOND":
			return value.IntValue(vv.SubSecondNanos / int32(time.Microsecond)), nil
		}
		return nil, fmt.Errorf("EXTRACT: unexpected part %s for interval", part)
	case value.DateValue, value.DatetimeValue, value.TimeValue, value.TimestampValue:
		t, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		if _, ok := v.(value.TimestampValue); ok {
			loc, err := value.ToLocation(zone)
			if err != nil {
				return nil, err
			}
			t = t.In(loc)
		}
		switch part {
		case "ISOYEAR":
			year, _ := t.ISOWeek()
			return value.IntValue(year), nil
		case "YEAR":
			return value.IntValue(t.Year()), nil
		case "MONTH":
			return value.IntValue(t.Month()), nil
		case "ISOWEEK":
			_, week := t.ISOWeek()
			return value.IntValue(week), nil
		case "WEEK":
			_, week := t.AddDate(0, 0, -int(t.Weekday())).ISOWeek()
			return value.IntValue(week), nil
		case "DAY":
			return value.IntValue(t.Day()), nil
		case "DAYOFYEAR":
			return value.IntValue(t.YearDay()), nil
		case "DAYOFWEEK":
			return value.IntValue(int(t.Weekday()) + 1), nil
		case "QUARTER":
			day := t.YearDay()
			const quarterDays = 91
			switch {
			case day <= quarterDays:
				return value.IntValue(1), nil
			case day <= quarterDays*2:
				return value.IntValue(2), nil
			case day <= quarterDays*3:
				return value.IntValue(3), nil
			}
			return value.IntValue(4), nil
		case "HOUR":
			return value.IntValue(t.Hour()), nil
		case "MINUTE":
			return value.IntValue(t.Minute()), nil
		case "SECOND":
			return value.IntValue(t.Second()), nil
		case "MILLISECOND":
			return value.IntValue(t.Nanosecond() / int(time.Millisecond)), nil
		case "MICROSECOND":
			return value.IntValue(t.Nanosecond() / int(time.Microsecond)), nil
		case "DATE":
			return value.DateValue(t), nil
		case "DATETIME":
			return value.DatetimeValue(t), nil
		case "TIME":
			return value.TimeValue(t), nil
		}
		return nil, fmt.Errorf("EXTRACT: unexpected part %s for data/datetime/time/timestamp", part)
	}
	return nil, fmt.Errorf("EXTRACT: value type must be INTERVAL or DATE or DATETIME or TIME or TIMESTAMP")
}

func BindExtract(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	part, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	zone := "UTC"
	if len(args) == 3 {
		timeZone, err := args[2].ToString()
		if err != nil {
			return nil, err
		}
		zone = timeZone
	}
	return EXTRACT(args[0], part, zone)
}

func BindExtractDate(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	zone := "UTC"
	if len(args) == 2 {
		timeZone, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		zone = timeZone
	}
	return EXTRACT(args[0], "DATE", zone)
}
