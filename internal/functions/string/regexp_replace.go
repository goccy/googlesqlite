package string

import (
	"fmt"
	"regexp"

	"github.com/goccy/googlesqlite/internal/value"
)

func REGEXP_REPLACE(val, exprValue, replacementValue value.Value) (value.Value, error) {
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		expr, err := exprValue.ToString()
		if err != nil {
			return nil, err
		}
		replacement, err := replacementValue.ToString()
		if err != nil {
			return nil, err
		}
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		normalized, err := normalizeReplacement(replacement)
		if err != nil {
			return nil, err
		}
		return value.StringValue(re.ReplaceAllString(v, normalized)), nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		expr, err := exprValue.ToBytes()
		if err != nil {
			return nil, err
		}
		replacement, err := replacementValue.ToBytes()
		if err != nil {
			return nil, err
		}
		re, err := regexp.Compile(string(expr))
		if err != nil {
			return nil, err
		}
		normalized, err := normalizeReplacement(string(replacement))
		if err != nil {
			return nil, err
		}
		return value.BytesValue(re.ReplaceAll(v, []byte(normalized))), nil
	}
	return nil, fmt.Errorf("REGEXP_REPLACE: val must be STRING or BYTES, %s", val)
}
