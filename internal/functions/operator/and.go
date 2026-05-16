package operator

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func AND(args ...value.Value) (value.Value, error) {
	for _, v := range args {
		if v == nil {
			continue
		}
		cond, err := v.ToBool()
		if err != nil {
			return nil, err
		}
		if !cond {
			return value.BoolValue(false), nil
		}
	}
	// if exists null value and not exists false value, returns null.
	if helper.ExistsNull(args) {
		return nil, nil
	}
	return value.BoolValue(true), nil
}

// BindAnd observes NULL itself for three-valued logic, so it must
// use the KeepNull variant.
var BindAnd = helper.ScalarNKeepNull(AND)
