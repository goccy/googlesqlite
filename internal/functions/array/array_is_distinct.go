package array

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// ARRAY_IS_DISTINCT returns TRUE when the input array has no
// duplicate elements (comparing by string-form key). NULL inputs
// propagate; NULL elements are treated as their own distinct
// value (a single NULL is fine; two NULLs are duplicates).
func ARRAY_IS_DISTINCT(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ARRAY_IS_DISTINCT: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	arr, err := args[0].ToArray()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(arr.Values))
	nullSeen := false
	for _, v := range arr.Values {
		if v == nil {
			if nullSeen {
				return value.BoolValue(false), nil
			}
			nullSeen = true
			continue
		}
		key, err := v.ToString()
		if err != nil {
			return nil, err
		}
		if _, exists := seen[key]; exists {
			return value.BoolValue(false), nil
		}
		seen[key] = struct{}{}
	}
	return value.BoolValue(true), nil
}
