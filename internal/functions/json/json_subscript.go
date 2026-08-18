package json

import (
	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_SUBSCRIPT(v string, field value.Value) (value.Value, error) {
	var raw json.RawMessage
	switch f := field.(type) {
	case value.IntValue:
		var elems []json.RawMessage
		if err := json.Unmarshal([]byte(v), &elems); err != nil || f < 0 || int(f) >= len(elems) {
			return nil, nil
		}
		raw = elems[f]
	case value.StringValue:
		var members map[string]json.RawMessage
		if err := json.Unmarshal([]byte(v), &members); err != nil {
			return nil, nil
		}
		var ok bool
		if raw, ok = members[string(f)]; !ok {
			return nil, nil
		}
	default:
		return nil, nil
	}
	return value.JsonValue(string(raw)), nil
}

var BindSubscript = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	jsonValue, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_SUBSCRIPT(jsonValue, b)
})
