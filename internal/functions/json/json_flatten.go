package json

import (
	"fmt"

	gjson "github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"
)

// JSON_FLATTEN flattens nested JSON arrays into an
// ARRAY<JSON> with elements one level shallower. Given
// `[[1,2],[3,4]]` it returns `[JSON 1, JSON 2, JSON 3, JSON 4]`.
// Non-array inputs return a single-element array containing the
// original JSON.
func JSON_FLATTEN(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("JSON_FLATTEN: invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	body, err := args[0].ToJSON()
	if err != nil {
		return nil, err
	}
	var node any
	if err := gjson.Unmarshal([]byte(body), &node); err != nil {
		return nil, err
	}
	out := &value.ArrayValue{}
	appendOne := func(e any) error {
		b, err := gjson.Marshal(e)
		if err != nil {
			return err
		}
		out.Values = append(out.Values, value.JsonValue(b))
		return nil
	}
	if arr, ok := node.([]any); ok {
		for _, e := range arr {
			if inner, ok := e.([]any); ok {
				for _, ee := range inner {
					if err := appendOne(ee); err != nil {
						return nil, err
					}
				}
				continue
			}
			if err := appendOne(e); err != nil {
				return nil, err
			}
		}
	} else {
		if err := appendOne(node); err != nil {
			return nil, err
		}
	}
	return out, nil
}
