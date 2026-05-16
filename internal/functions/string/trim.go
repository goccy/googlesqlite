package string

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TRIM(v, cutsetV value.Value) (value.Value, error) {
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		if cutsetV == nil {
			// BigQuery: strip the full Unicode whitespace class.
			return value.StringValue(strings.TrimFunc(s, unicode.IsSpace)), nil
		}
		cutset, err := cutsetV.ToString()
		if err != nil {
			return nil, err
		}
		return value.StringValue(strings.Trim(s, cutset)), nil
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		if cutsetV == nil {
			// BYTES variant has no Unicode notion: fall back to ASCII whitespace.
			return value.BytesValue(bytes.TrimFunc(b, asciiSpace)), nil
		}
		cb, err := cutsetV.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.BytesValue(bytes.Trim(b, string(cb))), nil
	}
	return nil, fmt.Errorf("TRIM: expression type is must be STRING or BYTES type")
}

func asciiSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}

var BindTrim = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("TRIM: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	if len(args) == 2 {
		return TRIM(args[0], args[1])
	}
	return TRIM(args[0], nil)
})
