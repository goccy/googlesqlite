package geography

import (
	"fmt"
	"math"
	"strings"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"

	"github.com/goccy/googlesqlite/internal/value"
)

// BindStNPoints is an alias of ST_NUMPOINTS — both BigQuery and
// Spanner expose the same operation under two names.
func BindStNPoints(args ...value.Value) (value.Value, error) {
	return BindStNumPoints(args...)
}

// BindStExteriorRing returns the outer ring of a POLYGON as a
// LINESTRING. Returns NULL for non-polygon inputs.
func BindStExteriorRing(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_EXTERIORRING", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	rings, ok := g.PolygonRings()
	if !ok || len(rings) == 0 {
		return nil, nil
	}
	return value.NewGeographyLineString(rings[0]), nil
}

// BindStInteriorRings returns the inner rings of a POLYGON as an
// ARRAY<GEOGRAPHY> of LineStrings. An empty array is returned when
// the polygon has no holes (or for the full-globe geography); NULL
// is returned only when the input itself is NULL.
//
// Each emitted ring is rewritten into BigQuery's canonical form:
// reversed so the ring winds CCW, then rotated to start at the
// vertex with the smallest (lat, lng) tuple. That matches the
// LINESTRING display in the upstream Examples table.
func BindStInteriorRings(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_INTERIORRINGS", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	rings, ok := g.PolygonRings()
	if !ok || len(rings) <= 1 {
		return &value.ArrayValue{Values: []value.Value{}}, nil
	}
	inner := rings[1:]
	vals := make([]value.Value, len(inner))
	for i, r := range inner {
		vals[i] = value.NewGeographyLineString(canonicaliseRing(r))
	}
	return &value.ArrayValue{Values: vals}, nil
}

// canonicaliseRing returns a copy of `ring` (which is expected to be
// closed — first vertex equals last) reversed if the original is CW
// and rotated to start at the vertex with the smallest (lat, lng).
// Used by ST_INTERIORRINGS to match BigQuery's display canon.
func canonicaliseRing(ring [][2]float64) [][2]float64 {
	if len(ring) < 4 {
		return append([][2]float64{}, ring...)
	}
	// Drop the trailing closing vertex; we'll reattach at the end.
	open := ring[:len(ring)-1]
	if signedPlanarAreaXY(open) < 0 {
		open = reverseRing(open)
	}
	start := 0
	for i := 1; i < len(open); i++ {
		if lessLatLng(open[i], open[start]) {
			start = i
		}
	}
	rotated := make([][2]float64, 0, len(open)+1)
	rotated = append(rotated, open[start:]...)
	rotated = append(rotated, open[:start]...)
	rotated = append(rotated, rotated[0])
	return rotated
}

// signedPlanarAreaXY is the Shoelace formula treating (lng, lat) as
// planar coordinates; positive when wound CCW.
func signedPlanarAreaXY(ring [][2]float64) float64 {
	n := len(ring)
	if n < 3 {
		return 0
	}
	var area float64
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += ring[i][0] * ring[j][1]
		area -= ring[j][0] * ring[i][1]
	}
	return area / 2
}

func reverseRing(ring [][2]float64) [][2]float64 {
	n := len(ring)
	out := make([][2]float64, n)
	for i := 0; i < n; i++ {
		out[i] = ring[n-1-i]
	}
	return out
}

// lessLatLng compares two (lng, lat) tuples ordered primarily by
// latitude then by longitude.
func lessLatLng(a, b [2]float64) bool {
	if a[1] != b[1] {
		return a[1] < b[1]
	}
	return a[0] < b[0]
}

