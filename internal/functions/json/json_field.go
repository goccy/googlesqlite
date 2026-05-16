package json

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_FIELD(v, fieldName string) (value.Value, error) {
	p, err := json.CreatePath(fmt.Sprintf(`$.%q`, fieldName))
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
	return value.JsonValue(string(extracted[0])), nil
}

var BindJsonField = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	jsonValue, err := a.ToString()
	if err != nil {
		return nil, err
	}
	fieldName, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_FIELD(jsonValue, fieldName)
})
