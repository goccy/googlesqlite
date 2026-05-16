package json

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_TYPE(v value.JsonValue) (value.Value, error) {
	return value.StringValue(v.Type()), nil
}

func BindJsonType(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("JSON_TYPE: invalid number of arguments: got %d, want 1", len(args))
	}
	value, ok := args[0].(value.JsonValue)
	if !ok {
		return nil, fmt.Errorf("JSON_TYPE: failed to convert %T to JSON value", args[0])
	}
	return JSON_TYPE(value)
}
