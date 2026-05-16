package json

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_VALUE(v, path string) (value.Value, error) {
	p, err := json.CreatePath(path)
	if err != nil {
		return nil, err
	}
	if p.UsedSingleQuotePathSelector() {
		return nil, fmt.Errorf("JSON_VALUE: doesn't use single quote path selector")
	}
	var values []any
	if err := p.Unmarshal([]byte(v), &values); err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	val := values[0]
	if !reflect.ValueOf(val).IsValid() {
		return nil, nil
	}
	switch reflect.ValueOf(val).Type().Kind() {
	case reflect.Map, reflect.Slice:
		return nil, nil
	}
	return value.StringValue(fmt.Sprint(val)), nil
}

var BindJsonValue = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	path, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_VALUE(v, path)
})
