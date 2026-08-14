package json

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func queryJSONText(name, doc, path string, allowSingleQuotes bool) (text string, found bool, err error) {
	p, err := json.CreatePath(path)
	if err != nil {
		return "", false, err
	}
	if !allowSingleQuotes && p.UsedSingleQuotePathSelector() {
		return "", false, fmt.Errorf("%s: doesn't use single quote path selector", name)
	}
	extracted, err := p.Extract([]byte(doc))
	if err != nil {
		return "", false, err
	}
	if len(extracted) == 0 {
		return "", false, nil
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, extracted[0]); err != nil {
		return "", false, fmt.Errorf("failed to format json %q: %w", extracted[0], err)
	}
	return buf.String(), true, nil
}

// json_functions.md#json_query "Return type": results are typed like the input;
// a JSON null is SQL NULL only for a STRING input, and never for array elements.
func scalarLikeInput(input value.Value, text string) value.Value {
	if _, ok := input.(value.JsonValue); ok {
		return value.JsonValue(text)
	}
	if text == "null" {
		return nil
	}
	return value.StringValue(text)
}

func elementLikeInput(input value.Value, text string) value.Value {
	if _, ok := input.(value.JsonValue); ok {
		return value.JsonValue(text)
	}
	return value.StringValue(text)
}

func JSON_QUERY(input value.Value, path string) (value.Value, error) {
	doc, err := input.ToString()
	if err != nil {
		return nil, err
	}
	text, found, err := queryJSONText("JSON_QUERY", doc, path, false)
	if err != nil || !found {
		return nil, err
	}
	return scalarLikeInput(input, text), nil
}

var BindJsonQuery = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	path, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_QUERY(a, path)
})
