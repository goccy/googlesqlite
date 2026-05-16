package array

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func generateArray(start, end, step value.Value) (value.Value, error) {
	if start == nil || end == nil || step == nil {
		return nil, nil
	}
	isLT, err := start.LTE(end)
	if err != nil {
		return nil, err
	}
	arr := &value.ArrayValue{}
	isPositiveStepValue, err := step.GT(value.IntValue(0))
	if err != nil {
		return nil, err
	}
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
		after, err := cur.Add(step)
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
