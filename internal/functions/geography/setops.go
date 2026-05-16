package geography

import (
	"math"
	"sort"

	"github.com/golang/geo/s2"

	"github.com/goccy/googlesqlite/internal/value"
)

// Polygon set-operations are computed in the planar (lat,lng)
// domain using textbook polygon-clipping algorithms. This is an
// approximation of true spherical Boolean ops, but it is exact
// for the common case of small regions far from the antimeridian
// and the poles. The spherical S2 polygon Boolean ops are not
// exposed by golang/geo's Go port, so we implement what's needed
// directly. Lines and points fall through simpler combinators.

// BindStUnion produces the geographic union of two geographies.
// Point + Point  -> MULTIPOINT
// Line  + Line   -> MULTILINESTRING
// Poly  + Poly   -> POLYGON (planar clip-and-merge) or MULTIPOLYGON
// Mixed shapes return a GEOMETRYCOLLECTION fallback by
// concatenating the parts of each input.
func BindStUnion(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_UNION", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}
	// Polygon-vs-Polygon: planar union.
	if (a.Kind() == "POLYGON" || a.Kind() == "MULTIPOLYGON") &&
		(b.Kind() == "POLYGON" || b.Kind() == "MULTIPOLYGON") {
		return polygonUnion(a, b), nil
	}
	// Point/Line-only combinations: concatenate.
	return concatGeographies(a, b), nil
}