// BindStDump returns an ARRAY<GEOGRAPHY> containing the simple
// components (POINT / LINESTRING / POLYGON) of the input. Simple
// geographies surface as a single-element array. MULTI* expand to
// one element per component. GEOMETRYCOLLECTION expands to its
// parts (recursively descending into nested collections).
//
// If the optional `dimension` argument is provided (0 for points,
// 1 for lines, 2 for polygons), only components of that dimension
// are kept. `dimension = -1` is equivalent to omitting the argument.
func BindStDump(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("ST_DUMP", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	dim := int64(-1)
	if len(args) == 2 && args[1] != nil {
		d, err := args[1].ToInt64()
		if err != nil {
			return nil, sqError("ST_DUMP", "dimension argument must be INT64: %v", err)
		}
		dim = d
	}
	parts := stDumpExplode(g)
	if dim >= 0 {
		filtered := parts[:0]
		for _, p := range parts {
			if stDumpDimension(p) == dim {
				filtered = append(filtered, p)
			}
		}
		parts = filtered
	}
	vals := make([]value.Value, len(parts))
	for i, p := range parts {
		vals[i] = p
	}
	return &value.ArrayValue{Values: vals}, nil
}

// stDumpExplode flattens the geography into a slice of simple
// (POINT / LINESTRING / POLYGON) components. MULTI* and
// GEOMETRYCOLLECTION are walked; nested collections are descended.
func stDumpExplode(g *value.GeographyValue) []*value.GeographyValue {
	switch g.Kind() {
	case "POINT", "LINESTRING", "POLYGON":
		return []*value.GeographyValue{g}
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		out := make([]*value.GeographyValue, len(pts))
		for i, p := range pts {
			out[i] = value.NewGeographyPoint(p[0], p[1])
		}
		return out
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		out := make([]*value.GeographyValue, len(lines))
		for i, ls := range lines {
			out[i] = value.NewGeographyLineString(ls)
		}
		return out
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		out := make([]*value.GeographyValue, len(polys))
		for i, rings := range polys {
			out[i] = value.NewGeographyPolygon(rings)
		}
		return out
	case "GEOMETRYCOLLECTION":
		parts, _ := g.CollectionParts()
		var out []*value.GeographyValue
		for _, p := range parts {
			out = append(out, stDumpExplode(p)...)
		}
		return out
	}
	return []*value.GeographyValue{g}
}

// stDumpDimension returns 0 for POINT, 1 for LINESTRING, 2 for
// POLYGON. Any other geography kind (which the explosion above does
// not produce) returns -1.
func stDumpDimension(g *value.GeographyValue) int64 {
	switch g.Kind() {
	case "POINT":
		return 0
	case "LINESTRING":
		return 1
	case "POLYGON":
		return 2
	}
	return -1
}

// BindStDumpPoints returns every coordinate position in the
// geography as an ARRAY<GEOGRAPHY> of Points. We surface it as a
// MULTIPOINT (driver ARRAY-of-GEOGRAPHY isn't yet plumbed end to
// end).
func BindStDumpPoints(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_DUMPPOINTS", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	pts := stDumpPointsExplode(g)
	vals := make([]value.Value, len(pts))
	for i, p := range pts {
		vals[i] = value.NewGeographyPoint(p[0], p[1])
	}
	return &value.ArrayValue{Values: vals}, nil
}

func stDumpPointsExplode(g *value.GeographyValue) [][2]float64 {
	var out [][2]float64
	switch g.Kind() {
	case "POINT":
		lng, lat, _ := g.PointCoordinates()
		out = [][2]float64{{lng, lat}}
	case "MULTIPOINT":
		out, _ = g.MultiPointPoints()
	case "LINESTRING":
		out, _ = g.LineStringPoints()
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		for _, ls := range lines {
			out = append(out, ls...)
		}
	case "POLYGON":
		rings, _ := g.PolygonRings()
		for _, r := range rings {
			out = append(out, r...)
		}
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		for _, rings := range polys {
			for _, r := range rings {
				out = append(out, r...)
			}
		}
	case "GEOMETRYCOLLECTION":
		parts, _ := g.CollectionParts()
		for _, p := range parts {
			out = append(out, stDumpPointsExplode(p)...)
		}
	}
	return out
}

// BindStIsClosed reports whether every constituent LINESTRING is
// closed (first == last vertex). Points and polygons are
// considered closed by definition.
func BindStIsClosed(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ISCLOSED", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	switch g.Kind() {
	case "POINT", "MULTIPOINT", "POLYGON", "MULTIPOLYGON":
		return value.BoolValue(true), nil
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		return value.BoolValue(len(pts) >= 2 && pts[0] == pts[len(pts)-1]), nil
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		for _, ls := range lines {
			if len(ls) < 2 || ls[0] != ls[len(ls)-1] {
				return value.BoolValue(false), nil
			}
		}
		return value.BoolValue(true), nil
	}
	return value.BoolValue(false), nil
}

// BindStIsRing reports whether the input is a closed LINESTRING
// with no self-intersection.
func BindStIsRing(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ISRING", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	if g.Kind() != "LINESTRING" {
		return value.BoolValue(false), nil
	}
	pts, _ := g.LineStringPoints()
	if len(pts) < 4 || pts[0] != pts[len(pts)-1] {
		return value.BoolValue(false), nil
	}
	// A simple O(n^2) self-intersection check: for each pair of
	// non-adjacent edges, see whether they cross. Adequate for the
	// vertex counts that fit a SQLite UDF row.
	pl := newPolyline(pts)
	n := pl.NumEdges()
	for i := 0; i < n; i++ {
		ei := pl.Edge(i)
		for j := i + 2; j < n; j++ {
			// Adjacent edges share a vertex; skip those (i+1) and
			// the closing-edge pair (0, n-1).
			if i == 0 && j == n-1 {
				continue
			}
			ej := pl.Edge(j)
			c := s2.CrossingSign(ei.V0, ei.V1, ej.V0, ej.V1)
			if c == s2.Cross {
				return value.BoolValue(false), nil
			}
		}
	}
	return value.BoolValue(true), nil
}

