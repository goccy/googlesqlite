package date

import (
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func CURRENT_DATE_WITH_TIME(v time.Time) (value.Value, error) {
	return value.DateValue(v), nil
}
