package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func REPEAT(originalValue value.Value, repetitions int64) (value.Value, error) {
	if repetitions < 0 {
		return nil, fmt.Errorf("REPEAT: second argument (repeat count) cannot be negative")
	}
	switch originalValue.(type) {
	case value.StringValue:
		v, err := originalValue.ToString()
		if err != nil {
			return nil, err
		}
		// An empty value repeats to empty whatever the count is, so the
		// count is never sized against the machine's int.
		if v == "" {
			return value.StringValue(""), nil
		}
		reps, err := helper.SafeInt(repetitions)
		if err != nil {
			return nil, err
		}
		return value.StringValue(strings.Repeat(v, reps)), nil
	case value.BytesValue:
		v, err := originalValue.ToBytes()
		if err != nil {
			return nil, err
		}
		if len(v) == 0 {
			return value.BytesValue(nil), nil
		}
		reps, err := helper.SafeInt(repetitions)
		if err != nil {
			return nil, err
		}
		return value.BytesValue(bytes.Repeat(v, reps)), nil
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