// BindStAzimuth returns the geodesic bearing from point A to
// point B, in radians clockwise from north. Returns NULL when
// either input is not a Point or when they coincide.
func BindStAzimuth(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_AZIMUTH", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	lng1, lat1, ok1 := a.PointCoordinates()
	lng2, lat2, ok2 := b.PointCoordinates()
	if !ok1 || !ok2 {
		return nil, nil
	}
	if lng1 == lng2 && lat1 == lat2 {
		return nil, nil
	}
	phi1 := degToRad(lat1)
	phi2 := degToRad(lat2)
	dl := degToRad(lng2 - lng1)
	y := math.Sin(dl) * math.Cos(phi2)
	x := math.Cos(phi1)*math.Sin(phi2) - math.Sin(phi1)*math.Cos(phi2)*math.Cos(dl)
	az := math.Atan2(y, x)
	if az < 0 {
		az += 2 * math.Pi
	}
	return value.FloatValue(az), nil
}

// BindStAngle returns the clockwise angle, in radians from the
// half-line B→A to the half-line B→C, on the sphere. The result is
// in the range [0, 2π). Returns NULL if either A or C is equal to
// B, or is antipodal to B (the direction would be undefined).
func BindStAngle(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, sqError("ST_ANGLE", "invalid number of arguments: got %d, want 3", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	c := geographyArg(args[2])
	if a == nil || b == nil || c == nil {
		return nil, nil
	}
	lngA, latA, ok1 := a.PointCoordinates()
	lngB, latB, ok2 := b.PointCoordinates()
	lngC, latC, ok3 := c.PointCoordinates()
	if !ok1 || !ok2 || !ok3 {
		return nil, nil
	}
	pa := s2.PointFromLatLng(s2.LatLngFromDegrees(latA, lngA))
	pb := s2.PointFromLatLng(s2.LatLngFromDegrees(latB, lngB))
	pc := s2.PointFromLatLng(s2.LatLngFromDegrees(latC, lngC))
	// Equal or antipodal: dot product near +1 (equal) or -1 (antipodal).
	const eps = 1e-12
	if dot := pa.Dot(pb.Vector); dot > 1-eps || dot < -1+eps {
		return nil, nil
	}
	if dot := pc.Dot(pb.Vector); dot > 1-eps || dot < -1+eps {
		return nil, nil
	}
	// Bearings from B (in radians clockwise from north).
	bearA := initialBearingRadians(latB, lngB, latA, lngA)
	bearC := initialBearingRadians(latB, lngB, latC, lngC)
	angle := bearC - bearA
	for angle < 0 {
		angle += 2 * math.Pi
	}
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}
	return value.FloatValue(angle), nil
}

// initialBearingRadians computes the initial bearing (forward
// azimuth) from point (lat1, lng1) to (lat2, lng2) on a sphere, in
// radians clockwise from north, in the range [0, 2π).
func initialBearingRadians(lat1, lng1, lat2, lng2 float64) float64 {
	phi1 := degToRad(lat1)
	phi2 := degToRad(lat2)
	dLng := degToRad(lng2 - lng1)
	y := math.Sin(dLng) * math.Cos(phi2)
	x := math.Cos(phi1)*math.Sin(phi2) - math.Sin(phi1)*math.Cos(phi2)*math.Cos(dLng)
	b := math.Atan2(y, x)
	for b < 0 {
		b += 2 * math.Pi
	}
	return b
}

func degToRad(d float64) float64 { return d * math.Pi / 180 }

