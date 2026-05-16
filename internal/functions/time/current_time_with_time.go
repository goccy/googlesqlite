package time

import (
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func CURRENT_TIME_WITH_TIME(v time.Time) (value.Value, error) {
	return value.TimeValue(v), nil
}
