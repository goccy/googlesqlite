package timestamp

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/date"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIMESTAMP_DIFF(a, b time.Time, part string) (value.Value, error) {
	diff := a.Sub(b)
	switch part {
	case "MICROSECOND":
		return value.IntValue(diff / time.Microsecond), nil
	case "MILLISECOND":
		return value.IntValue(diff / time.Millisecond), nil
	case "SECOND":
		return value.IntValue(diff / time.Second), nil
	case "MINUTE":
		return value.IntValue(diff / time.Minute), nil
	case "HOUR":
		return value.IntValue(diff / time.Hour), nil
	default:
		dateDiff, err := date.DATE_DIFF(a, b, part)
		if err != nil {
			return nil, fmt.Errorf("TIMESTAMP_DIFF: %w", err)
		}

		return dateDiff, nil
	}
}

var BindTimestampDiff = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
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
	return TIMESTAMP_DIFF(t, t2, part)
})
