package timestamp

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func FORMAT_TIMESTAMP(format string, t time.Time, zone string) (value.Value, error) {
	loc, err := value.ToLocation(zone)
	if err != nil {
		return nil, err
	}
	t = t.In(loc)
	s, err := helper.FormatTime(format, &t, helper.FormatTypeTimestamp)
	if err != nil {
		return nil, err
	}
	return value.StringValue(s), nil
}

func BindFormatTimestamp(args ...value.Value) (value.Value, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, fmt.Errorf("FORMAT_TIMESTAMP: invalid number of arguments: got %d, want 2 or 3", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	format, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	t, err := args[1].ToTime()
	if err != nil {
		return nil, err
	}
	var zone string
	if len(args) == 3 {
		z, err := args[2].ToString()
		if err != nil {
			return nil, err
		}
		zone = z
	}
	return FORMAT_TIMESTAMP(format, t, zone)
}