// BindStHausdorffDistance computes the Hausdorff distance between
// two geographies by sampling each vertex of one and taking the
// closest vertex on the other. When the optional `directed`
// argument is TRUE, only the directed `hAB` value (max over vertices
// of `a` of the closest distance to any vertex of `b`) is returned.
// When FALSE or absent the symmetric Hausdorff distance — the larger
// of the two directed values — is returned.
//
// The analyzer materialises named arguments as positional after the
// two geographies, so the runtime can see up to 3 args. Any boolean
// argument with value TRUE enables `directed`.
func BindStHausdorffDistance(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, sqError("ST_HAUSDORFFDISTANCE", "invalid number of arguments: got %d, want between 2 and 3", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	directed := false
	if len(args) == 3 && args[2] != nil {
		if v, err := args[2].ToBool(); err == nil {
			directed = v
		}
	}
	pa := allVertices(a)
	pb := allVertices(b)
	if len(pa) == 0 || len(pb) == 0 {
		return nil, nil
	}
	// Hausdorff over linestrings / polygons must consider the
	// closest point on each segment of the other geometry, not just
	// the vertices — picking just vertex-to-vertex distances over-
	// estimates by a factor of 2 on the example data. The max over
	// all points of A of min-distance-to-B is, however, still
	// achieved at one of A's vertices (the per-segment distance
	// function is piecewise convex), so we iterate over vertices of
	// A and use point-to-segment distance against B.
	ea := allEdges(a)
	eb := allEdges(b)
	hAB := directedHausdorff(pa, eb)
	if directed {
		return value.FloatValue(hAB), nil
	}
	hBA := directedHausdorff(pb, ea)
	if hBA > hAB {
		return value.FloatValue(hBA), nil
	}
	return value.FloatValue(hAB), nil
}

// directedHausdorff computes max over `vertices` of min point-to-edge
// distance against `edges`. `edges` is a flat list of consecutive
// (p1, p2) pairs in lng/lat space, transformed to s2.Point internally.
func directedHausdorff(vertices []s2.Point, edges [][2]s2.Point) float64 {
	h := 0.0
	for _, v := range vertices {
		minD := math.Inf(1)
		for _, e := range edges {
			d := distanceAngleToMeters(s2.DistanceFromSegment(v, e[0], e[1]))
			if d < minD {
				minD = d
			}
		}
		if minD > h {
			h = minD
		}
	}
	return h
}

// allEdges returns every line segment in the geography as a
// (start, end) s2.Point pair. Points contribute a degenerate edge
// (p, p) so the directedHausdorff helper can call DistanceFromSegment
// on them uniformly.
func allEdges(g *value.GeographyValue) [][2]s2.Point {
	if g == nil {
		return nil
	}
	var out [][2]s2.Point
	emit := func(ring [][2]float64) {
		if len(ring) == 0 {
			return
		}
		prev := s2.PointFromLatLng(s2.LatLngFromDegrees(ring[0][1], ring[0][0]))
		if len(ring) == 1 {
			out = append(out, [2]s2.Point{prev, prev})
			return
		}
		for i := 1; i < len(ring); i++ {
			cur := s2.PointFromLatLng(s2.LatLngFromDegrees(ring[i][1], ring[i][0]))
			out = append(out, [2]s2.Point{prev, cur})
			prev = cur
		}
	}
	switch g.Kind() {
	case "POINT":
		lng, lat, ok := g.PointCoordinates()
		if ok {
			emit([][2]float64{{lng, lat}})
		}
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		for _, p := range pts {
			emit([][2]float64{p})
		}
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		emit(pts)
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		for _, ls := range lines {
			emit(ls)
		}
	case "POLYGON":
		rings, _ := g.PolygonRings()
		for _, r := range rings {
			emit(r)
		}
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		for _, rings := range polys {
			for _, r := range rings {
				emit(r)
			}
		}
	case "GEOMETRYCOLLECTION":
		parts, _ := g.CollectionParts()
		for _, p := range parts {
			out = append(out, allEdges(p)...)
		}
	}
	return out
}

// BindStHausdorffDWithin reports whether the Hausdorff distance
// is within the given threshold (meters). When the optional
// `directed` named argument is TRUE, the comparison uses the
// directed hAB Hausdorff distance; otherwise the symmetric (max of
// both directed) value is used.
func BindStHausdorffDWithin(args ...value.Value) (value.Value, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, sqError("ST_HAUSDORFFDWITHIN", "invalid number of arguments: got %d, want between 3 and 4", len(args))
	}
	hArgs := []value.Value{args[0], args[1]}
	if len(args) == 4 {
		hArgs = append(hArgs, args[3])
	}
	r, err := BindStHausdorffDistance(hArgs...)
	if err != nil || r == nil {
		return r, err
	}
	d, err := r.ToFloat64()
	if err != nil {
		return nil, err
	}
	thr, err := args[2].ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.BoolValue(d <= thr), nil
}

func allVertices(g *value.GeographyValue) []s2.Point {
	var out []s2.Point
	out = append(out, geogToS2Points(g)...)
	for _, pl := range geogToS2Polylines(g) {
		for i := range *pl {
			out = append(out, (*pl)[i])
		}
	}
	for _, pg := range geogToS2Polygons(g) {
		for i := 0; i < pg.NumLoops(); i++ {
			loop := pg.Loop(i)
			for j := 0; j < loop.NumVertices(); j++ {
				out = append(out, loop.Vertex(j))
			}
		}
	}
	return out
}

// BindS2CellIDFromPoint returns the S2 cell-id covering the
// input POINT at the requested level (default 30).
func BindS2CellIDFromPoint(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("S2_CELLIDFROMPOINT", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	lng, lat, ok := g.PointCoordinates()
	if !ok {
		return nil, sqError("S2_CELLIDFROMPOINT", "argument is not a POINT")
	}
	level := 30
	if len(args) == 2 && args[1] != nil {
		l, err := args[1].ToInt64()
		if err != nil {
			return nil, err
		}
		level = int(l)
		if level < 0 || level > 30 {
			return nil, sqError("S2_CELLIDFROMPOINT", "level out of range [0, 30]")
		}
	}
	cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng)).Parent(level)
	return value.IntValue(int64(cellID)), nil
}

