package math

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LEAST(args ...value.Value) (value.Value, error) {
	var min value.Value
	for _, arg := range args {
		if arg == nil {
			return nil, nil
		}
		if min == nil {
			min = arg
			continue
		}
		less, err := arg.LT(min)
		if err != nil {
			return nil, err
		}
		if less {
			min = arg
		}
	}
	return min, nil
}

var BindLeast = helper.ScalarN(LEAST)
