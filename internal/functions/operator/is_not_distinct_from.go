package operator

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func IS_NOT_DISTINCT_FROM(a, b value.Value) (value.Value, error) {
	if a == nil || b == nil {
		return value.BoolValue(a == nil && b == nil), nil
	}
	cond, err := a.EQ(b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(cond), nil
}

// BindIsNotDistinctFrom observes NULL itself, so it must use the
// KeepNull variant.
var BindIsNotDistinctFrom = helper.Scalar2KeepNull(IS_NOT_DISTINCT_FROM)