// BindS2CoveringCellIDs returns an array of S2 cell-ids covering
// the input geography at a fixed level. Since the runtime doesn't
// have ARRAY<INT64> bridging from a single UDF cleanly here, we
// emit a JSON array literal so callers can JSON_EXTRACT_ARRAY it.
func BindS2CoveringCellIDs(args ...value.Value) (value.Value, error) {
	// Upstream signature:
	//   S2_COVERINGCELLIDS(geog [, min_level [, max_level [, max_cells
	//                       [, buffer]]]])
	// We honour min_level / max_level / max_cells; the `buffer`
	// argument is accepted but ignored (our approximation already
	// returns a conservative covering).
	if len(args) < 1 || len(args) > 5 {
		return nil, sqError("S2_COVERINGCELLIDS", "invalid number of arguments: got %d, want between 1 and 5", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	// BigQuery defaults: min_level=0, max_level=30, max_cells=8.
	// Picking min=max=14 (the previous default) silently coerced the
	// covering into a single level, which doesn't match upstream
	// behaviour when only `min_level` is supplied.
	minLevel := 0
	maxLevel := 30
	maxCells := 8
	if len(args) >= 2 && args[1] != nil {
		l, err := args[1].ToInt64()
		if err != nil {
			return nil, err
		}
		minLevel = int(l)
	}
	if len(args) >= 3 && args[2] != nil {
		l, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		maxLevel = int(l)
	}
	if len(args) >= 4 && args[3] != nil {
		n, err := args[3].ToInt64()
		if err != nil {
			return nil, err
		}
		if n > 0 {
			maxCells = int(n)
		}
	}
	cov := &s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel, MaxCells: maxCells}
	var cu s2.CellUnion
	switch g.Kind() {
	case "POINT", "MULTIPOINT":
		for _, p := range geogToS2Points(g) {
			cell := s2.CellFromPoint(p).ID()
			if cell.Level() > maxLevel {
				cell = cell.Parent(maxLevel)
			}
			if cell.Level() < minLevel {
				continue
			}
			cu = append(cu, cell)
		}
	case "LINESTRING", "MULTILINESTRING":
		for _, pl := range geogToS2Polylines(g) {
			cu = append(cu, cov.Covering(pl)...)
		}
	case "POLYGON", "MULTIPOLYGON":
		for _, pg := range geogToS2Polygons(g) {
			cu = append(cu, cov.Covering(pg)...)
		}
	}
	cu.Normalize()
	vals := make([]value.Value, len(cu))
	for i, c := range cu {
		vals[i] = value.IntValue(int64(c))
	}
	return &value.ArrayValue{Values: vals}, nil
}

// BindStClosestPoint returns the geodesic-nearest point on `a`
// to `b` (a representative vertex if no projection makes sense).
func BindStClosestPoint(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, sqError("ST_CLOSESTPOINT", "invalid number of arguments: got %d, want between 2 and 3", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	target := allVertices(b)
	if len(target) == 0 {
		return nil, nil
	}
	candidates := allVertices(a)
	if len(candidates) == 0 {
		return nil, nil
	}
	bestIdx := 0
	bestD := math.Inf(1)
	for i, p := range candidates {
		for _, t := range target {
			d := float64(p.Distance(t))
			if d < bestD {
				bestD = d
				bestIdx = i
			}
		}
	}
	lat, lng := pointToLatLng(candidates[bestIdx])
	return value.NewGeographyPoint(lng, lat), nil
}

// BindStLineInterpolatePoint returns the point at the given
// fraction (0..1) along a LINESTRING's geodesic length.
func BindStLineInterpolatePoint(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_LINEINTERPOLATEPOINT", "invalid number of arguments: got %d, want 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil || args[1] == nil {
		return nil, nil
	}
	frac, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	plis := geogToS2Polylines(g)
	if len(plis) == 0 {
		return nil, nil
	}
	pl := plis[0]
	p, _ := pl.Interpolate(frac)
	lat, lng := pointToLatLng(p)
	return value.NewGeographyPoint(lng, lat), nil
}

// BindStLineLocatePoint returns the normalised position (0..1)
// of the closest point on a LINESTRING to a POINT.
func BindStLineLocatePoint(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_LINELOCATEPOINT", "invalid number of arguments: got %d, want 2", len(args))
	}
	line := geographyArg(args[0])
	pt := geographyArg(args[1])
	if line == nil || pt == nil {
		return nil, nil
	}
	lng, lat, ok := pt.PointCoordinates()
	if !ok {
		return nil, nil
	}
	plis := geogToS2Polylines(line)
	if len(plis) == 0 {
		return nil, nil
	}
	pl := plis[0]
	target := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
	// Walk edges; project the target onto each and keep the one
	// with the smallest geodesic distance.
	bestFrac := 0.0
	bestDist := math.Inf(1)
	cumLength := 0.0
	total := pl.Length().Radians()
	if total == 0 {
		return value.FloatValue(0), nil
	}
	for i := 0; i < pl.NumEdges(); i++ {
		e := pl.Edge(i)
		segLen := e.V0.Distance(e.V1).Radians()
		// Closest point on the geodesic segment to the target.
		proj := s2.Project(target, e.V0, e.V1)
		d := float64(proj.Distance(target))
		if d < bestDist {
			bestDist = d
			offset := float64(e.V0.Distance(proj).Radians())
			bestFrac = (cumLength + offset) / total
		}
		cumLength += segLen
	}
	if bestFrac < 0 {
		bestFrac = 0
	}
	if bestFrac > 1 {
		bestFrac = 1
	}
	return value.FloatValue(bestFrac), nil
}

// BindStLineSubstring returns the slice of a LINESTRING between
// two fractional positions (0..1) along its length. The returned
// geometry preserves every vertex of the original line that falls
// strictly between the start and end fractions, with new endpoint
// vertices interpolated at the exact cut positions. When start ==
// end, returns the single interpolated POINT.
func BindStLineSubstring(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, sqError("ST_LINESUBSTRING", "invalid number of arguments: got %d, want 3", len(args))
	}
	g := geographyArg(args[0])
	if g == nil || args[1] == nil || args[2] == nil {
		return nil, nil
	}
	start, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	end, err := args[2].ToFloat64()
	if err != nil {
		return nil, err
	}
	if start < 0 {
		start = 0
	}
	if end > 1 {
		end = 1
	}
	if start > end {
		return nil, nil
	}
	plis := geogToS2Polylines(g)
	if len(plis) == 0 {
		return nil, nil
	}
	pl := plis[0]
	if start == end {
		p, _ := pl.Interpolate(start)
		lat, lng := pointToLatLng(p)
		return value.NewGeographyPoint(lng, lat), nil
	}
	pa, ka := pl.Interpolate(start)
	pb, kb := pl.Interpolate(end)
	verts := [][2]float64{}
	appendPoint := func(p s2.Point) {
		lat, lng := pointToLatLng(p)
		v := [2]float64{lng, lat}
		if len(verts) > 0 && verts[len(verts)-1] == v {
			return
		}
		verts = append(verts, v)
	}
	appendPoint(pa)
	// Vertices strictly between the two cuts.
	for i := ka; i < kb; i++ {
		if i < 0 || i >= len(*pl) {
			continue
		}
		appendPoint((*pl)[i])
	}
	appendPoint(pb)
	return value.NewGeographyLineString(verts), nil
}

