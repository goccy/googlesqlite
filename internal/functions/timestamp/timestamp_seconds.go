package timestamp

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIMESTAMP_SECONDS(sec int64) (value.Value, error) {
	return value.TimestampValue(time.Unix(sec, 0)), nil
}

var BindTimestampSeconds = helper.Scalar1(func(a value.Value) (value.Value, error) {
	sec, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return TIMESTAMP_SECONDS(sec)
})
