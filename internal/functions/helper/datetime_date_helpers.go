package helper

import "time"

// Shared helpers used by date / datetime / timestamp / time
// per-spec functions.

var WeekPartToOffset = map[string]int{
	"WEEK":           0,
	"WEEK_MONDAY":    1,
	"WEEK_TUESDAY":   2,
	"WEEK_WEDNESDAY": 3,
	"WEEK_THURSDAY":  4,
	"WEEK_FRIDAY":    5,
	"WEEK_SATURDAY":  6,
}

var QuarterStartMonths = []time.Month{time.January, time.April, time.July, time.October}

func AddMonth(t time.Time, m int) time.Time {
	curYear, curMonth, curDay := t.Date()

	first := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, t.Location())
	year, month, _ := first.AddDate(0, m, 0).Date()
	after := time.Date(year, month, curDay, 0, 0, 0, 0, time.UTC)
	if month != after.Month() {
		return first.AddDate(0, m+1, -1)
	}
	return t.AddDate(0, m, 0)
}

func AddYear(t time.Time, y int) time.Time {
	curYear, curMonth, curDay := t.Date()

	first := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, t.Location())
	year, month, _ := first.AddDate(y, 0, 0).Date()
	after := time.Date(year, month, curDay, 0, 0, 0, 0, t.Location())
	if month != after.Month() {
		return first.AddDate(y, 1, -1)
	}
	return t.AddDate(y, 0, 0)
}