// BindStGeoHash encodes a POINT into a base-32 geohash string.
func BindStGeoHash(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("ST_GEOHASH", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	lng, lat, ok := g.PointCoordinates()
	if !ok {
		return nil, sqError("ST_GEOHASH", "argument is not a POINT")
	}
	precision := 12
	if len(args) == 2 && args[1] != nil {
		p, err := args[1].ToInt64()
		if err != nil {
			return nil, err
		}
		precision = int(p)
		if precision < 1 || precision > 20 {
			return nil, sqError("ST_GEOHASH", "precision out of range [1, 20]")
		}
	}
	return value.StringValue(encodeGeohash(lat, lng, precision)), nil
}

// BindStGeogPointFromGeoHash decodes a geohash to its centre POINT.
func BindStGeogPointFromGeoHash(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_GEOGPOINTFROMGEOHASH", "invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	lat, lng, err := decodeGeohash(s)
	if err != nil {
		return nil, err
	}
	return value.NewGeographyPoint(lng, lat), nil
}

const geohashAlphabet = "0123456789bcdefghjkmnpqrstuvwxyz"

func encodeGeohash(lat, lng float64, precision int) string {
	minLat, maxLat := -90.0, 90.0
	minLng, maxLng := -180.0, 180.0
	var buf strings.Builder
	even := true
	bit := 0
	ch := 0
	for buf.Len() < precision {
		if even {
			mid := (minLng + maxLng) / 2
			if lng > mid {
				ch |= 1 << (4 - bit)
				minLng = mid
			} else {
				maxLng = mid
			}
		} else {
			mid := (minLat + maxLat) / 2
			if lat > mid {
				ch |= 1 << (4 - bit)
				minLat = mid
			} else {
				maxLat = mid
			}
		}
		even = !even
		if bit < 4 {
			bit++
		} else {
			buf.WriteByte(geohashAlphabet[ch])
			bit = 0
			ch = 0
		}
	}
	return buf.String()
}

func decodeGeohash(s string) (lat, lng float64, err error) {
	minLat, maxLat := -90.0, 90.0
	minLng, maxLng := -180.0, 180.0
	even := true
	for _, r := range strings.ToLower(s) {
		idx := strings.IndexRune(geohashAlphabet, r)
		if idx < 0 {
			return 0, 0, sqError("ST_GEOGPOINTFROMGEOHASH", "invalid geohash char %q", r)
		}
		for bit := 4; bit >= 0; bit-- {
			set := idx>>bit&1 == 1
			if even {
				mid := (minLng + maxLng) / 2
				if set {
					minLng = mid
				} else {
					maxLng = mid
				}
			} else {
				mid := (minLat + maxLat) / 2
				if set {
					minLat = mid
				} else {
					maxLat = mid
				}
			}
			even = !even
		}
	}
	return (minLat + maxLat) / 2, (minLng + maxLng) / 2, nil
}

// BindStAsKML renders a POINT / LINESTRING / POLYGON as a KML
// fragment. MULTI* unwraps into a <MultiGeometry>.
func BindStAsKML(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ASKML", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	return value.StringValue(toKML(g)), nil
}

func toKML(g *value.GeographyValue) string {
	switch g.Kind() {
	case "POINT":
		lng, lat, _ := g.PointCoordinates()
		return fmt.Sprintf("<Point><coordinates>%s,%s</coordinates></Point>", gjFloat(lng), gjFloat(lat))
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		return fmt.Sprintf("<LineString><coordinates>%s</coordinates></LineString>", kmlCoords(pts))
	case "POLYGON":
		rings, _ := g.PolygonRings()
		var b strings.Builder
		b.WriteString("<Polygon>")
		for i, r := range rings {
			tag := "outerBoundaryIs"
			if i > 0 {
				tag = "innerBoundaryIs"
			}
			fmt.Fprintf(&b, "<%s><LinearRing><coordinates>%s</coordinates></LinearRing></%s>",
				tag, kmlCoords(r), tag)
		}
		b.WriteString("</Polygon>")
		return b.String()
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		var b strings.Builder
		b.WriteString("<MultiGeometry>")
		for _, p := range pts {
			b.WriteString(toKML(value.NewGeographyPoint(p[0], p[1])))
		}
		b.WriteString("</MultiGeometry>")
		return b.String()
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		var b strings.Builder
		b.WriteString("<MultiGeometry>")
		for _, ls := range lines {
			b.WriteString(toKML(value.NewGeographyLineString(ls)))
		}
		b.WriteString("</MultiGeometry>")
		return b.String()
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		var b strings.Builder
		b.WriteString("<MultiGeometry>")
		for _, rings := range polys {
			b.WriteString(toKML(value.NewGeographyPolygon(rings)))
		}
		b.WriteString("</MultiGeometry>")
		return b.String()
	}
	return ""
}

func kmlCoords(pts [][2]float64) string {
	var b strings.Builder
	for i, p := range pts {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%s,%s", gjFloat(p[0]), gjFloat(p[1]))
	}
	return b.String()
}

// BindStGeogFromKML parses a small subset of KML (Point /
// LineString / Polygon).
func BindStGeogFromKML(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_GEOGFROMKML", "invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return parseKML(s)
}

func parseKML(s string) (*value.GeographyValue, error) {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(s, "<Point>"):
		c := extractCoords(s)
		if len(c) == 0 {
			return nil, sqError("ST_GEOGFROMKML", "missing coordinates")
		}
		return value.NewGeographyPoint(c[0][0], c[0][1]), nil
	case strings.HasPrefix(s, "<LineString>"):
		c := extractCoords(s)
		return value.NewGeographyLineString(c), nil
	case strings.HasPrefix(s, "<Polygon>"):
		// Rings: outerBoundaryIs first, then innerBoundaryIs in order.
		outer := extractCoordsBetween(s, "<outerBoundaryIs>", "</outerBoundaryIs>")
		inner := extractAllCoordsBetween(s, "<innerBoundaryIs>", "</innerBoundaryIs>")
		rings := [][][2]float64{outer}
		rings = append(rings, inner...)
		return value.NewGeographyPolygon(rings), nil
	}
	return nil, sqError("ST_GEOGFROMKML", "unsupported KML payload")
}

