package json

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func JSON_QUERY(v, path string) (value.Value, error) {
	p, err := json.CreatePath(path)
	if err != nil {
		return nil, err
	}
	if p.UsedSingleQuotePathSelector() {
		return nil, fmt.Errorf("JSON_QUERY: doesn't use single quote path selector")
	}
	extracted, err := p.Extract([]byte(v))
	if err != nil {
		return nil, err
	}
	if len(extracted) == 0 {
		return nil, nil
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, extracted[0]); err != nil {
		return nil, fmt.Errorf("failed to format json %q: %w", extracted[0], err)
	}
	jsonValue := buf.String()
	if jsonValue == "null" {
		return nil, nil
	}
	return value.JsonValue(jsonValue), nil
}

var BindJsonQuery = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	path, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return JSON_QUERY(v, path)
})
