package date

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATE_SUB(t time.Time, v int64, part string) (value.Value, error) {
	switch part {
	case "DAY":
		return value.DateValue(t.AddDate(0, 0, int(-v))), nil
	case "WEEK":
		return value.DateValue(t.AddDate(0, 0, int(-v*7))), nil
	case "MONTH":
		return value.DateValue(helper.AddMonth(t, int(-v))), nil
	case "YEAR":
		return value.DateValue(helper.AddYear(t, int(-v))), nil
	}
	return nil, fmt.Errorf("unexpected part value %s", part)
}

var BindDateSub = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
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
	return DATE_SUB(t, num, part)
})
