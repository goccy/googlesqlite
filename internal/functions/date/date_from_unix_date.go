package date

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATE_FROM_UNIX_DATE(unixdate int64) (value.Value, error) {
	t := time.Unix(int64(time.Duration(unixdate)*24*time.Hour/time.Second), 0)
	return value.DateValue(t), nil
}

var BindDateFromUnixDate = helper.Scalar1(func(a value.Value) (value.Value, error) {
	unixdate, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return DATE_FROM_UNIX_DATE(unixdate)
})
