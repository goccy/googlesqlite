package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LPAD(originalValue value.Value, returnLength int64, pattern value.Value) (value.Value, error) {
	switch originalValue.(type) {
	case value.StringValue:
		s, err := originalValue.ToString()
		if err != nil {
			return nil, err
		}
		runes := []rune(s)
		if len(runes) >= int(returnLength) {
			return value.StringValue(string(runes[:returnLength])), nil
		}
		remainLen := int(returnLength) - len(runes)
		var pat []rune
		if pattern == nil {
			pat = []rune(strings.Repeat(" ", remainLen))
		} else {
			p, err := pattern.ToString()
			if err != nil {
				return nil, err
			}
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
		if len(b) >= int(returnLength) {
			return value.BytesValue(b[:returnLength]), nil
		}
		remainLen := int(returnLength) - len(b)
		var pat []byte
		if pattern == nil {
			pat = bytes.Repeat([]byte{' '}, remainLen)
		} else {
			p, err := pattern.ToBytes()
			if err != nil {
				return nil, err
			}
			if remainLen-len(p) > 0 {
				// needs to repeat pattern
				repeatNum := ((remainLen - len(p)) / len(p)) + 2
				pat = bytes.Repeat(p, repeatNum)
			}
		}
		return value.BytesValue(append(pat[:remainLen], b...)), nil
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
