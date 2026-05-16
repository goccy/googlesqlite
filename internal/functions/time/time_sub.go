package time

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIME_SUB(t time.Time, v int64, part string) (value.Value, error) {
	switch part {
	case "MICROSECOND":
		return value.TimeValue(t.Add(-time.Duration(v) * time.Microsecond)), nil
	case "MILLISECOND":
		return value.TimeValue(t.Add(-time.Duration(v) * time.Millisecond)), nil
	case "SECOND":
		return value.TimeValue(t.Add(-time.Duration(v) * time.Second)), nil
	case "MINUTE":
		return value.TimeValue(t.Add(-time.Duration(v) * time.Minute)), nil
	case "HOUR":
		return value.TimeValue(t.Add(-time.Duration(v) * time.Hour)), nil
	}
	return nil, fmt.Errorf("TIME_SUB: unexpected part value %s", part)
}

var BindTimeSub = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
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
	return TIME_SUB(t, num, part)
})
