package conditional

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func IF(cond, trueV, falseV value.Value) (value.Value, error) {
	if cond == nil {
		return falseV, nil
	}
	b, err := cond.ToBool()
	if err != nil {
		return nil, err
	}
	if b {
		return trueV, nil
	}
	return falseV, nil
}

func BindIf(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("IF: invalid number of arguments: got %d, want 3", len(args))
	}
	return IF(args[0], args[1], args[2])
}
