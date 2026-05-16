package geography

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func ST_GEOGPOINT(longitude float64, latitude float64) (value.Value, error) {
	if latitude < -90 || latitude > 90 {
		return nil, fmt.Errorf("ST_GEOGPOINT: invalid latitude: %f", latitude)
	}

	return value.NewGeographyPoint(longitude, latitude), nil
}

func BindStGeogPoint(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("ST_GEOGPOINT: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}

	longitude, err := args[0].ToFloat64()
	if err != nil {
		return nil, fmt.Errorf("ST_GEOGPOINT: invalid longitude argument: %w", err)
	}
	latitude, err := args[1].ToFloat64()
	if err != nil {
		return nil, fmt.Errorf("ST_GEOGPOINT: invalid latitude argument: %w", err)
	}

	return ST_GEOGPOINT(longitude, latitude)
}
