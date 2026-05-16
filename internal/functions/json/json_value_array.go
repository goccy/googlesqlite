package json

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_VALUE_ARRAY(v, path string) (value.Value, error) {
	p, err := json.CreatePath(path)
	if err != nil {
		return nil, err
	}
	if p.UsedSingleQuotePathSelector() {
		return nil, fmt.Errorf("JSON_VALUE_ARRAY: doesn't use single quote path selector")
	}
	var values []any
	if err := p.Unmarshal([]byte(v), &values); err != nil {
		// invalid json content is ignored.
		return nil, nil
	}
	if len(values) == 0 {
		return nil, nil
	}
	val := values[0]
	rv := reflect.ValueOf(val)
	if !rv.IsValid() || rv.Type().Kind() != reflect.Slice {
		return nil, nil
	}
	ret := &value.ArrayValue{}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		elemV := reflect.ValueOf(elem)
		elemKind := elemV.Type().Kind()
		if elemKind == reflect.Map || elemKind == reflect.Slice {
			return nil, nil
		}
		jsonValue := fmt.Sprint(elem)
		if jsonValue == "null" {
			ret.Values = append(ret.Values, nil)
		} else {
			ret.Values = append(ret.Values, value.StringValue(jsonValue))
		}
	}
	return ret, nil
}

var BindJsonValueArray = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	path, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_VALUE_ARRAY(v, path)
})
