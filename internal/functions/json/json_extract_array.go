package json

import (
	"bytes"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_EXTRACT_ARRAY(v, path string) (value.Value, error) {
	p, err := json.CreatePath(path)
	if err != nil {
		return nil, err
	}
	extracted, err := p.Extract([]byte(v))
	if err != nil {
		return nil, err
	}
	if len(extracted) == 0 {
		return nil, nil
	}
	content := bytes.TrimLeft(extracted[0], " ")
	if len(content) != 0 && content[0] != '[' {
		// not array content
		return nil, nil
	}
	var values []json.RawMessage
	if err := json.Unmarshal(content, &values); err != nil {
		return nil, err
	}
	ret := &value.ArrayValue{}
	for _, val := range values {
		jsonValue := string(val)
		if jsonValue == "null" {
			ret.Values = append(ret.Values, nil)
		} else {
			ret.Values = append(ret.Values, value.JsonValue(jsonValue))
		}
	}
	return ret, nil
}

var BindJsonExtractArray = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	path, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_EXTRACT_ARRAY(v, path)
})
