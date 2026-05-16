package date

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func GENERATE_DATE_ARRAY(start, end value.Value, step ...value.Value) (value.Value, error) {
	if len(step) > 2 {
		return nil, fmt.Errorf("invalid step value %v", step)
	}
	var (
		stepValue int64 = 1
		interval        = "DAY"
	)
	if len(step) == 2 {
		stepV, err := step[0].ToInt64()
		if err != nil {
			return nil, err
		}
		intervalV, err := step[1].ToString()
		if err != nil {
			return nil, err
		}
		stepValue = stepV
		interval = intervalV
	} else if len(step) == 1 {
		stepV, err := step[0].ToInt64()
		if err != nil {
			return nil, err
		}
		stepValue = stepV
	}
	step32, err := helper.SafeInt(stepValue)
	if err != nil {
		return nil, err
	}
	return generateDateArray(start, end, step32, interval)
}

func BindGenerateDateArray(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("GENERATE_DATE_ARRAY: invalid number of arguments: got %d, want at least 2", len(args))
	}
	if len(args) == 2 {
		return GENERATE_DATE_ARRAY(args[0], args[1])
	}
	return GENERATE_DATE_ARRAY(args[0], args[1], args[2:]...)
}
