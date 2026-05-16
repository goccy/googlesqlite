package date

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// DATE_BUCKET returns the inclusive lower bound of the date
// bucket containing target, given a bucket width interval and an
// optional origin date. Default origin is 1950-01-01.
func DATE_BUCKET(target, origin time.Time, iv *value.IntervalValue) (value.Value, error) {
	if iv != nil && iv.IntervalValue != nil {
		// DATE_BUCKET requires date-only intervals.
		if iv.Hours != 0 || iv.Minutes != 0 || iv.Seconds != 0 || iv.SubSecondNanos != 0 {
			return nil, fmt.Errorf("DATE_BUCKET: bucket_width must consist of date parts only")
		}
	}
	lower, err := helper.BucketFloor(target, origin, iv)
	if err != nil {
		return nil, err
	}
	return value.DateValue(time.Date(lower.Year(), lower.Month(), lower.Day(), 0, 0, 0, 0, time.UTC)), nil
}

func BindDateBucket(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("DATE_BUCKET: invalid number of arguments: got %d, want between 2 and 3", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	target, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	iv, ok := args[1].(*value.IntervalValue)
	if !ok {
		return nil, fmt.Errorf("DATE_BUCKET: bucket_width must be an INTERVAL")
	}
	origin := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(args) == 3 && args[2] != nil {
		o, err := args[2].ToTime()
		if err != nil {
			return nil, err
		}
		origin = time.Date(o.Year(), o.Month(), o.Day(), 0, 0, 0, 0, time.UTC)
	}
	target = time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.UTC)
	return DATE_BUCKET(target, origin, iv)
}
