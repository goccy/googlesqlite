package date

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func FORMAT_DATE(format string, t time.Time) (value.Value, error) {
	s, err := helper.FormatTime(format, &t, helper.FormatTypeDate)
	if err != nil {
		return nil, err
	}
	return value.StringValue(s), nil
}

var BindFormatDate = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	format, err := a.ToString()
	if err != nil {
		return nil, err
	}
	t, err := b.ToTime()
	if err != nil {
		return nil, err
	}
	return FORMAT_DATE(format, t)
})
