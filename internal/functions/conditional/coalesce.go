package conditional

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func COALESCE(args ...value.Value) (value.Value, error) {
	for _, arg := range args {
		if arg == nil {
			continue
		}
		return arg, nil
	}
	return nil, nil
}

func BindCoalesce(args ...value.Value) (value.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("COALESCE: invalid number of arguments: got %d, want at least 1", len(args))
	}
	return COALESCE(args...)
}