func extractCoords(s string) [][2]float64 {
	start := strings.Index(s, "<coordinates>")
	end := strings.Index(s, "</coordinates>")
	if start < 0 || end < 0 || end < start {
		return nil
	}
	return parseCoordsRun(s[start+len("<coordinates>") : end])
}

func extractCoordsBetween(s, open, close string) [][2]float64 {
	i := strings.Index(s, open)
	if i < 0 {
		return nil
	}
	j := strings.Index(s[i:], close)
	if j < 0 {
		return nil
	}
	return extractCoords(s[i : i+j])
}

func extractAllCoordsBetween(s, open, close string) [][][2]float64 {
	var out [][][2]float64
	for {
		i := strings.Index(s, open)
		if i < 0 {
			return out
		}
		j := strings.Index(s[i:], close)
		if j < 0 {
			return out
		}
		c := extractCoords(s[i : i+j])
		out = append(out, c)
		s = s[i+j+len(close):]
	}
}

func parseCoordsRun(s string) [][2]float64 {
	tokens := strings.Fields(s)
	out := make([][2]float64, 0, len(tokens))
	for _, t := range tokens {
		parts := strings.Split(t, ",")
		if len(parts) < 2 {
			continue
		}
		lng, lng_err := parseFloat(parts[0])
		lat, lat_err := parseFloat(parts[1])
		if lng_err != nil || lat_err != nil {
			continue
		}
		out = append(out, [2]float64{lng, lat})
	}
	return out
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%g", &f)
	return f, err
}

