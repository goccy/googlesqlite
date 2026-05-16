package timestamp

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIMESTAMP_MILLIS(sec int64) (value.Value, error) {
	return value.TimestampValue(time.UnixMicro(sec * 1000)), nil
}

var BindTimestampMillis = helper.Scalar1(func(a value.Value) (value.Value, error) {
	millisec, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return TIMESTAMP_MILLIS(millisec)
})
