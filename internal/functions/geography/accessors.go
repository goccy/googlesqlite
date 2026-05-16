package geography

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// BindStStartPoint returns the first point of a LINESTRING.
// Other inputs (POINT, POLYGON, etc.) return NULL.
func BindStStartPoint(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_STARTPOINT", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	pts, ok := g.LineStringPoints()
	if !ok || len(pts) == 0 {
		return nil, nil
	}
	return value.NewGeographyPoint(pts[0][0], pts[0][1]), nil
}

// BindStEndPoint returns the last point of a LINESTRING.
func BindStEndPoint(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ENDPOINT", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	pts, ok := g.LineStringPoints()
	if !ok || len(pts) == 0 {
		return nil, nil
	}
	last := pts[len(pts)-1]
	return value.NewGeographyPoint(last[0], last[1]), nil
}

// BindStPointN returns the Nth (1-based) point of a LINESTRING.
// Negative N counts from the end; out-of-range returns NULL.
func BindStPointN(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_POINTN", "invalid number of arguments: got %d, want 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil || args[1] == nil {
		return nil, nil
	}
	n, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	pts, ok := g.LineStringPoints()
	if !ok || len(pts) == 0 {
		return nil, nil
	}
	idx, err := helper.SafeInt(n)
	if err != nil {
		return nil, err
	}
	if idx > 0 {
		idx--
	} else if idx < 0 {
		idx = len(pts) + idx
	} else {
		return nil, nil
	}
	if idx < 0 || idx >= len(pts) {
		return nil, nil
	}
	return value.NewGeographyPoint(pts[idx][0], pts[idx][1]), nil
}

// BindStGeometryN returns the Nth (1-based) sub-geometry of a
// MULTI* or GEOMETRYCOLLECTION. Singleton inputs return their
// single element only at N=1.
func BindStGeometryN(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_GEOMETRYN", "invalid number of arguments: got %d, want 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil || args[1] == nil {
		return nil, nil
	}
	n, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	nInt, err := helper.SafeInt(n)
	if err != nil {
		return nil, err
	}
	idx := nInt - 1
	if idx < 0 {
		return nil, nil
	}
	switch g.Kind() {
	case "POINT", "LINESTRING", "POLYGON":
		if idx == 0 {
			return g, nil
		}
		return nil, nil
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		if idx >= len(pts) {
			return nil, nil
		}
		return value.NewGeographyPoint(pts[idx][0], pts[idx][1]), nil
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		if idx >= len(lines) {
			return nil, nil
		}
		return value.NewGeographyLineString(lines[idx]), nil
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		if idx >= len(polys) {
			return nil, nil
		}
		return value.NewGeographyPolygon(polys[idx]), nil
	}
	return nil, nil
}

// BindStBoundary returns the boundary of the geography:
// - For LINESTRING / MULTILINESTRING: the start+end points as MULTIPOINT.
// - For POLYGON / MULTIPOLYGON: the outer + inner rings as MULTILINESTRING.
// - For POINT: the empty geometry (NULL here, matching upstream).
func BindStBoundary(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_BOUNDARY", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	switch g.Kind() {
	case "POINT", "MULTIPOINT":
		return nil, nil
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		if len(pts) < 2 {
			return nil, nil
		}
		endpts := [][2]float64{pts[0], pts[len(pts)-1]}
		return value.NewGeographyMultiPoint(endpts), nil
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		endpts := make([][2]float64, 0, len(lines)*2)
		for _, ls := range lines {
			if len(ls) >= 2 {
				endpts = append(endpts, ls[0], ls[len(ls)-1])
			}
		}
		return value.NewGeographyMultiPoint(endpts), nil
	case "POLYGON":
		rings, _ := g.PolygonRings()
		lines := make([][][2]float64, 0, len(rings))
		lines = append(lines, rings...)
		return value.NewGeographyMultiLineString(lines), nil
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		lines := make([][][2]float64, 0)
		for _, rings := range polys {
			lines = append(lines, rings...)
		}
		return value.NewGeographyMultiLineString(lines), nil
	}
	return nil, nil
}