// BindStGeogFrom dispatches between text (WKT / GeoJSON / KML)
// and binary (WKB) constructors based on the input shape.
func BindStGeogFrom(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_GEOGFROM", "invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	switch v := args[0].(type) {
	case value.BytesValue:
		return geographyFromWKB([]byte(v))
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	t := strings.TrimSpace(s)
	if strings.HasPrefix(t, "{") {
		return geographyFromGeoJSON(t)
	}
	if strings.HasPrefix(t, "<") {
		return parseKML(t)
	}
	return value.GeographyFromWKT(t)
}

// BindStExtent reduces a list of geographies to their union
// envelope. Variadic input; NULLs skip.
func BindStExtent(args ...value.Value) (value.Value, error) {
	minLng, minLat := 180.0, 90.0
	maxLng, maxLat := -180.0, -90.0
	any := false
	push := func(lng, lat float64) {
		any = true
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
	for _, a := range args {
		g := geographyArg(a)
		if g == nil {
			continue
		}
		for _, p := range allVertices(g) {
			ll := s2.LatLngFromPoint(p)
			push(ll.Lng.Degrees(), ll.Lat.Degrees())
		}
	}
	if !any {
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

// BindStSnapToGrid quantises every vertex to the given grid size.
func BindStSnapToGrid(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, sqError("ST_SNAPTOGRID", "invalid number of arguments: got %d, want between 2 and 3", len(args))
	}
	g := geographyArg(args[0])
	if g == nil || args[1] == nil {
		return nil, nil
	}
	size, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		return g, nil
	}
	snap := func(v float64) float64 {
		return math.Round(v/size) * size
	}
	switch g.Kind() {
	case "POINT":
		lng, lat, _ := g.PointCoordinates()
		return value.NewGeographyPoint(snap(lng), snap(lat)), nil
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		out := make([][2]float64, len(pts))
		for i, p := range pts {
			out[i] = [2]float64{snap(p[0]), snap(p[1])}
		}
		return value.NewGeographyMultiPoint(out), nil
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		out := make([][2]float64, len(pts))
		for i, p := range pts {
			out[i] = [2]float64{snap(p[0]), snap(p[1])}
		}
		return value.NewGeographyLineString(out), nil
	case "POLYGON":
		rings, _ := g.PolygonRings()
		out := make([][][2]float64, len(rings))
		for i, r := range rings {
			rr := make([][2]float64, len(r))
			for j, p := range r {
				rr[j] = [2]float64{snap(p[0]), snap(p[1])}
			}
			out[i] = rr
		}
		return value.NewGeographyPolygon(out), nil
	}
	return g, nil
}

// BindStIntersectsBox returns true when the geography intersects
// the axis-aligned bounding box defined by [lon_lo, lat_lo,
// lon_hi, lat_hi].
func BindStIntersectsBox(args ...value.Value) (value.Value, error) {
	if len(args) != 5 {
		return nil, sqError("ST_INTERSECTSBOX", "invalid number of arguments: got %d, want 5", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	lonLo, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	latLo, err := args[2].ToFloat64()
	if err != nil {
		return nil, err
	}
	lonHi, err := args[3].ToFloat64()
	if err != nil {
		return nil, err
	}
	latHi, err := args[4].ToFloat64()
	if err != nil {
		return nil, err
	}
	rect := s2.RectFromLatLng(s2.LatLngFromDegrees(latLo, lonLo))
	rect = rect.AddPoint(s2.LatLngFromDegrees(latHi, lonHi))
	for _, p := range allVertices(g) {
		if rect.ContainsPoint(p) {
			return value.BoolValue(true), nil
		}
	}
	// Also: polygon may fully contain the box. Check that case via
	// containment of the rectangle's corners.
	pgs := geogToS2Polygons(g)
	corners := []s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(latLo, lonLo)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(latLo, lonHi)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(latHi, lonLo)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(latHi, lonHi)),
	}
	for _, pg := range pgs {
		for _, c := range corners {
			if pg.ContainsPoint(c) {
				return value.BoolValue(true), nil
			}
		}
	}
	return value.BoolValue(false), nil
}

// silence unused-import warning when only some helpers use s1.
var _ s1.Angle
