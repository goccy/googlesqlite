package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/value"
)

func REPLACE(originalValue, fromValue, toValue value.Value) (value.Value, error) {
	switch originalValue.(type) {
	case value.StringValue:
		v, err := originalValue.ToString()
		if err != nil {
			return nil, err
		}
		from, err := fromValue.ToString()
		if err != nil {
			return nil, err
		}
		to, err := toValue.ToString()
		if err != nil {
			return nil, err
		}
		return value.StringValue(strings.ReplaceAll(v, from, to)), nil
	case value.BytesValue:
		v, err := originalValue.ToBytes()
		if err != nil {
			return nil, err
		}
		from, err := fromValue.ToBytes()
		if err != nil {
			return nil, err
		}
		to, err := toValue.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.BytesValue(bytes.ReplaceAll(v, from, to)), nil
	}
	return nil, fmt.Errorf("REPLACE: originalValue must be STRING or BYTES")
}
