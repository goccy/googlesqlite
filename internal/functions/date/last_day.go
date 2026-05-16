package date

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LAST_DAY(t time.Time, part string) (value.Value, error) {
	switch part {
	case "YEAR":
		return value.DateValue(time.Date(t.Year()+1, time.Month(1), 0, 0, 0, 0, 0, t.Location())), nil
	case "QUARTER":
		return nil, fmt.Errorf("LAST_DAY: unimplemented QUARTER part")
	case "MONTH":
		return value.DateValue(t.AddDate(0, 1, -t.Day())), nil
	case "WEEK":
		return value.DateValue(t.AddDate(0, 0, 6-int(t.Weekday()))), nil
	case "WEEK_MONDAY":
		return value.DateValue(t.AddDate(0, 0, 7-int(t.Weekday()))), nil
	case "WEEK_TUESDAY":
		return value.DateValue(t.AddDate(0, 0, 8-int(t.Weekday()))), nil
	case "WEEK_WEDNESDAY":
		return value.DateValue(t.AddDate(0, 0, 9-int(t.Weekday()))), nil
	case "WEEK_THURSDAY":
		return value.DateValue(t.AddDate(0, 0, 10-int(t.Weekday()))), nil
	case "WEEK_FRIDAY":
		return value.DateValue(t.AddDate(0, 0, 11-int(t.Weekday()))), nil
	case "WEEK_SATURDAY":
		return value.DateValue(t.AddDate(0, 0, 12-int(t.Weekday()))), nil
	case "ISOWEEK":
		return value.DateValue(t.AddDate(0, 0, 6-int(t.Weekday()))), nil
	case "ISOYEAR":
		return value.DateValue(time.Date(t.Year()+1, time.Month(1), 0, 0, 0, 0, 0, t.Location())), nil
	}
	return nil, fmt.Errorf("LAST_DAY: unexpected part %s", part)
}

func BindLastDay(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("LAST_DAY: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	var part = "MONTH"
	if len(args) == 2 {
		p, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		part = p
	}
	return LAST_DAY(t, part)
}
