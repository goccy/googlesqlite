package string

import (
	"fmt"
	"regexp"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func REGEXP_EXTRACT_ALL(val value.Value, expr string) (value.Value, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		matches := re.FindAllStringSubmatch(v, -1)
		ret := &value.ArrayValue{}
		for _, match := range matches {
			ret.Values = append(ret.Values, value.StringValue(match[len(match)-1]))
		}
		return ret, nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		matches := re.FindAllSubmatch(v, -1)
		ret := &value.ArrayValue{}
		for _, match := range matches {
			ret.Values = append(ret.Values, value.BytesValue(match[len(match)-1]))
		}
		return ret, nil
	}
	return nil, fmt.Errorf("REGEXP_EXTRACT_ALL: val argument must be STRING or BYTES")
}

var BindRegexpExtractAll = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	expr, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return REGEXP_EXTRACT_ALL(a, expr)
})
