package json

import (
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/value"
)

// JSON_ARRAY builds a JSON array containing the supplied values
// (one element per argument). NULL arguments become JSON null.
func JSON_ARRAY(args ...value.Value) (value.Value, error) {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if a == nil {
			parts = append(parts, "null")
			continue
		}
		s, err := a.ToJSON()
		if err != nil {
			return nil, err
		}
		parts = append(parts, s)
	}
	return value.JsonValue("[" + strings.Join(parts, ",") + "]"), nil
}

// JSON_OBJECT(key, value, ...) builds a JSON object from
// alternating key/value pairs. Keys must be STRING, values are
// converted via ToJSON.
func JSON_OBJECT(args ...value.Value) (value.Value, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("JSON_OBJECT: needs an even argument count, got %d", len(args))
	}
	parts := make([]string, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, err := args[i].ToString()
		if err != nil {
			return nil, err
		}
		var valStr string
		if args[i+1] == nil {
			valStr = "null"
		} else {
			s, err := args[i+1].ToJSON()
			if err != nil {
				return nil, err
			}
			valStr = s
		}
		parts = append(parts, fmt.Sprintf("%q:%s", key, valStr))
	}
	return value.JsonValue("{" + strings.Join(parts, ",") + "}"), nil
}
