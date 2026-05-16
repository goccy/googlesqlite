package timestamp

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/date"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIMESTAMP_TRUNC(t time.Time, part, zone string) (value.Value, error) {
	loc, err := value.ToLocation(zone)
	if err != nil {
		return nil, err
	}
	t = t.In(loc)

	switch part {
	case "MICROSECOND":
		return value.TimestampValue(t), nil
	case "MILLISECOND":
		nsec := (t.Nanosecond() / int(time.Millisecond)) * int(time.Millisecond)
		return value.TimestampValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			t.Second(),
			nsec,
			loc,
		)), nil
	case "SECOND":
		return value.TimestampValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			t.Second(),
			0,
			loc,
		)), nil
	case "MINUTE":
		return value.TimestampValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			0,
			0,
			loc,
		)), nil
	case "HOUR":
		return value.TimestampValue(time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			0,
			0,
			0,
			loc,
		)), nil
	default:
		date, err := date.DATE_TRUNC(t, part)
		if err != nil {
			return nil, fmt.Errorf("TIMESTAMP_TRUNC: %w", err)
		}
		dateTime, err := date.ToTime()
		if err != nil {
			return nil, fmt.Errorf("TIMESTAMP_TRUNC: %w", err)
		}
		return value.TimestampValue(time.Date(
			dateTime.Year(),
			dateTime.Month(),
			dateTime.Day(),
			0,
			0,
			0,
			0,
			loc,
		)), nil
	}
}

func BindTimestampTrunc(args ...value.Value) (value.Value, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, fmt.Errorf("TIMESTAMP_TRUNC: invalid number of arguments: got %d, want 2 or 3", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	part, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	var zone string
	if len(args) == 3 {
		z, err := args[2].ToString()
		if err != nil {
			return nil, err
		}
		zone = z
	}
	return TIMESTAMP_TRUNC(t, part, zone)
}
