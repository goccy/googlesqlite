package timestamp

import (
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func CURRENT_TIMESTAMP_WITH_TIME(v time.Time) (value.Value, error) {
	return value.TimestampValue(v), nil
}
