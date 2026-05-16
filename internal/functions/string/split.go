package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func SPLIT(val, delimValue value.Value) (value.Value, error) {
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		var delim = ","
		if delimValue != nil {
			delimV, err := delimValue.ToString()
			if err != nil {
				return nil, err
			}
			delim = delimV
		}
		ret := &value.ArrayValue{}
		for splitted := range strings.SplitSeq(v, delim) {
			ret.Values = append(ret.Values, value.StringValue(splitted))
		}
		return ret, nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		if delimValue == nil {
			return nil, fmt.Errorf("SPLIT: delimiter must be specified for bytes val")
		}
		delim, err := delimValue.ToBytes()
		if err != nil {
			return nil, err
		}
		ret := &value.ArrayValue{}
		for splitted := range bytes.SplitSeq(v, delim) {
			ret.Values = append(ret.Values, value.BytesValue(splitted))
		}
		return ret, nil
	}
	return nil, fmt.Errorf("SPLIT: val must be STRING or BYTES")
}

func BindSplit(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return &value.ArrayValue{}, nil
	}
	var delim value.Value
	if len(args) > 1 {
		delim = args[1]
	}
	return SPLIT(args[0], delim)
}
