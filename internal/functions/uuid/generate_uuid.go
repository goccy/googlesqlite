package uuid

import (
	"github.com/goccy/googlesqlite/internal/value"
	"github.com/google/uuid"
)

func GENERATE_UUID() (value.Value, error) {
	id := uuid.NewString()
	return value.StringValue(id), nil
}
