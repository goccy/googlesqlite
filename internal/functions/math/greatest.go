package math

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func GREATEST(args ...value.Value) (value.Value, error) {
	var max value.Value
	for _, arg := range args {
		if arg == nil {
			return nil, nil
		}
		if max == nil {
			max = arg
			continue
		}
		gt, err := arg.GT(max)
		if err != nil {
			return nil, err
		}
		if gt {
			max = arg
		}
	}
	return max, nil
}

var BindGreatest = helper.ScalarN(GREATEST)
