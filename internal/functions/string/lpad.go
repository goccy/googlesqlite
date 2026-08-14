package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LPAD(originalValue value.Value, returnLength int64, pattern value.Value) (value.Value, error) {
	if returnLength < 0 {
		return nil, fmt.Errorf("LPAD: second argument (output size) cannot be negative")
	}
	retLen, err := helper.SafeInt(returnLength)
	if err != nil {
		return nil, err
	}
	switch originalValue.(type) {
	case value.StringValue:
		s, err := originalValue.ToString()
		if err != nil {
			return nil, err
		}
		var p string
		if pattern != nil {
			p, err = pattern.ToString()
			if err != nil {
				return nil, err
			}
			if p == "" {
				return nil, fmt.Errorf("LPAD: third argument (pad pattern) cannot be empty")
			}
		}
		runes := []rune(s)
		if len(runes) >= retLen {
			return value.StringValue(string(runes[:retLen])), nil
		}
		remainLen := retLen - len(runes)
		var pat []rune
		if pattern == nil {
			pat = []rune(strings.Repeat(" ", remainLen))
		} else {
			pat = []rune(p)
			if remainLen-len(pat) > 0 {
				// needs to repeat pattern
				repeatNum := ((remainLen - len(pat)) / len(pat)) + 2
				pat = []rune(strings.Repeat(string(pat), repeatNum))
			}
		}
		return value.StringValue(string(pat[:remainLen]) + s), nil
	case value.BytesValue:
		b, err := originalValue.ToBytes()
		if err != nil {
			return nil, err
		}
		var p []byte
		if pattern != nil {
			p, err = pattern.ToBytes()
			if err != nil {
				return nil, err
			}
			if len(p) == 0 {
				return nil, fmt.Errorf("LPAD: third argument (pad pattern) cannot be empty")
			}
		}
		if len(b) >= retLen {
			return value.BytesValue(b[:retLen]), nil
		}
		remainLen := retLen - len(b)
		var pat []byte
		if pattern == nil {
			pat = bytes.Repeat([]byte{' '}, remainLen)
		} else {
			pat = p
			if remainLen-len(p) > 0 {
				// needs to repeat pattern
				repeatNum := ((remainLen - len(p)) / len(p)) + 2
				pat = bytes.Repeat(p, repeatNum)
			}
		}
		out := make([]byte, 0, retLen)
		out = append(out, pat[:remainLen]...)
		return value.BytesValue(append(out, b...)), nil
	}
	return nil, fmt.Errorf("LPAD: original value type is must be STRING or BYTES type")
}

var BindLpad = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, fmt.Errorf("LPAD: invalid number of arguments: got %d, want 2 or 3", len(args))
	}
	var pattern value.Value
	if len(args) == 3 {
		pattern = args[2]
	}
	length, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	return LPAD(args[0], length, pattern)
})
