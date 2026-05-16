package json

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func TO_JSON_STRING(v value.Value, prettyPrint bool) (value.Value, error) {
	if v == nil {
		// BigQuery surfaces TO_JSON_STRING(NULL) as the JSON null
		// literal rather than SQL NULL.
		return value.StringValue("null"), nil
	}
	s, err := v.ToJSON()
	if err != nil {
		return nil, err
	}
	return value.StringValue(s), nil
}

func BindToJsonString(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("TO_JSON_STRING: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	var prettyPrint bool
	if len(args) == 2 {
		if args[1] != nil {
			b, err := args[1].ToBool()
			if err != nil {
				return nil, err
			}
			prettyPrint = b
		}
	}
	return TO_JSON_STRING(args[0], prettyPrint)
}
