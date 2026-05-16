package json

import (
	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func PARSE_JSON(expr, mode string) (value.Value, error) {
	var v any
	if err := json.Unmarshal([]byte(expr), &v); err != nil {
		return nil, err
	}
	return value.JsonValue(expr), nil
}

var BindParseJson = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	mode, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return PARSE_JSON(v, mode)
})
