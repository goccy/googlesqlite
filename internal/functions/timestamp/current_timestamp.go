package timestamp

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func CURRENT_TIMESTAMP(zone string) (value.Value, error) {
	loc, err := value.ToLocation(zone)
	if err != nil {
		return nil, err
	}
	return CURRENT_TIMESTAMP_WITH_TIME(time.Now().In(loc))
}

func BindCurrentTimestamp(args ...value.Value) (value.Value, error) {
	if len(args) == 0 {
		return CURRENT_TIMESTAMP("")
	}
	if len(args) == 2 {
		unixNano, err := args[0].ToInt64()
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
		return CURRENT_TIMESTAMP_WITH_TIME(value.TimeFromUnixNano(unixNano).In(loc))
	}
	switch args[0].(type) {
	case value.IntValue:
		unixNano, err := args[0].ToInt64()
		if err != nil {
			return nil, err
		}
		return CURRENT_TIMESTAMP_WITH_TIME(value.TimeFromUnixNano(unixNano))
	case value.StringValue:
		zone, err := args[0].ToString()
		if err != nil {
			return nil, err
		}
		return CURRENT_TIMESTAMP(zone)
	}
	return nil, fmt.Errorf("CURRENT_TIMESTAMP: unexpected argument type %T", args[0])
}
