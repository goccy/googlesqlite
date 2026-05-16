package date

import (
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func DATE(args ...value.Value) (value.Value, error) {
	if len(args) == 3 {
		year, err := args[0].ToInt64()
		if err != nil {
			return nil, err
		}
		month, err := args[1].ToInt64()
		if err != nil {
			return nil, err
		}
		day, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		return value.DateValue(time.Time{}.AddDate(int(year)-1, int(month)-1, int(day)-1)), nil
	} else if len(args) == 2 {
		t, err := args[0].ToTime()
		if err != nil {
			return nil, err
		}
		zone, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		loc, err := value.ToLocation(zone)
		if err != nil {
			return nil, err
		}
		return value.DateValue(t.In(loc)), nil
	} else {
		t, err := args[0].ToTime()
		if err != nil {
			return nil, err
		}
		return value.DateValue(t), nil
	}
}

// BindDate short-circuits to NULL when any argument is NULL; DATE
// itself performs the arity dispatch.
var BindDate = helper.ScalarN(DATE)
