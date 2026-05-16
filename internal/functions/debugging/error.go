package debugging

import (
	"errors"

	"github.com/goccy/googlesqlite/internal/value"
)

// BindError implements the SQL ERROR(...) debugging function: a
// no-op-on-NULL forwarder that raises the user-supplied STRING as an
// error. Per debugging_functions.md.
func BindError(args ...value.Value) (value.Value, error) {
	if args[0] == nil {
		return nil, nil
	}
	v, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return nil, errors.New(v)
}
