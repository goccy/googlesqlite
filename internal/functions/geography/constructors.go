package geography

import (
	"github.com/golang/geo/s2"

	"github.com/goccy/googlesqlite/internal/value"
)

// BindStMakeLine joins two points into a 2-vertex LINESTRING.
// Variadic >2 inputs chain them in order. NULLs skip.
func BindStMakeLine(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, sqError("ST_MAKELINE", "needs at least 2 points")
	}
	var pts [][2]float64
	for _, v := range args {
		g := geographyArg(v)
		if g == nil {
			continue
		}
		switch g.Kind() {
		case "POINT":
			lng, lat, _ := g.PointCoordinates()
			pts = append(pts, [2]float64{lng, lat})
		case "LINESTRING":
			lp, _ := g.LineStringPoints()
			pts = append(pts, lp...)
		default:
			return nil, sqError("ST_MAKELINE", "unsupported input kind %q", g.Kind())
		}
	}
	if len(pts) < 2 {
		return nil, sqError("ST_MAKELINE", "fewer than 2 vertices")
	}
	return value.NewGeographyLineString(pts), nil
}

// BindStMakePolygon builds a polygon from an outer LINESTRING (and
// optional holes, also LINESTRINGs). Rings auto-close if their
// first/last vertex differ.
func BindStMakePolygon(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, sqError("ST_MAKEPOLYGON", "needs an outer ring")
	}
	rings := make([][][2]float64, 0, len(args))
	for _, v := range args {
		g := geographyArg(v)
		if g == nil {
			continue
		}
		if g.Kind() != "LINESTRING" {
			return nil, sqError("ST_MAKEPOLYGON", "ring must be a LINESTRING, got %q", g.Kind())
		}
		pts, _ := g.LineStringPoints()
		if len(pts) > 0 && pts[0] != pts[len(pts)-1] {
			pts = append(pts, pts[0])
		}
		if len(pts) < 4 {
			return nil, sqError("ST_MAKEPOLYGON", "ring needs at least 3 distinct vertices")
		}
		rings = append(rings, pts)
	}
	return value.NewGeographyPolygon(rings), nil
}

// BindStMakePolygonOriented is identical to ST_MAKEPOLYGON for our
// purposes — orientation is preserved as given.
func BindStMakePolygonOriented(args ...value.Value) (value.Value, error) {
	return BindStMakePolygon(args...)
}

// BindStCentroid returns the geographic centroid.
//   - POINT: returns itself.
//   - LINESTRING: returns the polyline centroid.
//   - POLYGON: returns the polygon centroid (s2.Polygon.Centroid).
//   - MULTI*: returns the area-weighted centroid of the underlying
//     pieces (or the unweighted average where area is undefined).
func BindStCentroid(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_CENTROID", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	switch g.Kind() {
	case "POINT":
		return g, nil
	case "LINESTRING", "MULTILINESTRING":
		var totalWeight float64
		var weighted s2.Point
		for _, pl := range geogToS2Polylines(g) {
			c := pl.Centroid()
			w := pl.Length().Radians()
			weighted = s2.Point{Vector: weighted.Add(c.Mul(w))}
			totalWeight += w
		}
		if totalWeight == 0 {
			return nil, nil
		}
		lat, lng := pointToLatLng(s2.Point{Vector: weighted.Normalize()})
		return value.NewGeographyPoint(lng, lat), nil
	case "POLYGON", "MULTIPOLYGON":
		var weighted s2.Point
		var totalArea float64
		for _, pg := range geogToS2Polygons(g) {
			a := pg.Area()
			weighted = s2.Point{Vector: weighted.Add(pg.Centroid().Mul(a))}
			totalArea += a
		}
		if totalArea == 0 {
			return nil, nil
		}
		lat, lng := pointToLatLng(s2.Point{Vector: weighted.Normalize()})
		return value.NewGeographyPoint(lng, lat), nil
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		if len(pts) == 0 {
			return nil, nil
		}
		var sum s2.Point
		for _, p := range pts {
			pt := s2.PointFromLatLng(s2.LatLngFromDegrees(p[1], p[0]))
			sum = s2.Point{Vector: sum.Add(pt.Vector)}
		}
		lat, lng := pointToLatLng(s2.Point{Vector: sum.Normalize()})
		return value.NewGeographyPoint(lng, lat), nil
	}
	return nil, nil
}

