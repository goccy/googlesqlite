package json

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TO_JSON(v value.Value, stringifyWideNumbers bool) (value.Value, error) {
	s, err := v.ToJSON()
	if err != nil {
		return nil, err
	}
	return value.JsonValue(s), nil
}

func BindToJson(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("TO_JSON: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	var stringifyWideNumbers bool
	if len(args) == 2 {
		b, err := args[1].ToBool()
		if err != nil {
			return nil, err
		}
		stringifyWideNumbers = b
	}
	return TO_JSON(args[0], stringifyWideNumbers)
}
