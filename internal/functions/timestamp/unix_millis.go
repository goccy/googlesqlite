package timestamp

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func UNIX_MILLIS(t time.Time) (value.Value, error) {
	return value.IntValue(t.UnixMilli()), nil
}

var BindUnixMillis = helper.Scalar1(func(a value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	return UNIX_MILLIS(t)
})
