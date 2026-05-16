package timestamp

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIMESTAMP_SUB(t time.Time, v int64, part string) (value.Value, error) {
	switch part {
	case "MICROSECOND":
		return value.TimestampValue(t.Add(-time.Duration(v) * time.Microsecond)), nil
	case "MILLISECOND":
		return value.TimestampValue(t.Add(-time.Duration(v) * time.Millisecond)), nil
	case "SECOND":
		return value.TimestampValue(t.Add(-time.Duration(v) * time.Second)), nil
	case "MINUTE":
		return value.TimestampValue(t.Add(-time.Duration(v) * time.Minute)), nil
	case "HOUR":
		return value.TimestampValue(t.Add(-time.Duration(v) * time.Hour)), nil
	case "DAY":
		return value.TimestampValue(t.AddDate(0, 0, -int(v))), nil
	}
	return nil, fmt.Errorf("TIMESTAMP_SUB: unexpected part value %s", part)
}

var BindTimestampSub = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	num, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	part, err := c.ToString()
	if err != nil {
		return nil, err
	}
	return TIMESTAMP_SUB(t, num, part)
})
