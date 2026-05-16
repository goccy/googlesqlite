package geography

import (
	"github.com/golang/geo/s2"

	"github.com/goccy/googlesqlite/internal/value"
)

// geographyToCells returns a coarse S2 CellUnion covering the
// geography. Used to short-circuit topological predicates with a
// cheap pre-filter.
func geographyToCells(g *value.GeographyValue) s2.CellUnion {
	cov := &s2.RegionCoverer{MinLevel: 0, MaxLevel: 22, MaxCells: 32}
	switch g.Kind() {
	case "POINT", "MULTIPOINT":
		pts := geogToS2Points(g)
		out := s2.CellUnion{}
		for _, p := range pts {
			out = append(out, s2.CellFromPoint(p).ID())
		}
		return out
	case "LINESTRING", "MULTILINESTRING":
		out := s2.CellUnion{}
		for _, pl := range geogToS2Polylines(g) {
			out = append(out, cov.Covering(pl)...)
		}
		return out
	case "POLYGON", "MULTIPOLYGON":
		out := s2.CellUnion{}
		for _, pg := range geogToS2Polygons(g) {
			out = append(out, cov.Covering(pg)...)
		}
		return out
	}
	return nil
}

// stIntersectsTopo returns whether two geographies have any
// shared point. Implementation: combine the two via cell-coverings
// then fall back to per-shape pairwise checks.
func stIntersectsTopo(a, b *value.GeographyValue) bool {
	if a == nil || b == nil {
		return false
	}
	ca := geographyToCells(a)
	cb := geographyToCells(b)
	if !ca.Intersects(cb) {
		return false
	}
	// Fine-grained checks: point-in-polygon and polyline crossing.
	aPolys := geogToS2Polygons(a)
	bPolys := geogToS2Polygons(b)
	aLines := geogToS2Polylines(a)
	bLines := geogToS2Polylines(b)
	aPoints := geogToS2Points(a)
	bPoints := geogToS2Points(b)
	// Point vs polygon
	for _, p := range aPoints {
		for _, pg := range bPolys {
			if pg.ContainsPoint(p) {
				return true
			}
		}
	}
	for _, p := range bPoints {
		for _, pg := range aPolys {
			if pg.ContainsPoint(p) {
				return true
			}
		}
	}
	// Polygon vs polygon
	for _, pa := range aPolys {
		for _, pb := range bPolys {
			if pa.Intersects(pb) {
				return true
			}
		}
	}
	// Line vs polygon
	for _, pl := range aLines {
		for _, pg := range bPolys {
			for i := 0; i < pl.NumEdges(); i++ {
				e := pl.Edge(i)
				if pg.ContainsPoint(e.V0) || pg.ContainsPoint(e.V1) {
					return true
				}
			}
		}
	}
	for _, pl := range bLines {
		for _, pg := range aPolys {
			for i := 0; i < pl.NumEdges(); i++ {
				e := pl.Edge(i)
				if pg.ContainsPoint(e.V0) || pg.ContainsPoint(e.V1) {
					return true
				}
			}
		}
	}
	// Line vs line via edge crossing.
	for _, la := range aLines {
		for _, lb := range bLines {
			if la.Intersects(lb) {
				return true
			}
		}
	}
	// Point-vs-point coincidence.
	for _, pa := range aPoints {
		for _, pb := range bPoints {
			if pa == pb {
				return true
			}
		}
	}
	// Point on a polyline.
	for _, p := range aPoints {
		for _, pl := range bLines {
			for i := 0; i < pl.NumEdges(); i++ {
				e := pl.Edge(i)
				if e.V0 == p || e.V1 == p {
					return true
				}
			}
		}
	}
	for _, p := range bPoints {
		for _, pl := range aLines {
			for i := 0; i < pl.NumEdges(); i++ {
				e := pl.Edge(i)
				if e.V0 == p || e.V1 == p {
					return true
				}
			}
		}
	}
	return false
}

// BindStIntersects returns whether the two geographies share any
// point.
func BindStIntersects(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_INTERSECTS", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	return value.BoolValue(stIntersectsTopo(a, b)), nil
}

// BindStDisjoint is the complement of ST_INTERSECTS.
func BindStDisjoint(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_DISJOINT", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	return value.BoolValue(!stIntersectsTopo(a, b)), nil
}

// stContainsTopo returns whether `a` topologically contains `b`:
// every point of `b` lies inside `a`. Implemented via
// per-element membership checks.
func stContainsTopo(a, b *value.GeographyValue) bool {
	if a == nil || b == nil {
		return false
	}
	aPolys := geogToS2Polygons(a)
	if len(aPolys) == 0 {
		return false
	}
	for _, p := range geogToS2Points(b) {
		if !anyPolygonContains(aPolys, p) {
			return false
		}
	}
	for _, pl := range geogToS2Polylines(b) {
		n := pl.NumEdges()
		if n == 0 {
			if len(*pl) > 0 && !anyPolygonContains(aPolys, (*pl)[0]) {
				return false
			}
			continue
		}
		for i := 0; i < n; i++ {
			e := pl.Edge(i)
			if !anyPolygonContains(aPolys, e.V0) || !anyPolygonContains(aPolys, e.V1) {
				return false
			}
		}
	}
	for _, pg := range geogToS2Polygons(b) {
		for i := 0; i < pg.NumLoops(); i++ {
			loop := pg.Loop(i)
			for j := 0; j < loop.NumVertices(); j++ {
				if !anyPolygonContains(aPolys, loop.Vertex(j)) {
					return false
				}
			}
		}
	}
	return true
}

