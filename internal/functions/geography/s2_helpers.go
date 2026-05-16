// Package geography uses golang/geo/s2 for spherical geometry
// computations. This file translates between the driver-native
// GeographyValue (point/line/polygon coordinate tuples) and the
// equivalent s2 objects, then exposes a few high-level helpers
// that geography functions reuse.
package geography

import (
	"fmt"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"

	"github.com/goccy/googlesqlite/internal/value"
)

const earthRadiusMeters = 6371010.0

// geogToS2Points returns the point list for any geography kind by
// flattening multi-geometries; the boolean indicates whether the
// input is a single Point.
func geogToS2Points(g *value.GeographyValue) []s2.Point {
	if g == nil {
		return nil
	}
	switch g.Kind() {
	case "POINT":
		lng, lat, ok := g.PointCoordinates()
		if !ok {
			// POINT EMPTY: no coordinates to surface.
			return nil
		}
		return []s2.Point{s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))}
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		out := make([]s2.Point, 0, len(pts))
		for _, p := range pts {
			out = append(out, s2.PointFromLatLng(s2.LatLngFromDegrees(p[1], p[0])))
		}
		return out
	}
	return nil
}

// geogToS2Polylines returns the polylines underlying a (MULTI)LINESTRING.
func geogToS2Polylines(g *value.GeographyValue) []*s2.Polyline {
	if g == nil {
		return nil
	}
	switch g.Kind() {
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		return []*s2.Polyline{newPolyline(pts)}
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		out := make([]*s2.Polyline, 0, len(lines))
		for _, ls := range lines {
			out = append(out, newPolyline(ls))
		}
		return out
	}
	return nil
}

func newPolyline(pts [][2]float64) *s2.Polyline {
	points := make([]s2.Point, len(pts))
	for i, p := range pts {
		points[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(p[1], p[0]))
	}
	pl := s2.Polyline(points)
	return &pl
}

// geogToS2Polygons returns the polygons underlying a (MULTI)POLYGON.
func geogToS2Polygons(g *value.GeographyValue) []*s2.Polygon {
	if g == nil {
		return nil
	}
	switch g.Kind() {
	case "POLYGON":
		rings, _ := g.PolygonRings()
		if pg, ok := newPolygon(rings); ok {
			return []*s2.Polygon{pg}
		}
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		out := make([]*s2.Polygon, 0, len(polys))
		for _, rings := range polys {
			if pg, ok := newPolygon(rings); ok {
				out = append(out, pg)
			}
		}
		return out
	}
	return nil
}

func newPolygon(rings [][][2]float64) (*s2.Polygon, bool) {
	if len(rings) == 0 {
		return nil, false
	}
	loops := make([]*s2.Loop, 0, len(rings))
	for _, ring := range rings {
		loop := newLoop(ring)
		if loop == nil {
			continue
		}
		loops = append(loops, loop)
	}
	if len(loops) == 0 {
		return nil, false
	}
	return s2.PolygonFromLoops(loops), true
}

func newLoop(ring [][2]float64) *s2.Loop {
	if len(ring) < 3 {
		return nil
	}
	// Drop a trailing closing vertex if present (WKT rings repeat
	// the first vertex at the end).
	if len(ring) > 1 && ring[0] == ring[len(ring)-1] {
		ring = ring[:len(ring)-1]
	}
	if len(ring) < 3 {
		return nil
	}
	pts := make([]s2.Point, len(ring))
	for i, p := range ring {
		pts[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(p[1], p[0]))
	}
	return s2.LoopFromPoints(pts)
}

// numPointsTotal counts every coordinate position across the
// geography (one per Point, len(ring) per ring, etc.). For an
// empty geometry this returns 0.
func numPointsTotal(g *value.GeographyValue) int {
	if g == nil {
		return 0
	}
	switch g.Kind() {
	case "POINT":
		return 1
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		return len(pts)
	case "POLYGON":
		rings, _ := g.PolygonRings()
		n := 0
		for _, r := range rings {
			n += len(r)
		}
		return n
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		return len(pts)
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		n := 0
		for _, ls := range lines {
			n += len(ls)
		}
		return n
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		n := 0
		for _, rings := range polys {
			for _, r := range rings {
				n += len(r)
			}
		}
		return n
	}
	return 0
}

// numSubGeometries returns the count of constituent geometries in
// a MULTI* / GEOMETRYCOLLECTION; for a singleton kind it returns
// 1; for the empty geometry it returns 0.
func numSubGeometries(g *value.GeographyValue) int {
	if g == nil {
		return 0
	}
	switch g.Kind() {
	case "POINT", "LINESTRING", "POLYGON":
		return 1
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		return len(pts)
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		return len(lines)
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		return len(polys)
	}
	return 0
}

// geographyArg is a defensive accessor that converts the SQLite
// boundary value into a *GeographyValue. Returns nil when the
// argument is NULL or not a geography.
func geographyArg(v value.Value) *value.GeographyValue {
	if v == nil {
		return nil
	}
	g, _ := v.(*value.GeographyValue)
	return g
}

// distanceAngle wraps s1.ChordAngle math to return meters for a
// pair of unit-sphere points scaled to Earth's radius.
func distanceAngleToMeters(angle s1.Angle) float64 {
	return angle.Radians() * earthRadiusMeters
}

// polylineLength returns the geodesic length in meters of an s2
// Polyline.
func polylineLength(pl *s2.Polyline) float64 {
	return distanceAngleToMeters(pl.Length())
}

// polygonArea returns the spherical area of an s2 Polygon in
// square meters. The Polygon Area() method already accounts for
// holes (inner rings reduce the total).
func polygonArea(pg *s2.Polygon) float64 {
	return pg.Area() * earthRadiusMeters * earthRadiusMeters
}

// pointToLatLng converts an s2.Point back to (lat, lng) degrees.
func pointToLatLng(p s2.Point) (float64, float64) {
	ll := s2.LatLngFromPoint(p)
	return ll.Lat.Degrees(), ll.Lng.Degrees()
}

// sqError formats a function-specific error message.
func sqError(fn, msg string, args ...any) error {
	return fmt.Errorf("%s: "+msg, append([]any{fn}, args...)...)
}