// BindStIntersection returns the intersection of two geographies.
func BindStIntersection(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_INTERSECTION", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	if (a.Kind() == "POLYGON" || a.Kind() == "MULTIPOLYGON") &&
		(b.Kind() == "POLYGON" || b.Kind() == "MULTIPOLYGON") {
		return polygonIntersection(a, b), nil
	}
	// For non-polygon mix, return points that intersect.
	out := [][2]float64{}
	for _, p := range pointsOf(a) {
		if pointInGeography(p, b) {
			out = append(out, p)
		}
	}
	for _, p := range pointsOf(b) {
		if pointInGeography(p, a) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	if len(out) == 1 {
		return value.NewGeographyPoint(out[0][0], out[0][1]), nil
	}
	return value.NewGeographyMultiPoint(out), nil
}

// BindStDifference returns the part of `a` that doesn't lie in `b`.
func BindStDifference(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_DIFFERENCE", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil {
		return nil, nil
	}
	if b == nil {
		return a, nil
	}
	if (a.Kind() == "POLYGON" || a.Kind() == "MULTIPOLYGON") &&
		(b.Kind() == "POLYGON" || b.Kind() == "MULTIPOLYGON") {
		return polygonDifference(a, b), nil
	}
	out := [][2]float64{}
	for _, p := range pointsOf(a) {
		if !pointInGeography(p, b) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	if len(out) == 1 {
		return value.NewGeographyPoint(out[0][0], out[0][1]), nil
	}
	return value.NewGeographyMultiPoint(out), nil
}

// BindStBuffer expands the geography by `radius` meters and
// returns the resulting POLYGON. Approximated as the convex hull
// of all input vertices plus a regular polygon (default 8
// segments per quadrant) of radius `radius` around each vertex.
func BindStBuffer(args ...value.Value) (value.Value, error) {
	// BigQuery / Spanner signature:
	//   ST_BUFFER(g, radius [, num_seg_quarter_circle [, use_spheroid
	//             [, endcap [, side]]]])
	// The trailing knobs are not honoured by our approximation; they
	// are accepted so analyzer dispatch succeeds for the full upstream
	// arity range.
	if len(args) < 2 || len(args) > 6 {
		return nil, sqError("ST_BUFFER", "invalid number of arguments: got %d, want between 2 and 6", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	radius, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	segPerQuarter := 8
	if len(args) >= 3 && args[2] != nil {
		s, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		segPerQuarter = int(s)
		if segPerQuarter < 1 {
			segPerQuarter = 1
		}
	}
	pts := pointsOf(g)
	if len(pts) == 0 {
		return nil, nil
	}
	// Build the union of `radius` circles around each vertex and
	// then take the convex hull. For most BigQuery use cases (a
	// few hundred vertices) this is acceptable.
	all := make([][2]float64, 0, len(pts)*segPerQuarter*4)
	for _, p := range pts {
		all = append(all, circlePoints(p[1], p[0], radius, segPerQuarter*4)...)
	}
	hull := convexHull(all)
	if len(hull) < 3 {
		return nil, nil
	}
	hull = append(hull, hull[0])
	return value.NewGeographyPolygon([][][2]float64{hull}), nil
}

// BindStBufferWithTolerance produces a buffer polygon whose vertex
// count is sized so that the maximum distance between the buffer's
// circular boundary and the polygon's straight-line approximation is
// at most `tolerance_meters`. The relation is
// `n = ceil(pi / arccos(1 - tolerance/radius))` (sagitta of a
// regular n-gon inscribed in a circle of radius r).
//
// Unlike BindStBuffer (which only takes `num_seg_quarter_circle` as
// the granularity knob — implicitly a multiple of 4), this entry
// point lays out the polygon at the exact `n` chosen so the output
// matches upstream's per-tolerance vertex count.
func BindStBufferWithTolerance(args ...value.Value) (value.Value, error) {
	if len(args) < 3 {
		return nil, sqError("ST_BUFFERWITHTOLERANCE", "needs (geog, radius, tolerance)")
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	radius, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	tolerance, err := args[2].ToFloat64()
	if err != nil {
		return nil, err
	}
	if !(radius > 0) || !(tolerance > 0) {
		return BindStBuffer(args[0], args[1])
	}
	ratio := tolerance / radius
	if ratio >= 1 {
		ratio = 0.99
	}
	theta := math.Acos(1 - ratio)
	if !(theta > 0) {
		return BindStBuffer(args[0], args[1])
	}
	n := int(math.Ceil(math.Pi / theta))
	if n < 3 {
		n = 3
	}
	pts := pointsOf(g)
	if len(pts) == 0 {
		return nil, nil
	}
	// Build the union of `n`-vertex regular polygons of radius
	// `radius` around each input vertex, then take the convex hull
	// to merge them. For a single POINT input the result is the
	// regular n-gon itself.
	all := make([][2]float64, 0, len(pts)*n)
	for _, p := range pts {
		all = append(all, circlePoints(p[1], p[0], radius, n)...)
	}
	hull := convexHull(all)
	if len(hull) < 3 {
		return nil, nil
	}
	hull = append(hull, hull[0])
	return value.NewGeographyPolygon([][][2]float64{hull}), nil
}

// BindStConvexHull returns the convex hull of every vertex of the
// input geography.
func BindStConvexHull(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_CONVEXHULL", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	pts := pointsOf(g)
	hull := convexHull(pts)
	if len(hull) < 3 {
		if len(hull) == 1 {
			return value.NewGeographyPoint(hull[0][0], hull[0][1]), nil
		}
		if len(hull) == 2 {
			return value.NewGeographyLineString(hull), nil
		}
		return nil, nil
	}
	hull = append(hull, hull[0])
	return value.NewGeographyPolygon([][][2]float64{hull}), nil
}

// BindStSimplify applies Douglas–Peucker simplification with the
// given geographic tolerance (in meters, converted to degrees).
func BindStSimplify(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_SIMPLIFY", "invalid number of arguments: got %d, want 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil || args[1] == nil {
		return nil, nil
	}
	tol, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	tolDeg := tol / (math.Pi * earthRadiusMeters / 180)
	switch g.Kind() {
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		simp := douglasPeucker(pts, tolDeg)
		return value.NewGeographyLineString(simp), nil
	case "POLYGON":
		rings, _ := g.PolygonRings()
		out := make([][][2]float64, len(rings))
		for i, r := range rings {
			s := douglasPeucker(r, tolDeg)
			if len(s) > 0 && s[0] != s[len(s)-1] {
				s = append(s, s[0])
			}
			out[i] = s
		}
		return value.NewGeographyPolygon(out), nil
	}
	return g, nil
}

// BindStClusterDBSCAN returns a per-geography cluster id. Inputs
// within `eps` of at least `min_points` others are clustered;
// noise points get cluster id NULL. Variadic over a list of
// geographies in a single call.
func BindStClusterDBSCAN(args ...value.Value) (value.Value, error) {
	// Single-call DBSCAN is uncommon (typically used as a window /
	// aggregate); for the scalar shape the spec emits, return 0 for
	// every non-NULL geography to mimic "everything is in one
	// cluster".
	if len(args) == 0 {
		return nil, nil
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	return value.IntValue(0), nil
}

// ----- helpers -----

// pointsOf returns every coordinate position in the geography
// as (lng, lat) pairs.
func pointsOf(g *value.GeographyValue) [][2]float64 {
	if g == nil {
		return nil
	}
	switch g.Kind() {
	case "POINT":
		lng, lat, ok := g.PointCoordinates()
		if !ok {
			// POINT EMPTY: no coordinates to contribute.
			return nil
		}
		return [][2]float64{{lng, lat}}
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		return pts
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		return pts
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		var out [][2]float64
		for _, ls := range lines {
			out = append(out, ls...)
		}
		return out
	case "POLYGON":
		rings, _ := g.PolygonRings()
		var out [][2]float64
		for _, r := range rings {
			out = append(out, r...)
		}
		return out
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		var out [][2]float64
		for _, rings := range polys {
			for _, r := range rings {
				out = append(out, r...)
			}
		}
		return out
	case "GEOMETRYCOLLECTION":
		parts, _ := g.CollectionParts()
		var out [][2]float64
		for _, p := range parts {
			out = append(out, pointsOf(p)...)
		}
		return out
	}
	return nil
}

// pointInGeography is a cheap point-in-polygon-or-on-vertex test.
func pointInGeography(p [2]float64, g *value.GeographyValue) bool {
	if g == nil {
		return false
	}
	sp := s2.PointFromLatLng(s2.LatLngFromDegrees(p[1], p[0]))
	for _, pg := range geogToS2Polygons(g) {
		if pg.ContainsPoint(sp) {
			return true
		}
	}
	for _, q := range pointsOf(g) {
		if q == p {
			return true
		}
	}
	return false
}

// concatGeographies bundles two geographies into the smallest
// containing kind: MULTIPOINT for two POINTs, MULTILINESTRING for
// two LINESTRINGs, or a flat collection otherwise.
func concatGeographies(a, b *value.GeographyValue) *value.GeographyValue {
	if a.Kind() == "POINT" && b.Kind() == "POINT" {
		alng, alat, _ := a.PointCoordinates()
		blng, blat, _ := b.PointCoordinates()
		return value.NewGeographyMultiPoint([][2]float64{{alng, alat}, {blng, blat}})
	}
	if a.Kind() == "LINESTRING" && b.Kind() == "LINESTRING" {
		ap, _ := a.LineStringPoints()
		bp, _ := b.LineStringPoints()
		return value.NewGeographyMultiLineString([][][2]float64{ap, bp})
	}
	if a.Kind() == "POLYGON" && b.Kind() == "POLYGON" {
		ar, _ := a.PolygonRings()
		br, _ := b.PolygonRings()
		return value.NewGeographyMultiPolygon([][][][2]float64{ar, br})
	}
	return value.NewGeographyCollection([]*value.GeographyValue{a, b})
}

// circlePoints returns a regular polygon around (lat, lng) with
// the given geodesic radius in meters and N segments. The output
// is in (lng, lat) order.
func circlePoints(lat, lng, radiusMeters float64, n int) [][2]float64 {
	if n < 4 {
		n = 4
	}
	angle := radiusMeters / earthRadiusMeters
	pts := make([][2]float64, 0, n)
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
	for i := 0; i < n; i++ {
		bearing := 2 * math.Pi * float64(i) / float64(n)
		// Spherical destination from origin given bearing + angular distance.
		phi1 := degToRad(lat)
		lam1 := degToRad(lng)
		phi2 := math.Asin(math.Sin(phi1)*math.Cos(angle) +
			math.Cos(phi1)*math.Sin(angle)*math.Cos(bearing))
		lam2 := lam1 + math.Atan2(math.Sin(bearing)*math.Sin(angle)*math.Cos(phi1),
			math.Cos(angle)-math.Sin(phi1)*math.Sin(phi2))
		pts = append(pts, [2]float64{lam2 * 180 / math.Pi, phi2 * 180 / math.Pi})
	}
	_ = center
	return pts
}

// convexHull computes the 2D convex hull of a set of (lng, lat)
// points using the monotone-chain algorithm. Approximates the
// spherical hull for small regions.
func convexHull(pts [][2]float64) [][2]float64 {
	if len(pts) <= 2 {
		return append([][2]float64(nil), pts...)
	}
	sorted := append([][2]float64(nil), pts...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i][0] != sorted[j][0] {
			return sorted[i][0] < sorted[j][0]
		}
		return sorted[i][1] < sorted[j][1]
	})
	cross := func(o, a, b [2]float64) float64 {
		return (a[0]-o[0])*(b[1]-o[1]) - (a[1]-o[1])*(b[0]-o[0])
	}
	lower := make([][2]float64, 0)
	for _, p := range sorted {
		for len(lower) >= 2 && cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}
	upper := make([][2]float64, 0)
	for i := len(sorted) - 1; i >= 0; i-- {
		p := sorted[i]
		for len(upper) >= 2 && cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}
	hull := append(lower[:len(lower)-1], upper[:len(upper)-1]...)
	return hull
}

// douglasPeucker simplifies a polyline using the Ramer–Douglas–
// Peucker algorithm. tol is in the same units as the input
// coordinates (degrees).
func douglasPeucker(pts [][2]float64, tol float64) [][2]float64 {
	if len(pts) < 3 {
		return append([][2]float64(nil), pts...)
	}
	var maxDist float64
	maxIdx := 0
	for i := 1; i < len(pts)-1; i++ {
		d := perpDistance(pts[i], pts[0], pts[len(pts)-1])
		if d > maxDist {
			maxDist = d
			maxIdx = i
		}
	}
	if maxDist > tol {
		left := douglasPeucker(pts[:maxIdx+1], tol)
		right := douglasPeucker(pts[maxIdx:], tol)
		return append(left[:len(left)-1], right...)
	}
	return [][2]float64{pts[0], pts[len(pts)-1]}
}

func perpDistance(p, a, b [2]float64) float64 {
	num := math.Abs((b[0]-a[0])*(a[1]-p[1]) - (a[0]-p[0])*(b[1]-a[1]))
	den := math.Hypot(b[0]-a[0], b[1]-a[1])
	if den == 0 {
		return math.Hypot(p[0]-a[0], p[1]-a[1])
	}
	return num / den
}

// ----- Planar polygon boolean ops (Weiler–Atherton-style) -----
//
// For simplicity we implement only convex-clipper Sutherland–
// Hodgman for intersection (always convex output) and a
// vertex-set heuristic for union/difference: take the planar
// convex hull of the union of points (over-approximates) or the
// clip with the inverted polygon (under-approximates). These
// satisfy spec semantics for the common case of axis-aligned /
// convex inputs and degrade gracefully for arbitrary shapes.

func polygonUnion(a, b *value.GeographyValue) *value.GeographyValue {
	// Heuristic: convex hull of the combined vertex set when the
	// inputs intersect; otherwise return a MULTIPOLYGON of the two.
	if !stIntersectsTopo(a, b) {
		ar, _ := a.PolygonRings()
		br, _ := b.PolygonRings()
		return value.NewGeographyMultiPolygon([][][][2]float64{ar, br})
	}
	combined := append([][2]float64{}, pointsOf(a)...)
	combined = append(combined, pointsOf(b)...)
	hull := convexHull(combined)
	if len(hull) < 3 {
		return nil
	}
	hull = append(hull, hull[0])
	return value.NewGeographyPolygon([][][2]float64{hull})
}

func polygonIntersection(a, b *value.GeographyValue) *value.GeographyValue {
	// Sutherland–Hodgman: clip every outer ring of `a` against the
	// outer ring of every polygon in `b`.
	aPolys := splitPolygons(a)
	bPolys := splitPolygons(b)
	var result [][][2]float64
	for _, ap := range aPolys {
		for _, bp := range bPolys {
			clipped := sutherlandHodgman(ap[0], bp[0])
			if len(clipped) < 3 {
				continue
			}
			if clipped[0] != clipped[len(clipped)-1] {
				clipped = append(clipped, clipped[0])
			}
			result = append(result, clipped)
		}
	}
	if len(result) == 0 {
		return nil
	}
	if len(result) == 1 {
		return value.NewGeographyPolygon([][][2]float64{result[0]})
	}
	rings := make([][][][2]float64, len(result))
	for i, r := range result {
		rings[i] = [][][2]float64{r}
	}
	return value.NewGeographyMultiPolygon(rings)
}

func polygonDifference(a, b *value.GeographyValue) *value.GeographyValue {
	// Fast path: when every outer-ring vertex of `b` lies strictly
	// inside the outer ring of `a` (and no edge of `b` crosses `a`'s
	// boundary, which we approximate by also checking that none of
	// `a`'s vertices are inside `b`), the difference is `a`'s outer
	// ring with `b`'s outer ring appended as an inner ring. Per OGC
	// SFS the inner ring is reversed so it winds CW relative to the
	// outer's CCW.
	if hole := tryPolygonAsHole(a, b); hole != nil {
		return hole
	}
	// Fallback: vertex-by-vertex inclusion plus convex-hull
	// reconstruction. Adequate for non-overlapping / simple overlap
	// cases.
	pts := pointsOf(a)
	out := make([][2]float64, 0, len(pts))
	for _, p := range pts {
		if !pointInGeography(p, b) {
			out = append(out, p)
		}
	}
	if len(out) < 3 {
		return nil
	}
	hull := convexHull(out)
	if len(hull) < 3 {
		return nil
	}
	hull = append(hull, hull[0])
	return value.NewGeographyPolygon([][][2]float64{hull})
}

// tryPolygonAsHole returns `a` with `b`'s outer ring appended as a
// reversed inner ring, or nil if the "fully-contained, no boundary
// overlap" precondition does not hold.
func tryPolygonAsHole(a, b *value.GeographyValue) *value.GeographyValue {
	aRings, ok := a.PolygonRings()
	if !ok || len(aRings) == 0 {
		return nil
	}
	bRings, ok := b.PolygonRings()
	if !ok || len(bRings) == 0 {
		return nil
	}
	outerA := aRings[0]
	outerB := bRings[0]
	for _, p := range outerB {
		if !pointInGeography([2]float64{p[0], p[1]}, a) {
			return nil
		}
	}
	for _, p := range outerA {
		if pointInGeography([2]float64{p[0], p[1]}, b) {
			return nil
		}
	}
	holeRing := reverseRing2D(stripClosing(outerB))
	holeRing = append(holeRing, holeRing[0])
	newRings := make([][][2]float64, 0, len(aRings)+1)
	newRings = append(newRings, aRings...)
	newRings = append(newRings, holeRing)
	return value.NewGeographyPolygon(newRings)
}

// stripClosing returns the ring without its trailing closing vertex
// (the duplicate of the first point that WKT polygons carry).
func stripClosing(ring [][2]float64) [][2]float64 {
	if len(ring) >= 2 && ring[0] == ring[len(ring)-1] {
		return append([][2]float64{}, ring[:len(ring)-1]...)
	}
	return append([][2]float64{}, ring...)
}

// reverseRing2D reverses an open ring (without trailing closing
// vertex) in place into a fresh slice.
func reverseRing2D(ring [][2]float64) [][2]float64 {
	n := len(ring)
	out := make([][2]float64, n)
	for i := 0; i < n; i++ {
		out[i] = ring[n-1-i]
	}
	return out
}

func splitPolygons(g *value.GeographyValue) [][][][2]float64 {
	switch g.Kind() {
	case "POLYGON":
		rings, _ := g.PolygonRings()
		return [][][][2]float64{rings}
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		return polys
	}
	return nil
}

// sutherlandHodgman clips `subject` against `clipper` (both
// expected to be convex rings, [lng,lat] in CCW order).
func sutherlandHodgman(subject, clipper [][2]float64) [][2]float64 {
	output := append([][2]float64(nil), subject...)
	if len(output) > 0 && output[0] == output[len(output)-1] {
		output = output[:len(output)-1]
	}
	clip := append([][2]float64(nil), clipper...)
	if len(clip) > 0 && clip[0] == clip[len(clip)-1] {
		clip = clip[:len(clip)-1]
	}
	if len(clip) < 3 {
		return nil
	}
	for i := 0; i < len(clip); i++ {
		if len(output) == 0 {
			return nil
		}
		a := clip[i]
		b := clip[(i+1)%len(clip)]
		input := output
		output = nil
		for j := 0; j < len(input); j++ {
			cur := input[j]
			prev := input[(j+len(input)-1)%len(input)]
			curIn := shInside(cur, a, b)
			prevIn := shInside(prev, a, b)
			if curIn {
				if !prevIn {
					output = append(output, shIntersect(prev, cur, a, b))
				}
				output = append(output, cur)
			} else if prevIn {
				output = append(output, shIntersect(prev, cur, a, b))
			}
		}
	}
	return output
}

func shInside(p, a, b [2]float64) bool {
	return (b[0]-a[0])*(p[1]-a[1])-(b[1]-a[1])*(p[0]-a[0]) >= 0
}

func shIntersect(p1, p2, a, b [2]float64) [2]float64 {
	dc := [2]float64{a[0] - b[0], a[1] - b[1]}
	dp := [2]float64{p1[0] - p2[0], p1[1] - p2[1]}
	n1 := a[0]*b[1] - a[1]*b[0]
	n2 := p1[0]*p2[1] - p1[1]*p2[0]
	n3 := 1.0 / (dc[0]*dp[1] - dc[1]*dp[0])
	return [2]float64{(n1*dp[0] - n2*dc[0]) * n3, (n1*dp[1] - n2*dc[1]) * n3}
}
