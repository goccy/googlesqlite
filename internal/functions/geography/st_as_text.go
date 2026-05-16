package geography

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// ST_AS_TEXT returns the WKT representation of a geography.
func ST_AS_TEXT(geo *value.GeographyValue) (value.Value, error) {
	if geo == nil {
		return nil, nil
	}
	wkt, err := geo.ToWKT()
	if err != nil {
		return nil, err
	}
	return value.StringValue(wkt), nil
}

func BindStAsText(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ST_AS_TEXT: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	geo, ok := args[0].(*value.GeographyValue)
	if !ok {
		return nil, fmt.Errorf("ST_AS_TEXT: expected GEOGRAPHY, got %T", args[0])
	}
	return ST_AS_TEXT(geo)
}
