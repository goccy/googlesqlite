package geography

import (
	"github.com/goccy/googlesqlite/internal/value"
)

// BindStLength returns the total geodesic length of a (multi)
// linestring in meters. Points and polygons return 0; NULL
// inputs propagate.
func BindStLength(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("ST_LENGTH", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	total := 0.0
	for _, pl := range geogToS2Polylines(g) {
		total += polylineLength(pl)
	}
	return value.FloatValue(total), nil
}

// BindStPerimeter returns the geodesic perimeter of a (multi)
// polygon in meters; non-polygon inputs return 0.
func BindStPerimeter(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("ST_PERIMETER", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	total := 0.0
	for _, pg := range geogToS2Polygons(g) {
		for i := 0; i < pg.NumLoops(); i++ {
			loop := pg.Loop(i)
			n := loop.NumVertices()
			for j := 0; j < n; j++ {
				a := loop.Vertex(j)
				b := loop.Vertex((j + 1) % n)
				total += distanceAngleToMeters(a.Distance(b))
			}
		}
	}
	return value.FloatValue(total), nil
}

// BindStArea returns the geodesic area of a (multi)polygon in
// square meters; non-polygon inputs return 0.
func BindStArea(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("ST_AREA", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	total := 0.0
	for _, pg := range geogToS2Polygons(g) {
		total += polygonArea(pg)
	}
	return value.FloatValue(total), nil
}

// BindStNumPoints returns the total point count contributing to
// the geography.
func BindStNumPoints(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_NUMPOINTS", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	return value.IntValue(int64(numPointsTotal(g))), nil
}

// BindStNumGeometries returns the number of constituent
// geometries: 1 for a singleton kind, N for a MULTI* / collection.
func BindStNumGeometries(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_NUMGEOMETRIES", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	return value.IntValue(int64(numSubGeometries(g))), nil
}

// BindStDimension returns the topological dimension: 0 for
// points, 1 for lines, 2 for polygons. -1 for empty.
func BindStDimension(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_DIMENSION", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	switch g.Kind() {
	case "POINT", "MULTIPOINT":
		return value.IntValue(0), nil
	case "LINESTRING", "MULTILINESTRING":
		return value.IntValue(1), nil
	case "POLYGON", "MULTIPOLYGON":
		return value.IntValue(2), nil
	}
	return value.IntValue(-1), nil
}

// BindStGeometryType returns the WKT-style kind string of the
// geography (e.g. "POINT", "LINESTRING") in lowercase per
// upstream output.
func BindStGeometryType(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_GEOMETRYTYPE", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	return value.StringValue("ST_" + titleCase(g.Kind())), nil
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	// BigQuery's ST_GeometryType returns names like "ST_Point",
	// "ST_LineString", "ST_Polygon", "ST_MultiPoint", etc. Convert
	// the all-caps WKT tag into the camelcase form.
	switch s {
	case "POINT":
		return "Point"
	case "LINESTRING":
		return "LineString"
	case "POLYGON":
		return "Polygon"
	case "MULTIPOINT":
		return "MultiPoint"
	case "MULTILINESTRING":
		return "MultiLineString"
	case "MULTIPOLYGON":
		return "MultiPolygon"
	case "GEOMETRYCOLLECTION":
		return "GeometryCollection"
	}
	return s
}

// BindStIsEmpty returns true when the geography contains zero
// constituent points.
func BindStIsEmpty(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ISEMPTY", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	return value.BoolValue(numPointsTotal(g) == 0), nil
}

// BindStIsCollection returns true when the geography is a
// MULTI* or GEOMETRYCOLLECTION.
func BindStIsCollection(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ISCOLLECTION", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	switch g.Kind() {
	case "MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON", "GEOMETRYCOLLECTION":
		return value.BoolValue(true), nil
	}
	return value.BoolValue(false), nil
}
