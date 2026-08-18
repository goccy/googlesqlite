package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func FORMAT(format string, args ...value.Value) (value.Value, error) {
	result, isNull, err := parseFormat(format, args...)
	if err != nil || isNull {
		return nil, err
	}
	return value.StringValue(result), nil
}

var BindFormat = helper.ScalarNKeepNull(func(args ...value.Value) (value.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("FORMAT: invalid number of arguments: got %d, want at least 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	format, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	if len(args) > 1 {
		return FORMAT(format, args[1:]...)
	}
	return FORMAT(format)
})
