package security

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func SESSION_USER() (value.Value, error) {
	return value.StringValue("dummy"), nil
}