func anyPolygonContains(pgs []*s2.Polygon, p s2.Point) bool {
	for _, pg := range pgs {
		if pg.ContainsPoint(p) {
			return true
		}
	}
	return false
}

// BindStContains returns whether `a` contains `b`.
func BindStContains(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_CONTAINS", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	return value.BoolValue(stContainsTopo(a, b)), nil
}

// BindStWithin returns whether `a` is within `b` (inverse of
// ST_CONTAINS).
func BindStWithin(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_WITHIN", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	return value.BoolValue(stContainsTopo(b, a)), nil
}

// BindStCovers / BindStCoveredBy reuse the same vertex-membership
// check as Contains: for our supported geometries the distinction
// between Covers and Contains (boundary handling) collapses
// because boundary points are reported as "contained" by s2.
func BindStCovers(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_COVERS", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	return value.BoolValue(stContainsTopo(a, b)), nil
}

func BindStCoveredBy(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_COVEREDBY", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	return value.BoolValue(stContainsTopo(b, a)), nil
}

// BindStTouches reports topological tangency: the two geographies
// share at least a boundary point but no interior point.
// Approximated here as Intersects && !Contains either direction.
func BindStTouches(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_TOUCHES", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	if !stIntersectsTopo(a, b) {
		return value.BoolValue(false), nil
	}
	if stContainsTopo(a, b) || stContainsTopo(b, a) {
		return value.BoolValue(false), nil
	}
	return value.BoolValue(true), nil
}

// BindStDWithin reports whether the two geographies lie within
// `distance` meters of each other.
func BindStDWithin(args ...value.Value) (value.Value, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, sqError("ST_DWITHIN", "invalid number of arguments: got %d, want between 3 and 4", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil || args[2] == nil {
		return nil, nil
	}
	dist, err := args[2].ToFloat64()
	if err != nil {
		return nil, err
	}
	d, err := geographyDistanceMeters(a, b)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(d <= dist), nil
}

// geographyDistanceMeters returns the minimum geodesic distance
// between two arbitrary geographies. Falls back to a "any-point
// of A vs any-point of B" minimum when either side is not a
// polygon/polyline/point.
func geographyDistanceMeters(a, b *value.GeographyValue) (float64, error) {
	if a == nil || b == nil {
		return 0, nil
	}
	if stIntersectsTopo(a, b) {
		return 0, nil
	}
	// Collect a representative point set for each side.
	collect := func(g *value.GeographyValue) []s2.Point {
		var out []s2.Point
		out = append(out, geogToS2Points(g)...)
		for _, pl := range geogToS2Polylines(g) {
			for i := 0; i < pl.NumEdges(); i++ {
				e := pl.Edge(i)
				out = append(out, e.V0, e.V1)
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
	pa := collect(a)
	pb := collect(b)
	if len(pa) == 0 || len(pb) == 0 {
		return 0, nil
	}
	minAngle := pa[0].Distance(pb[0])
	for _, x := range pa {
		for _, y := range pb {
			d := x.Distance(y)
			if d < minAngle {
				minAngle = d
			}
		}
	}
	return distanceAngleToMeters(minAngle), nil
}

// BindStMaxDistance returns the maximum geodesic distance
// between any two vertices of the two geographies, in meters.
func BindStMaxDistance(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, sqError("ST_MAXDISTANCE", "invalid number of arguments: got %d, want 2", len(args))
	}
	a := geographyArg(args[0])
	b := geographyArg(args[1])
	if a == nil || b == nil {
		return nil, nil
	}
	collect := func(g *value.GeographyValue) []s2.Point {
		var out []s2.Point
		out = append(out, geogToS2Points(g)...)
		for _, pl := range geogToS2Polylines(g) {
			for i := 0; i < pl.NumEdges(); i++ {
				e := pl.Edge(i)
				out = append(out, e.V0, e.V1)
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
	pa := collect(a)
	pb := collect(b)
	if len(pa) == 0 || len(pb) == 0 {
		return value.FloatValue(0), nil
	}
	maxAngle := pa[0].Distance(pb[0])
	for _, x := range pa {
		for _, y := range pb {
			d := x.Distance(y)
			if d > maxAngle {
				maxAngle = d
			}
		}
	}
	return value.FloatValue(distanceAngleToMeters(maxAngle)), nil
}
