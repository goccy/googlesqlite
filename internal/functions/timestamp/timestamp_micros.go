package timestamp

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIMESTAMP_MICROS(sec int64) (value.Value, error) {
	return value.TimestampValue(time.UnixMicro(sec)), nil
}

var BindTimestampMicros = helper.Scalar1(func(a value.Value) (value.Value, error) {
	microsec, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return TIMESTAMP_MICROS(microsec)
})
