package json

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// JSON_PATH_EXISTS returns TRUE when the path resolves to any
// value (including null) in the JSON document. Implementation
// reuses the path resolver used by JSON_VALUE / JSON_QUERY.
func JSON_PATH_EXISTS(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("JSON_PATH_EXISTS: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	body, err := args[0].ToJSON()
	if err != nil {
		return nil, err
	}
	path, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	got, err := JSON_QUERY(body, path)
	if err != nil {
		return value.BoolValue(false), nil
	}
	return value.BoolValue(got != nil), nil
}
