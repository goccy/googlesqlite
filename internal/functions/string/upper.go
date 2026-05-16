package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func UPPER(v value.Value) (value.Value, error) {
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		return value.StringValue(strings.ToUpper(s)), nil
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.BytesValue(bytes.ToUpper(b)), nil
	}
	return nil, fmt.Errorf("UPPER: value type is must be STRING or BYTES type")
}

var BindUpper = helper.Scalar1(UPPER)
