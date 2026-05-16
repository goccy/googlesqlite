package interval

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func JUSTIFY_HOURS(v *value.IntervalValue) (value.Value, error) {
	if v.Seconds > 59 {
		v.Minutes += v.Seconds / 60
		v.Seconds %= 60
	} else if v.Seconds < -59 {
		v.Minutes += v.Seconds / 60
		v.Seconds %= 60
	}
	if v.Minutes > 59 {
		v.Hours += v.Minutes / 60
		v.Minutes %= 60
	} else if v.Minutes < -59 {
		v.Hours += v.Hours / 60
		v.Minutes %= 60
	}
	if v.Hours > 23 {
		v.Days += v.Hours / 24
		v.Hours %= 24
	} else if v.Hours < -23 {
		v.Days += v.Hours / 24
		v.Hours %= 24
	}
	return v, nil
}

func BindJustifyHours(args ...value.Value) (value.Value, error) {
	interval, ok := args[0].(*value.IntervalValue)
	if !ok {
		return nil, fmt.Errorf("JUSTIFY_HOURS: unexpected argument type %T", args[0])
	}
	return JUSTIFY_HOURS(interval)
}
