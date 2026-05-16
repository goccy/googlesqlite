package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func STARTS_WITH(val, starts value.Value) (value.Value, error) {
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		s, err := starts.ToString()
		if err != nil {
			return nil, err
		}
		return value.BoolValue(strings.HasPrefix(v, s)), nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		s, err := starts.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.BoolValue(bytes.HasPrefix(v, s)), nil
	}
	return nil, fmt.Errorf("ENDS_WITH: argument type must be STRING or BYTES")
}

var BindStartsWith = helper.Scalar2(STARTS_WITH)
