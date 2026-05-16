package spanner

import (
	"fmt"

	gotime "time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// BindAddDate implements Spanner's mysql.adddate(date, n) which adds
// `n` days to `date`. The GoogleSQL DATE_ADD takes an INTERVAL; the
// mysql alias takes a bare INT64 representing days.
func BindAddDate(args ...value.Value) (value.Value, error) {
	return shiftDate("ADDDATE", args, 1)
}

// BindSubDate implements Spanner's mysql.subdate(date, n).
func BindSubDate(args ...value.Value) (value.Value, error) {
	return shiftDate("SUBDATE", args, -1)
}

func shiftDate(name string, args []value.Value, sign int) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments: got %d, want 2", name, len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	dv, ok := args[0].(value.DateValue)
	if !ok {
		return nil, fmt.Errorf("%s: first argument must be DATE", name)
	}
	n, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	shifted := gotime.Time(dv).AddDate(0, 0, sign*int(n))
	return value.DateValue(shifted), nil
}
