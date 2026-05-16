package datetime

import (
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func CURRENT_DATETIME_WITH_TIME(v time.Time) (value.Value, error) {
	return value.DatetimeValue(v), nil
}
