package array

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func MAKE_ARRAY(args ...value.Value) (value.Value, error) {
	return &value.ArrayValue{Values: args}, nil
}

// BindMakeArray builds an array literal. NULL arguments are kept as
// NULL elements, so the call site must not short-circuit on NULL.
var BindMakeArray = helper.ScalarNKeepNull(MAKE_ARRAY)
