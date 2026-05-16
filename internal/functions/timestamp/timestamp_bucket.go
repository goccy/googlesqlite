package timestamp

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// TIMESTAMP_BUCKET returns the inclusive lower bound of the
// timestamp bucket containing target, given a bucket width
// interval and an optional origin timestamp. Default origin is
// 1950-01-01 00:00:00 UTC.
func TIMESTAMP_BUCKET(target, origin time.Time, iv *value.IntervalValue) (value.Value, error) {
	lower, err := helper.BucketFloor(target, origin, iv)
	if err != nil {
		return nil, err
	}
	return value.TimestampValue(lower), nil
}

func BindTimestampBucket(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("TIMESTAMP_BUCKET: invalid number of arguments: got %d, want between 2 and 3", len(args))
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
		return nil, fmt.Errorf("TIMESTAMP_BUCKET: bucket_width must be an INTERVAL")
	}
	origin := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(args) == 3 && args[2] != nil {
		o, err := args[2].ToTime()
		if err != nil {
			return nil, err
		}
		origin = o.UTC()
	}
	return TIMESTAMP_BUCKET(target.UTC(), origin, iv)
}
