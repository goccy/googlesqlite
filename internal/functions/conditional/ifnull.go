package conditional

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func IFNULL(expr, nullResult value.Value) (value.Value, error) {
	if expr == nil {
		return nullResult, nil
	}
	return expr, nil
}

// BindIfNull observes NULL itself, so it must use the KeepNull
// variant rather than short-circuiting on a NULL argument.
var BindIfNull = helper.Scalar2KeepNull(IFNULL)