// BindStEnvelope returns the bounding rectangle of the geography
// as a POLYGON. The rectangle uses the geography's latitude /
// longitude axis-aligned cap.
func BindStEnvelope(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ENVELOPE", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	minLng, minLat := 180.0, 90.0
	maxLng, maxLat := -180.0, -90.0
	push := func(lng, lat float64) {
		if lng < minLng {
			minLng = lng
		}
		if lng > maxLng {
			maxLng = lng
		}
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
	}
	switch g.Kind() {
	case "POINT":
		lng, lat, _ := g.PointCoordinates()
		push(lng, lat)
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		for _, p := range pts {
			push(p[0], p[1])
		}
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		for _, p := range pts {
			push(p[0], p[1])
		}
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		for _, ls := range lines {
			for _, p := range ls {
				push(p[0], p[1])
			}
		}
	case "POLYGON":
		rings, _ := g.PolygonRings()
		for _, r := range rings {
			for _, p := range r {
				push(p[0], p[1])
			}
		}
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		for _, rings := range polys {
			for _, r := range rings {
				for _, p := range r {
					push(p[0], p[1])
				}
			}
		}
	}
	if minLng > maxLng || minLat > maxLat {
		return nil, nil
	}
	ring := [][2]float64{
		{minLng, minLat},
		{maxLng, minLat},
		{maxLng, maxLat},
		{minLng, maxLat},
		{minLng, minLat},
	}
	return value.NewGeographyPolygon([][][2]float64{ring}), nil
}

// BindStBoundingBox returns the axis-aligned bounding box of the
// input geography as a STRUCT<xmin FLOAT64, ymin FLOAT64, xmax FLOAT64,
// ymax FLOAT64>. POINT EMPTY (and any geography that has no vertices)
// returns NULL.
//
// When the geometry crosses the antimeridian (longitude delta > 180°
// between consecutive vertices of a linestring / polygon ring), the
// runtime unwraps negative longitudes by +360° before computing the
// longitude range so the box stays narrow. This matches BigQuery's
// convention where, e.g., POLYGON((172, -130, -141, 172)) has
// xmin=172 and xmax=230 (the polygon hugs the pole, not the long
// equatorial route).
func BindStBoundingBox(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_BOUNDINGBOX", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	// `oriented => TRUE` parsing can mark a polygon as covering the
	// complement of its ring (everything except a small spot near
	// the antimeridian or a pole). In that case the bounding box is
	// the whole globe.
	if g.Inverted() {
		keys := []string{"xmin", "ymin", "xmax", "ymax"}
		values := []value.Value{
			value.FloatValue(-180),
			value.FloatValue(-90),
			value.FloatValue(180),
			value.FloatValue(90),
		}
		m := map[string]value.Value{}
		for i, k := range keys {
			m[k] = values[i]
		}
		return &value.StructValue{Keys: keys, Values: values, M: m}, nil
	}
	pts := pointsOf(g)
	if len(pts) == 0 {
		return nil, nil
	}
	minLng, maxLng, minLat, maxLat := antimeridianAwareBBox(pts, crossesAntimeridian(g))
	keys := []string{"xmin", "ymin", "xmax", "ymax"}
	values := []value.Value{
		value.FloatValue(minLng),
		value.FloatValue(minLat),
		value.FloatValue(maxLng),
		value.FloatValue(maxLat),
	}
	m := map[string]value.Value{}
	for i, k := range keys {
		m[k] = values[i]
	}
	return &value.StructValue{Keys: keys, Values: values, M: m}, nil
}

// crossesAntimeridian reports whether the geometry has at least one
// edge that spans more than 180° of longitude — a strong signal that
// the upstream caller intended the antimeridian-hugging interior.
// Points and edges within a single hemisphere never trigger this.
func crossesAntimeridian(g *value.GeographyValue) bool {
	if g == nil {
		return false
	}
	rings := walkRings(g)
	for _, ring := range rings {
		for i := 1; i < len(ring); i++ {
			if absDelta(ring[i-1][0], ring[i][0]) > 180 {
				return true
			}
		}
	}
	return false
}

// walkRings returns every ring / linestring contained in the
// geometry, recursing into MULTI* and GEOMETRYCOLLECTION.
func walkRings(g *value.GeographyValue) [][][2]float64 {
	if g == nil {
		return nil
	}
	switch g.Kind() {
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		return [][][2]float64{pts}
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		return lines
	case "POLYGON":
		rings, _ := g.PolygonRings()
		return rings
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		var out [][][2]float64
		for _, rings := range polys {
			out = append(out, rings...)
		}
		return out
	case "GEOMETRYCOLLECTION":
		parts, _ := g.CollectionParts()
		var out [][][2]float64
		for _, p := range parts {
			out = append(out, walkRings(p)...)
		}
		return out
	}
	return nil
}

func absDelta(a, b float64) float64 {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d
}

// antimeridianAwareBBox computes (xmin, xmax, ymin, ymax) over the
// supplied lng/lat points. When `crossing` is true, negative
// longitudes are unwrapped by adding 360 before the min/max scan,
// so the resulting box describes the narrow side of the
// antimeridian-crossing region.
func antimeridianAwareBBox(pts [][2]float64, crossing bool) (minLng, maxLng, minLat, maxLat float64) {
	minLng, maxLng = 180, -180
	minLat, maxLat = 90, -90
	for _, p := range pts {
		lng, lat := p[0], p[1]
		if crossing && lng < 0 {
			lng += 360
		}
		if lng < minLng {
			minLng = lng
		}
		if lng > maxLng {
			maxLng = lng
		}
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
	}
	return
}
