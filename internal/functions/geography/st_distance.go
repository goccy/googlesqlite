package geography

import (
	"errors"
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func ST_DISTANCE(geo1 *value.GeographyValue, geo2 *value.GeographyValue) (value.Value, error) {
	if geo1 == nil || geo2 == nil {
		return nil, errors.New("nil geography")
	}
	dist, err := geo1.DistanceTo(geo2)
	if err != nil {
		return nil, err
	}
	return value.FloatValue(dist), nil
}

func BindStDistance(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("ST_DISTANCE: invalid number of arguments: got %d, want 2", len(args))
	}

	geo1, ok := args[0].(*value.GeographyValue)
	if !ok {
		return nil, fmt.Errorf("ST_DISTANCE: unexpected argument type %T", args[0])
	}
	geo2, ok := args[1].(*value.GeographyValue)
	if !ok {
		return nil, fmt.Errorf("ST_DISTANCE: unexpected argument type %T", args[0])
	}

	return ST_DISTANCE(geo1, geo2)
}
