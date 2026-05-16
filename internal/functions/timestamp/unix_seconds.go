package timestamp

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func UNIX_SECONDS(t time.Time) (value.Value, error) {
	return value.IntValue(t.Unix()), nil
}

var BindUnixSeconds = helper.Scalar1(func(a value.Value) (value.Value, error) {
	t, err := a.ToTime()
	if err != nil {
		return nil, err
	}
	return UNIX_SECONDS(t)
})

// BindMysqlUnixTimestamp implements MySQL's UNIX_TIMESTAMP, which
// accepts zero or one argument: with no arg it returns the current
// instant as Unix seconds; with one arg it behaves like UNIX_SECONDS.
func BindMysqlUnixTimestamp(args ...value.Value) (value.Value, error) {
	if len(args) == 0 {
		return UNIX_SECONDS(time.Now().UTC())
	}
	return BindUnixSeconds(args...)
}
