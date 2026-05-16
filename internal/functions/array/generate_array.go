package array

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func GENERATE_ARRAY(start, end value.Value, step ...value.Value) (value.Value, error) {
	var stepValue value.Value
	if len(step) > 0 {
		stepValue = step[0]
	} else {
		stepValue = value.IntValue(1)
	}
	return generateArray(start, end, stepValue)
}

func BindGenerateArray(args ...value.Value) (value.Value, error) {
	if len(args) != 3 && len(args) != 2 {
		return nil, fmt.Errorf("GENERATE_ARRAY: invalid number of arguments: got %d, want 3 or 2", len(args))
	}
	if len(args) == 3 {
		return GENERATE_ARRAY(args[0], args[1], args[2])
	}
	return GENERATE_ARRAY(args[0], args[1])
}
