package timestamp

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func GENERATE_TIMESTAMP_ARRAY(start, end value.Value, step int64, part string) (value.Value, error) {
	if start == nil || end == nil || step == 0 {
		return nil, nil
	}
	isLT, err := start.LTE(end)
	if err != nil {
		return nil, err
	}
	arr := &value.ArrayValue{}
	isPositiveStepValue := step > 0
	if isLT && !isPositiveStepValue {
		// start less than end and step is negative value
		return arr, nil
	} else if !isLT && isPositiveStepValue {
		// start greater than end and step is positive value
		return arr, nil
	}
	cur := start
	for {
		arr.Values = append(arr.Values, cur)
		after, err := cur.(value.TimestampValue).AddValueWithPart(step, part)
		if err != nil {
			return nil, err
		}
		if isLT {
			cond, err := after.LTE(end)
			if err != nil {
				return nil, err
			}
			if !cond {
				break
			}
		} else {
			cond, err := after.GTE(end)
			if err != nil {
				return nil, err
			}
			if !cond {
				break
			}
		}
		cur = after
	}
	return arr, nil
}

func BindGenerateTimestampArray(args ...value.Value) (value.Value, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("GENERATE_TIMESTAMP_ARRAY: invalid number of arguments: got %d, want 4", len(args))
	}
	step, err := args[2].ToInt64()
	if err != nil {
		return nil, err
	}
	part, err := args[3].ToString()
	if err != nil {
		return nil, err
	}
	return GENERATE_TIMESTAMP_ARRAY(args[0], args[1], step, part)
}
