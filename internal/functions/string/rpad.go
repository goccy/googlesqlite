package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func RPAD(originalValue value.Value, returnLength int64, pattern value.Value) (value.Value, error) {
	if returnLength < 0 {
		return nil, fmt.Errorf("RPAD: second argument (output size) cannot be negative")
	}
	retLen, err := helper.SafeInt(returnLength)
	if err != nil {
		return nil, err
	}
	switch originalValue.(type) {
	case value.StringValue:
		v, err := originalValue.ToString()
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
				return nil, fmt.Errorf("RPAD: third argument (pad pattern) cannot be empty")
			}
		}
		runes := []rune(v)
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
		return value.StringValue(v + string(pat[:remainLen])), nil
	case value.BytesValue:
		v, err := originalValue.ToBytes()
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
				return nil, fmt.Errorf("RPAD: third argument (pad pattern) cannot be empty")
			}
		}
		if len(v) >= retLen {
			return value.BytesValue(v[:retLen]), nil
		}
		remainLen := retLen - len(v)
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
		out = append(out, v...)
		return value.BytesValue(append(out, pat[:remainLen]...)), nil
	}
	return nil, fmt.Errorf("RPAD: originalValue must be STRING or BYTES")
}

var BindRpad = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	length, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	var pat value.Value
	if len(args) > 2 {
		pat = args[2]
	}
	return RPAD(args[0], length, pat)
})
