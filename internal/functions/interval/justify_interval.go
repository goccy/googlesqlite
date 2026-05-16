package interval

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func JUSTIFY_INTERVAL(v *value.IntervalValue) (value.Value, error) {
	if _, err := JUSTIFY_HOURS(v); err != nil {
		return nil, err
	}
	return JUSTIFY_DAYS(v)
}

func BindJustifyInterval(args ...value.Value) (value.Value, error) {
	interval, ok := args[0].(*value.IntervalValue)
	if !ok {
		return nil, fmt.Errorf("JUSTIFY_INTERVAL: unexpected argument type %T", args[0])
	}
	return JUSTIFY_INTERVAL(interval)
}
