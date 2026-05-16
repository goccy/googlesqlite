package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func REPEAT(originalValue value.Value, repetitions int64) (value.Value, error) {
	switch originalValue.(type) {
	case value.StringValue:
		v, err := originalValue.ToString()
		if err != nil {
			return nil, err
		}
		return value.StringValue(strings.Repeat(v, int(repetitions))), nil
	case value.BytesValue:
		v, err := originalValue.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.BytesValue(bytes.Repeat(v, int(repetitions))), nil
	}
	return nil, fmt.Errorf("REPEAT: originalValue must be STRING or BYTES")
}

var BindRepeat = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	repetitions, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return REPEAT(a, repetitions)
})
