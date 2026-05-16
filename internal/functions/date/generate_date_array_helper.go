package date

import (
	"github.com/goccy/googlesqlite/internal/value"
)

// generateDateArray builds a DATE array between [start, end] stepped by
// `step interval` (e.g. step=1, interval="DAY"). Used by
// GENERATE_DATE_ARRAY.
func generateDateArray(start, end value.Value, step int, interval string) (value.Value, error) {
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
		return arr, nil
	} else if !isLT && isPositiveStepValue {
		return arr, nil
	}
	cur := start
	for {
		arr.Values = append(arr.Values, cur)
		after, err := cur.(value.DateValue).AddDateWithInterval(step, interval)
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
