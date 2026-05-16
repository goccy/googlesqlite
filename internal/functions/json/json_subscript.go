package json

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_SUBSCRIPT(v string, field value.Value) (value.Value, error) {
	var path *json.Path
	switch field.(type) {
	case value.IntValue:
		index, err := field.ToInt64()
		if err != nil {
			return nil, err
		}
		p, err := json.CreatePath(fmt.Sprintf(`$[%d]`, index))
		if err != nil {
			return nil, err
		}
		path = p
	case value.StringValue:
		name, err := field.ToString()
		if err != nil {
			return nil, err
		}
		p, err := json.CreatePath(fmt.Sprintf(`$.%q`, name))
		if err != nil {
			return nil, err
		}
		path = p
	}
	extracted, err := path.Extract([]byte(v))
	if err != nil {
		return nil, err
	}
	if len(extracted) == 0 {
		return nil, nil
	}
	return value.JsonValue(string(extracted[0])), nil
}

var BindSubscript = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	jsonValue, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_SUBSCRIPT(jsonValue, b)
})
