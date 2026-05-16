package helper

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

// BucketFloor computes the inclusive lower bound of the
// fixed-width bucket — anchored at origin — that contains target.
// Used by DATE_BUCKET / DATETIME_BUCKET / TIMESTAMP_BUCKET.
//
// The interval must be a single-part value (any of YEAR, MONTH,
// DAY, HOUR, MINUTE, SECOND, NANOSECOND, or a single calendar-
// week of 7 days). Mixed month + sub-month parts are rejected
// because month width is calendar-dependent and would change
// per-bucket.
func BucketFloor(target, origin time.Time, iv *value.IntervalValue) (time.Time, error) {
	if iv == nil || iv.IntervalValue == nil {
		return time.Time{}, fmt.Errorf("BUCKET: bucket_width must not be NULL")
	}
	months := int(iv.Years)*12 + int(iv.Months)
	days := int(iv.Days)
	subDayNanos := int64(iv.Hours)*int64(time.Hour) +
		int64(iv.Minutes)*int64(time.Minute) +
		int64(iv.Seconds)*int64(time.Second) +
		int64(iv.SubSecondNanos)
	if months != 0 && (days != 0 || subDayNanos != 0) {
		return time.Time{}, fmt.Errorf("BUCKET: mixed month and sub-month parts are not supported")
	}
	if months != 0 {
		if months < 0 {
			return time.Time{}, fmt.Errorf("BUCKET: bucket_width must be positive")
		}
		return monthBucketFloor(target, origin, months), nil
	}
	widthNanos := int64(days)*24*int64(time.Hour) + subDayNanos
	if widthNanos <= 0 {
		return time.Time{}, fmt.Errorf("BUCKET: bucket_width must be positive")
	}
	return nanosBucketFloor(target, origin, widthNanos), nil
}

// monthBucketFloor anchors month-width buckets on origin.
func monthBucketFloor(target, origin time.Time, monthsPerBucket int) time.Time {
	originMonths := origin.Year()*12 + int(origin.Month()) - 1
	targetMonths := target.Year()*12 + int(target.Month()) - 1
	delta := targetMonths - originMonths
	n := floorDivInt(delta, monthsPerBucket)
	candidate := addMonthsPreserveTimeOfMonth(origin, n*monthsPerBucket)
	// Day-of-month / time-of-day may push us across a boundary the
	// month delta missed. Walk by one bucket either way to settle.
	if candidate.After(target) {
		n--
		candidate = addMonthsPreserveTimeOfMonth(origin, n*monthsPerBucket)
	}
	next := addMonthsPreserveTimeOfMonth(origin, (n+1)*monthsPerBucket)
	if !next.After(target) {
		candidate = next
	}
	return candidate
}

func nanosBucketFloor(target, origin time.Time, widthNanos int64) time.Time {
	diff := target.Sub(origin).Nanoseconds()
	n := floorDivInt64(diff, widthNanos)
	return origin.Add(time.Duration(n * widthNanos))
}

func addMonthsPreserveTimeOfMonth(t time.Time, m int) time.Time {
	y, mo, d := t.Date()
	hh, mm, ss := t.Clock()
	first := time.Date(y, mo, 1, hh, mm, ss, t.Nanosecond(), t.Location())
	after := first.AddDate(0, m, 0)
	ny, nm, _ := after.Date()
	candidate := time.Date(ny, nm, d, hh, mm, ss, t.Nanosecond(), t.Location())
	if candidate.Month() != nm {
		// Day of month doesn't exist in target month — clamp to last day.
		return time.Date(ny, nm+1, 0, hh, mm, ss, t.Nanosecond(), t.Location())
	}
	return candidate
}

func floorDivInt(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

func floorDivInt64(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}
