package json

import (
	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_QUERY_ARRAY(input value.Value, path string) (value.Value, error) {
	doc, err := input.ToString()
	if err != nil {
		return nil, err
	}
	text, found, err := queryJSONText("JSON_QUERY_ARRAY", doc, path, false)
	if err != nil || !found {
		return nil, err
	}
	if len(text) == 0 || text[0] != '[' {
		return nil, nil
	}
	var elems []json.RawMessage
	if err := json.Unmarshal([]byte(text), &elems); err != nil {
		return nil, err
	}
	ret := &value.ArrayValue{}
	for _, e := range elems {
		ret.Values = append(ret.Values, elementLikeInput(input, string(e)))
	}
	return ret, nil
}

var BindJsonQueryArray = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	path, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_QUERY_ARRAY(a, path)
})
