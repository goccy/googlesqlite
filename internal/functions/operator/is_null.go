package operator

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func IS_NULL(a value.Value) (value.Value, error) {
	return value.BoolValue(a == nil), nil
}

// BindIsNull observes NULL itself, so it must use the KeepNull
// variant.
var BindIsNull = helper.Scalar1KeepNull(IS_NULL)
