package array

import (
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/value"
)

func ARRAY_TO_STRING(arr *value.ArrayValue, delim string, nullText ...string) (value.Value, error) {
	var elems []string
	for _, v := range arr.Values {
		if v == nil {
			if len(nullText) == 0 {
				continue
			} else {
				elems = append(elems, nullText[0])
			}
		} else {
			elems = append(elems, v.Format('t'))
		}
	}
	return value.StringValue(strings.Join(elems, delim)), nil
}

func BindArrayToString(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ARRAY_TO_STRING: invalid number of arguments: got %d, want at least 2", len(args))
	}
	arr, err := args[0].ToArray()
	if err != nil {
		return nil, err
	}
	delim, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	if len(args) == 3 {
		nullText, err := args[2].ToString()
		if err != nil {
			return nil, err
		}
		return ARRAY_TO_STRING(arr, delim, nullText)
	}
	return ARRAY_TO_STRING(arr, delim)
}
