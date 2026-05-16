package date

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATE_ADD(t time.Time, v int64, part string) (value.Value, error) {
	switch part {
	case "DAY":
		n, err := helper.SafeInt(v)
		if err != nil {
			return nil, err
		}
		return value.DateValue(t.AddDate(0, 0, n)), nil
	case "WEEK":
		n, err := helper.SafeInt(v * 7)
		if err != nil {
			return nil, err
		}
		return value.DateValue(t.AddDate(0, 0, n)), nil
	case "MONTH":
		n, err := helper.SafeInt(v)
		if err != nil {
			return nil, err
		}
		return value.DateValue(helper.AddMonth(t, n)), nil
	case "YEAR":
		n, err := helper.SafeInt(v)
		if err != nil {
			return nil, err
		}
		return value.DateValue(helper.AddYear(t, n)), nil
	case "QUARTER":
		return value.DateValue(helper.AddMonth(t, 3)), nil
	}
	return nil, fmt.Errorf("unexpected part value %s", part)
}

var BindDateAdd = helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
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
	return DATE_ADD(t, num, part)
})
