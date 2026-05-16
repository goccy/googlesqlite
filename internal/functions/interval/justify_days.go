package interval

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func JUSTIFY_DAYS(v *value.IntervalValue) (value.Value, error) {
	if v.Days > 29 {
		v.Months += v.Days / 30
		v.Days %= 30
	} else if v.Days < -29 {
		v.Months += v.Days / 30
		v.Days %= 30
	}
	if v.Months > 11 {
		v.Months -= 12
		v.Years++
	} else if v.Months < -11 {
		v.Months += 12
		v.Years--
	}
	return v, nil
}

func BindJustifyDays(args ...value.Value) (value.Value, error) {
	interval, ok := args[0].(*value.IntervalValue)
	if !ok {
		return nil, fmt.Errorf("JUSTIFY_DAYS: unexpected argument type %T", args[0])
	}
	return JUSTIFY_DAYS(interval)
}
