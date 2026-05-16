package time

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIME(args ...value.Value) (value.Value, error) {
	if len(args) == 3 {
		hour, err := args[0].ToInt64()
		if err != nil {
			return nil, err
		}
		min, err := args[1].ToInt64()
		if err != nil {
			return nil, err
		}
		sec, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		loc, err := value.ToLocation("")
		if err != nil {
			return nil, err
		}
		return value.TimeValue(time.Date(0, 0, 0, int(hour), int(min), int(sec), 0, loc)), nil
	}
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("TIME: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	switch args[0].(type) {
	case value.TimestampValue:
		t, err := args[0].ToTime()
		if err != nil {
			return nil, err
		}
		if len(args) == 2 {
			zone, err := args[1].ToString()
			if err != nil {
				return nil, err
			}
			loc, err := value.ToLocation(zone)
			if err != nil {
				return nil, err
			}
			return value.TimeValue(t.In(loc)), nil
		}
		return value.TimeValue(t), nil
	case value.DatetimeValue:
		t, err := args[0].ToTime()
		if err != nil {
			return nil, err
		}
		return value.TimeValue(t), nil
	}
	return nil, fmt.Errorf("TIME: invalid first argument type %T", args[0])
}

// BindTime short-circuits to NULL when any argument is NULL; TIME
// itself performs the arity dispatch.
var BindTime = helper.ScalarN(TIME)
