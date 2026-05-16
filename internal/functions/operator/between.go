package operator

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func BETWEEN(target, start, end value.Value) (value.Value, error) {
	greaterThanStart, err := target.GTE(start)
	if err != nil {
		return nil, err
	}
	lessThanEnd, err := target.LTE(end)
	if err != nil {
		return nil, err
	}

	return value.BoolValue(greaterThanStart && lessThanEnd), nil
}

// BindBetween short-circuits to NULL when any operand is NULL, per
// GoogleSQL three-valued logic (NULL, not FALSE).
var BindBetween = helper.Scalar3(BETWEEN)
