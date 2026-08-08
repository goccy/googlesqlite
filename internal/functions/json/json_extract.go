package json

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_EXTRACT(input value.Value, path string) (value.Value, error) {
	doc, err := input.ToString()
	if err != nil {
		return nil, err
	}
	text, found, err := queryJSONText("JSON_EXTRACT", doc, path, true)
	if err != nil || !found {
		return nil, err
	}
	return scalarLikeInput(input, text), nil
}

var BindJsonExtract = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	path, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_EXTRACT(a, path)
})
