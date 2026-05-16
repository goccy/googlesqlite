package geography

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// ST_EQUALS returns true when the two geographies are structurally
// equal — same kind and identical coordinates in declaration order.
// Spatial equality (same point set, possibly different vertex
// orderings) requires full S2 / GEOS support and is not modelled
// here.
func ST_EQUALS(geo1, geo2 *value.GeographyValue) (value.Value, error) {
	if geo1 == nil || geo2 == nil {
		return nil, nil
	}
	eq, err := geo1.EQ(geo2)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(eq), nil
}

func BindStEquals(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("ST_EQUALS: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	geo1, ok := args[0].(*value.GeographyValue)
	if !ok {
		return nil, fmt.Errorf("ST_EQUALS: expected GEOGRAPHY, got %T", args[0])
	}
	geo2, ok := args[1].(*value.GeographyValue)
	if !ok {
		return nil, fmt.Errorf("ST_EQUALS: expected GEOGRAPHY, got %T", args[1])
	}
	return ST_EQUALS(geo1, geo2)
}

// ST_DWITHIN / ST_INTERSECTS are implemented in topology.go using
// golang/geo/s2 across the full geometry surface (Points,
// LineStrings, Polygons, and their MULTI* variants).
