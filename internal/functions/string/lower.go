package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LOWER(v value.Value) (value.Value, error) {
	if v == nil {
		return nil, nil
	}
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		return value.StringValue(strings.ToLower(s)), nil
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.BytesValue(bytes.ToLower(b)), nil
	}
	return nil, fmt.Errorf("LOWER: value type is must be STRING or BYTES type")
}

var BindLower = helper.Scalar1(LOWER)
