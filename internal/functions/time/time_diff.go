package time

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIME_DIFF(a, b time.Time, part string) (value.Value, error) {
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
	}
	return nil, fmt.Errorf("TIME_DIFF: unexpected part value %s", part)
}

var BindTimeDiff = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
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
	return TIME_DIFF(t, t2, part)
})
