package geography

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// ST_X returns the longitude of a Point geography. Per BigQuery,
// returns NULL if the input is NULL or is not a single Point.
func ST_X(geo *value.GeographyValue) (value.Value, error) {
	if geo == nil {
		return nil, nil
	}
	lon, _, ok := geo.PointCoordinates()
	if !ok {
		return nil, nil
	}
	return value.FloatValue(lon), nil
}

func BindStX(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ST_X: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	geo, ok := args[0].(*value.GeographyValue)
	if !ok {
		return nil, fmt.Errorf("ST_X: expected GEOGRAPHY, got %T", args[0])
	}
	return ST_X(geo)
}

// ST_Y returns the latitude of a Point geography.
func ST_Y(geo *value.GeographyValue) (value.Value, error) {
	if geo == nil {
		return nil, nil
	}
	_, lat, ok := geo.PointCoordinates()
	if !ok {
		return nil, nil
	}
	return value.FloatValue(lat), nil
}

func BindStY(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ST_Y: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	geo, ok := args[0].(*value.GeographyValue)
	if !ok {
		return nil, fmt.Errorf("ST_Y: expected GEOGRAPHY, got %T", args[0])
	}
	return ST_Y(geo)
}
