package string

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// RTRIM trims characters from the right. cutset == "" means: strip the
// full Unicode whitespace class for STRING (ASCII whitespace for BYTES).
func RTRIM(val value.Value, cutset string) (value.Value, error) {
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		if cutset == "" {
			return value.StringValue(strings.TrimRightFunc(v, unicode.IsSpace)), nil
		}
		return value.StringValue(strings.TrimRight(v, cutset)), nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		if cutset == "" {
			return value.BytesValue(bytes.TrimRightFunc(v, asciiSpace)), nil
		}
		return value.BytesValue(bytes.TrimRight(v, cutset)), nil
	}
	return nil, fmt.Errorf("RTRIM: value1 must be STRING or BYTES")
}

var BindRtrim = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	var cutset = ""
	if len(args) > 1 {
		v, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		cutset = v
	}
	return RTRIM(args[0], cutset)
})
