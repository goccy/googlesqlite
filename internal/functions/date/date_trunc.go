package date

import (
	"fmt"
	"strings"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATE_TRUNC(t time.Time, part string) (value.Value, error) {
	yearISO, _ := t.ISOWeek()

	if strings.HasPrefix(part, "WEEK") {
		startOfWeek, ok := helper.WeekPartToOffset[part]
		if !ok {
			return nil, fmt.Errorf("unknown week part: %s", part)
		}

		for int(t.Weekday()) != startOfWeek {
			t = t.AddDate(0, 0, -1)
		}

		return value.DateValue(t), nil
	}

	switch part {
	case "DAY":
		return value.DateValue(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())), nil
	case "ISOWEEK":
		// ISO weeks start on Monday. Walk backwards from the input
		// date until Weekday() == Monday.
		walk := t
		for walk.Weekday() != time.Monday {
			walk = walk.AddDate(0, 0, -1)
		}
		return value.DateValue(time.Date(
			walk.Year(),
			walk.Month(),
			walk.Day(),
			0, 0, 0, 0,
			walk.Location(),
		)), nil
	case "MONTH":
		return value.DateValue(time.Time{}.AddDate(t.Year()-1, int(t.Month())-1, 0)), nil
	case "QUARTER":
		return value.DateValue( // 1, 4, 7, 10
			time.Date(
				t.Year(),
				helper.QuarterStartMonths[int64((t.Month()-1)/3)],
				1,
				0,
				0,
				0,
				0,
				t.Location(),
			),
		), nil
	case "YEAR":
		return value.DateValue(time.Time{}.AddDate(t.Year()-1, 0, 0)), nil
	case "ISOYEAR":
		firstDay := time.Date(
			yearISO,
			1,
			1,
			0,
			0,
			0,
			0,
			t.Location(),
		)
		return value.DateValue(firstDay.AddDate(0, 0, 1-int(firstDay.Weekday()))), nil
	}
	return nil, fmt.Errorf("unexpected part value %s", part)
}

var BindDateTrunc = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	part, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return DATE_TRUNC(t, part)
})
