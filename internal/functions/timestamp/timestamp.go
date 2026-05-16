package timestamp

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TIMESTAMP(v value.Value, zone string) (value.Value, error) {
	loc, err := value.ToLocation(zone)
	if err != nil {
		return nil, err
	}
	switch v.(type) {
	case value.StringValue:
		s, err := v.ToString()
		if err != nil {
			return nil, err
		}
		t, err := value.ParseTimestamp(s, loc)
		if err != nil {
			return nil, err
		}
		return value.TimestampValue(t), nil
	case value.DateValue, value.DatetimeValue:
		t, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		modified, err := value.ModifyTimeZone(t, loc)
		if err != nil {
			return nil, err
		}
		return value.TimestampValue(modified), nil
	}
	return nil, fmt.Errorf("TIMESTAMP: invalid first argument type %T", v)
}

func BindTimestamp(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("TIMESTAMP: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	var zone string
	if len(args) == 2 {
		z, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		zone = z
	}
	return TIMESTAMP(args[0], zone)
}
