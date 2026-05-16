package string

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// LTRIM trims characters from the left. cutset == "" means: strip the
// full Unicode whitespace class for STRING (ASCII whitespace for BYTES).
func LTRIM(v value.Value, cutset string) (value.Value, error) {
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		if cutset == "" {
			return value.StringValue(strings.TrimLeftFunc(s, unicode.IsSpace)), nil
		}
		return value.StringValue(strings.TrimLeft(s, cutset)), nil
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		if cutset == "" {
			return value.BytesValue(bytes.TrimLeftFunc(b, asciiSpace)), nil
		}
		return value.BytesValue(bytes.TrimLeft(b, cutset)), nil
	}
	return nil, fmt.Errorf("LTRIM: value type is must be STRING or BYTES type")
}

var BindLtrim = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("LTRIM: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	cutset := ""
	if len(args) == 2 {
		v, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		cutset = v
	}
	return LTRIM(args[0], cutset)
})
