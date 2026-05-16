package operator

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func IS_DISTINCT_FROM(a, b value.Value) (value.Value, error) {
	if a == nil || b == nil {
		eq := a == nil && b == nil
		return value.BoolValue(!eq), nil
	}
	cond, err := a.EQ(b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(!cond), nil
}

// BindIsDistinctFrom observes NULL itself, so it must use the
// KeepNull variant.
var BindIsDistinctFrom = helper.Scalar2KeepNull(IS_DISTINCT_FROM)
