package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ENDS_WITH(val, ends value.Value) (value.Value, error) {
	switch val.(type) {
	case value.StringValue:
		s, err := val.ToString()
		if err != nil {
			return nil, err
		}
		e, err := ends.ToString()
		if err != nil {
			return nil, err
		}
		return value.BoolValue(strings.HasSuffix(s, e)), nil
	case value.BytesValue:
		b, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		e, err := ends.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.BoolValue(bytes.HasSuffix(b, e)), nil
	}
	return nil, fmt.Errorf("ENDS_WITH: argument type must be STRING or BYTES")
}

var BindEndsWith = helper.Scalar2(ENDS_WITH)
