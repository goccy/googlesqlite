package date

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func UNIX_DATE(t time.Time) (value.Value, error) {
	return value.IntValue(t.Unix() / int64(24*time.Hour/time.Second)), nil
}

var BindUnixDate = helper.Scalar1(func(a value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	return UNIX_DATE(t)
})
