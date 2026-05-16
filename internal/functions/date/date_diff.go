package date

import (
	"fmt"
	"strings"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATE_DIFF(a, b time.Time, part string) (value.Value, error) {
	yearISOA, weekA := a.ISOWeek()
	yearISOB, weekB := b.ISOWeek()

	if strings.HasPrefix(part, "WEEK") {
		boundary, ok := helper.WeekPartToOffset[part]

		if !ok {
			return nil, fmt.Errorf("unsupported week date part: %s", part)
		}

		isNegative := false
		start, end := b, a
		if b.Unix() > a.Unix() {
			start, end = a, b
			isNegative = true
		}

		// Manually calculate the number of days based off Unix seconds
		// time.Time.Sub returns "Infinite" max duration for the case of 9999-12-31.Sub(0001-01-01)
		// The maximum time.Duration is ~290 years due to being represented in int64 nanosecond resolution
		days := (end.Unix() - start.Unix()) / 24 / 60 / 60
		// Calculate number of complete weeks between start and end
		fullWeeks := days / 7
		remainder := days % 7

		counts := [7]int64{}

		for _, day := range helper.WeekPartToOffset {
			counts[day] = fullWeeks
		}

		startingDay := int64(start.Weekday())

		for remainder > 0 {
			counts[(startingDay+remainder)%7]++
			remainder--
		}

		result := counts[boundary]

		if isNegative {
			result = -result
		}

		return value.IntValue(result), nil
	}

	switch part {
	case "DAY":
		// Per BigQuery semantics, DATE_DIFF / DATETIME_DIFF with DAY
		// counts the number of calendar-date boundaries between the
		// two arguments, not the floor of (a - b) / 24h. Truncate
		// each side to its calendar date first so the difference is
		// exact in 24-hour multiples regardless of any time-of-day
		// component.
		aDay := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, a.Location())
		bDay := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, b.Location())
		return value.IntValue(int64(aDay.Sub(bDay) / (24 * time.Hour))), nil
	case "ISOWEEK":
		return value.IntValue((a.Year()-b.Year())*48 + weekA - weekB), nil
	case "MONTH":
		return value.IntValue((a.Year()*12 + int(a.Month())) - (b.Year()*12 + int(b.Month()))), nil
	case "QUARTER":
		// Quarters: Jan-Mar = Q1, Apr-Jun = Q2, Jul-Sep = Q3,
		// Oct-Dec = Q4. The number of quarter boundaries between two
		// dates is the difference of (year*4 + quarter_index) values.
		quarterA := int(a.Month()-1)/3 + 1
		quarterB := int(b.Month()-1)/3 + 1
		return value.IntValue(int64((a.Year()*4 + quarterA) - (b.Year()*4 + quarterB))), nil
	case "YEAR":
		return value.IntValue(a.Year() - b.Year()), nil
	case "ISOYEAR":
		return value.IntValue(yearISOA - yearISOB), nil
	}
	return nil, fmt.Errorf("unexpected part value %s", part)
}

var BindDateDiff = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	t2, err := b.ToTime()
	if err != nil {
		return nil, err
	}
	part, err := c.ToString()
	if err != nil {
		return nil, err
	}
	return DATE_DIFF(t, t2, part)
})
