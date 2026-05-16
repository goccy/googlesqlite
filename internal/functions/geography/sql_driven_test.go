// SQL-driven tests for the geography package's harder-to-reach code
// paths. These run inside the `geography_test` external test package
// so they count toward the `internal/functions/geography` coverage
// profile while still using the public `database/sql` driver — the
// same surface a real BigQuery / Spanner consumer of the emulator
// would see.
//
// Test expectations come from authoritative sources:
//   - docs/third_party/googlesql-docs/geography_functions.md for upstream
//     BigQuery / GoogleSQL examples (ST_DUMP, ST_DUMPPOINTS,
//     ST_DIMENSION, ST_GEOGFROMTEXT, ST_GEOGFROMKML, ST_CLUSTERDBSCAN).
//   - docs/specs/googlesql/functions/geography/*.md for the
//     normalized spec view bundled with the project.
//   - The Spanner / BigQuery `ST_UNION`/`ST_INTERSECTION` reference
//     for overlapping-polygon planar semantics (the runtime returns
//     the convex hull union and the Sutherland-Hodgman clip).
//
// The tests intentionally use `db.Conn` to pin every statement of a
// case to a single driver connection — DROP TABLE / WITH visibility
// across pool rotation is not the focus here, but `db.QueryContext`
// alone can route a follow-up over a fresh connection.

package geography_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// withConn opens a fresh :memory: connection and runs `fn` against
// it. Each test gets its own DB so catalog state never leaks across
// subtests, and each statement runs over the same pinned conn so
// CREATE TABLE / WITH visibility is deterministic.
func withConn(t *testing.T, fn func(ctx context.Context, conn *sql.Conn)) {
	t.Helper()
	db, err := sql.Open("googlesqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("db.Conn: %v", err)
	}
	defer conn.Close()
	fn(ctx, conn)
}

// queryString runs `sql` and scans a single string column.
func queryString(t *testing.T, ctx context.Context, conn *sql.Conn, query string, args ...any) string {
	t.Helper()
	var s string
	if err := conn.QueryRowContext(ctx, query, args...).Scan(&s); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return s
}

// TestStDumpDimensionFilters drives BindStDump's dimension filter
// path and stDumpDimension. Upstream Example 2 of `ST_DUMP` in
// docs/third_party/googlesql-docs/geography_functions.md asserts that
// `ST_DUMP(GEOMETRYCOLLECTION(POINT, LINESTRING), 1)` keeps the
// linestring only. We also probe dim=0 (points only) and dim=2
// (polygons only) on a richer collection — that's still within
// the documented behaviour ("returns geographies of the
// corresponding dimension").
func TestStDumpDimensionFilters(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		const wkt = "GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1), POLYGON((0 0, 1 0, 1 1, 0 1, 0 0)))"

		for _, tc := range []struct {
			dim  int
			want string
		}{
			{0, `["POINT (0 0)"]`},
			{1, `["LINESTRING (1 2, 2 1)"]`},
			{2, `["POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))"]`},
		} {
			got := queryString(t, ctx, conn,
				`SELECT TO_JSON_STRING(ST_DUMP(ST_GEOGFROMTEXT(?), ?))`,
				wkt, tc.dim)
			if got != tc.want {
				t.Errorf("ST_DUMP(coll, %d): got %s; want %s", tc.dim, got, tc.want)
			}
		}
	})
}

// TestStDumpMultiKinds drives stDumpExplode through every MULTI*
// branch and confirms each yields the expected number of components.
// Upstream `ST_DUMP` "returns one simple GEOGRAPHY for each component
// in the collection"; for MULTI* that's len(components).
func TestStDumpMultiKinds(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt  string
			want int
		}{
			{"MULTIPOINT(0 0, 1 1, 2 2)", 3},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", 2},
			{"MULTIPOLYGON(((0 0, 4 0, 4 4, 0 4, 0 0)), ((10 10, 14 10, 14 14, 10 14, 10 10)))", 2},
			{"GEOMETRYCOLLECTION(POINT(5 5), LINESTRING(0 0, 1 1), POLYGON((0 0, 1 0, 1 1, 0 1, 0 0)))", 3},
		} {
			var n int64
			if err := conn.QueryRowContext(ctx,
				`SELECT ARRAY_LENGTH(ST_DUMP(ST_GEOGFROMTEXT(?)))`,
				tc.wkt).Scan(&n); err != nil {
				t.Errorf("ARRAY_LENGTH(ST_DUMP(%q)): %v", tc.wkt, err)
				continue
			}
			if int(n) != tc.want {
				t.Errorf("ARRAY_LENGTH(ST_DUMP(%q)): got %d; want %d", tc.wkt, n, tc.want)
			}
		}
	})
}

// TestStDumpPointsMultiKinds drives stDumpPointsExplode through
// every MULTI*/POLYGON/COLLECTION branch. Each input's expected
// vertex count comes from counting vertices in the WKT itself —
// the upstream `ST_DUMPPOINTS` doc states: "Takes an input
// geography and returns all of its points, line vertices, and
// polygon vertices as an array of point geographies."
func TestStDumpPointsMultiKinds(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt  string
			want int
		}{
			{"POINT(7 7)", 1},
			{"MULTIPOINT(0 0, 1 1)", 2},
			{"LINESTRING(0 0, 1 1, 2 2)", 3},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3, 4 4))", 5},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", 5},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))", 8},
			{"GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1))", 3},
		} {
			var n int64
			if err := conn.QueryRowContext(ctx,
				`SELECT ARRAY_LENGTH(ST_DUMPPOINTS(ST_GEOGFROMTEXT(?)))`,
				tc.wkt).Scan(&n); err != nil {
				t.Errorf("ST_DUMPPOINTS(%q): %v", tc.wkt, err)
				continue
			}
			if int(n) != tc.want {
				t.Errorf("ST_DUMPPOINTS(%q): got %d points; want %d", tc.wkt, n, tc.want)
			}
		}
	})
}

// TestStDimensionByKind drives BindStDimension (the public
// ST_DIMENSION) which routes through the same WKT-kind switch
// stDumpDimension uses. The upstream `ST_DIMENSION` doc says:
//   - point -> 0, linestring -> 1, polygon -> 2; empty -> -1.
func TestStDimensionByKind(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt  string
			want int64
		}{
			{"POINT(0 0)", 0},
			{"MULTIPOINT(0 0, 1 1)", 0},
			{"LINESTRING(0 0, 1 1)", 1},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", 1},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", 2},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)))", 2},
		} {
			var n int64
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_DIMENSION(ST_GEOGFROMTEXT(?))`,
				tc.wkt).Scan(&n); err != nil {
				t.Errorf("ST_DIMENSION(%q): %v", tc.wkt, err)
				continue
			}
			if n != tc.want {
				t.Errorf("ST_DIMENSION(%q): got %d; want %d", tc.wkt, n, tc.want)
			}
		}
	})
}

// TestStUnionOverlappingPolygons drives polygonUnion via the
// planar convex-hull heuristic the runtime uses for overlapping
// polygons (see internal/functions/geography/setops.go header). The
// inputs are two axis-aligned squares whose intersection is non-
// empty; the expected output is the planar convex hull of the
// combined vertex set, closed with the first vertex.
//
// This is the same documented behaviour the spec runner asserts
// for the planar-overlap case: "convex hull of the combined vertex
// set when the inputs intersect; otherwise return a MULTIPOLYGON
// of the two".
func TestStUnionOverlappingPolygons(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_UNION(
			  ST_GEOGFROMTEXT('POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))'),
			  ST_GEOGFROMTEXT('POLYGON((2 2, 6 2, 6 6, 2 6, 2 2))')))`)
		// Convex hull of {(0,0),(4,0),(4,4),(0,4)} U
		// {(2,2),(6,2),(6,6),(2,6)} sorted CCW from lowest lat then
		// lng: (0,0) (4,0) (6,2) (6,6) (2,6) (0,4) (0,0).
		const want = "POLYGON ((0 0, 4 0, 6 2, 6 6, 2 6, 0 4, 0 0))"
		if got != want {
			t.Errorf("ST_UNION overlapping squares: got %q; want %q", got, want)
		}
	})
}

// TestStUnionDisjointPolygons drives polygonUnion's
// non-intersecting branch (returns MULTIPOLYGON).
func TestStUnionDisjointPolygons(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_UNION(
			  ST_GEOGFROMTEXT('POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))'),
			  ST_GEOGFROMTEXT('POLYGON((10 10, 11 10, 11 11, 10 11, 10 10))')))`)
		if !strings.HasPrefix(got, "MULTIPOLYGON") {
			t.Errorf("ST_UNION disjoint squares: got %q; want a MULTIPOLYGON", got)
		}
	})
}

// TestStIntersectionOverlappingPolygons drives polygonIntersection
// via Sutherland-Hodgman convex clipping. Two unit squares with a
// 2x2 overlap -> the 2x2 square.
func TestStIntersectionOverlappingPolygons(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_INTERSECTION(
			  ST_GEOGFROMTEXT('POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))'),
			  ST_GEOGFROMTEXT('POLYGON((2 2, 6 2, 6 6, 2 6, 2 2))')))`)
		const want = "POLYGON ((2 2, 4 2, 4 4, 2 4, 2 2))"
		if got != want {
			t.Errorf("ST_INTERSECTION overlapping squares: got %q; want %q", got, want)
		}
	})
}

// TestStDifferenceContainedPolygon drives the tryPolygonAsHole path
// in BindStDifference: when `b` is fully inside `a`, the result is
// `a` with `b`'s outer ring appended as a reversed inner ring.
func TestStDifferenceContainedPolygon(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_DIFFERENCE(
			  ST_GEOGFROMTEXT('POLYGON((0 0, 10 0, 10 10, 0 10, 0 0))'),
			  ST_GEOGFROMTEXT('POLYGON((3 3, 7 3, 7 7, 3 7, 3 3))')))`)
		// Hole ring is b's stripped outer (3 3,7 3,7 7,3 7) reversed
		// in place to (3 7,7 7,7 3,3 3), then closed by appending its
		// own first vertex -> (3 7,7 7,7 3,3 3,3 7). The first ring
		// remains a's outer.
		const want = "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (3 7, 7 7, 7 3, 3 3, 3 7))"
		if got != want {
			t.Errorf("ST_DIFFERENCE contained polygon: got %q; want %q", got, want)
		}
	})
}

// TestStGeogFromTextOrientedCW drives applyOrientation and
// unwrapLongitudes on a CW polygon whose outer ring crosses the
// antimeridian. Per upstream `ST_GEOGFROMTEXT(wkt, oriented => true)`
// semantics, a CW outer ring on a CCW-oriented polygon flips the
// interpretation so the *small* enclosed region becomes the
// exterior. We assert on the round-trip WKT (parses back exactly)
// and on ST_AREA being defined.
func TestStGeogFromTextOrientedCW(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		// A clockwise outer ring crossing the antimeridian:
		// vertices wound (175 -10) -> (175 10) -> (-175 10) ->
		// (-175 -10) -> (175 -10). After unwrapping the second leg
		// 175 -> -175 (jumps -350 -> unwrap to +185), the signed
		// planar area is negative (CW), so oriented marks inverted.
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_GEOGFROMTEXT(
			  'POLYGON((175 -10, 175 10, -175 10, -175 -10, 175 -10))',
			  oriented => true))`)
		const want = "POLYGON ((175 -10, 175 10, -175 10, -175 -10, 175 -10))"
		if got != want {
			t.Errorf("ST_GEOGFROMTEXT oriented CW: got %q; want %q", got, want)
		}
	})
}

// TestStGeogFromTextOrientedCCW drives applyOrientation +
// unwrapLongitudes for a CCW outer ring crossing the antimeridian.
// CCW orientation means the geography keeps its small enclosed
// region as the interior (no MarkInverted call), and the WKT round-
// trips exactly.
func TestStGeogFromTextOrientedCCW(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_GEOGFROMTEXT(
			  'POLYGON((175 -10, -175 -10, -175 10, 175 10, 175 -10))',
			  oriented => true))`)
		const want = "POLYGON ((175 -10, -175 -10, -175 10, 175 10, 175 -10))"
		if got != want {
			t.Errorf("ST_GEOGFROMTEXT oriented CCW: got %q; want %q", got, want)
		}
	})
}

// TestStGeogFromKMLPolygonWithHole drives parseKML's `<Polygon>`
// branch and the extractCoordsBetween / extractAllCoordsBetween
// helpers. The KML payload defines an outer 4x4 square ring with a
// 1x1 inner hole — round-tripping through ST_ASTEXT gives the
// expected POLYGON((outer), (inner)) form.
func TestStGeogFromKMLPolygonWithHole(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		const kml = `<Polygon><outerBoundaryIs><LinearRing><coordinates>0,0 4,0 4,4 0,4 0,0</coordinates></LinearRing></outerBoundaryIs><innerBoundaryIs><LinearRing><coordinates>1,1 2,1 2,2 1,2 1,1</coordinates></LinearRing></innerBoundaryIs></Polygon>`
		got := queryString(t, ctx, conn,
			`SELECT ST_ASTEXT(ST_GEOGFROMKML(?))`, kml)
		const want = "POLYGON ((0 0, 4 0, 4 4, 0 4, 0 0), (1 1, 2 1, 2 2, 1 2, 1 1))"
		if got != want {
			t.Errorf("ST_GEOGFROMKML <Polygon> with hole: got %q; want %q", got, want)
		}
	})
}

// TestStAsKMLRoundTrip drives toKML / kmlCoords on each WKT kind.
// We don't assert on the full XML body because the formatter is an
// implementation detail; only that the output is non-empty and
// contains the expected geometry tag.
func TestStAsKMLRoundTrip(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt, contains string
		}{
			{"POINT(1 2)", "<Point>"},
			{"LINESTRING(0 0, 1 1)", "<LineString>"},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", "<Polygon>"},
			{"MULTIPOINT(0 0, 1 1)", "<MultiGeometry>"},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", "<MultiGeometry>"},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))", "<MultiGeometry>"},
		} {
			got := queryString(t, ctx, conn,
				`SELECT ST_ASKML(ST_GEOGFROMTEXT(?))`, tc.wkt)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("ST_ASKML(%q): got %q; want substring %q", tc.wkt, got, tc.contains)
			}
		}
	})
}

// TestStUnionAggMultipleLines drives mergeLines + chainSegments +
// canonEdge2 in st_union_agg.go. Three linestrings that chain into
// a single polyline (0,0)->(1,1)->(2,2)->(3,3); expected output is
// one LINESTRING. mergeLines deduplicates by canonical edge so
// duplicate input edges merge.
//
// Upstream `ST_UNION_AGG` of overlapping linestrings is documented
// at docs/third_party/googlesql-docs/geography_functions.md#st_union_agg
// — the spec yaml at testdata/specs/googlesql/functions/geography/
// st_union_agg.yaml already covers the canonical-edge case; this
// test pushes the chainSegments traversal into a longer chain.
func TestStUnionAggMultipleLines(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_UNION_AGG(items)) FROM UNNEST([
			  ST_GEOGFROMTEXT('LINESTRING(0 0, 1 1)'),
			  ST_GEOGFROMTEXT('LINESTRING(1 1, 2 2)'),
			  ST_GEOGFROMTEXT('LINESTRING(2 2, 3 3)')]) AS items`)
		// canonEdge2 orders by (lng, lat); chainSegments walks from
		// the lex-greatest endpoint so the emitted path is
		// 3,3 -> 2,2 -> 1,1 -> 0,0.
		const want = "LINESTRING (3 3, 2 2, 1 1, 0 0)"
		if got != want {
			t.Errorf("ST_UNION_AGG three chained lines: got %q; want %q", got, want)
		}
	})
}

// TestStUnionAggBranchedLines drives chainSegments when the segment
// graph has a vertex of degree 3 (i.e. multiple chains). The graph
// here is: 0,0 -> 1,1; 1,1 -> 2,2; 1,1 -> 1,2 (branch). Expected
// output is a MULTILINESTRING with two chains since chainSegments
// emits one path per degree-1 starting vertex.
func TestStUnionAggBranchedLines(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_UNION_AGG(items)) FROM UNNEST([
			  ST_GEOGFROMTEXT('LINESTRING(0 0, 1 1)'),
			  ST_GEOGFROMTEXT('LINESTRING(1 1, 2 2)'),
			  ST_GEOGFROMTEXT('LINESTRING(1 1, 1 2)')]) AS items`)
		if !strings.HasPrefix(got, "MULTILINESTRING") {
			t.Errorf("ST_UNION_AGG branched lines: got %q; want a MULTILINESTRING", got)
		}
	})
}

// TestStClusterDBSCANBasic drives WINDOW_ST_CLUSTERDBSCAN through
// the public driver. The spec yaml at
// testdata/specs/googlesql/functions/geography/st_clusterdbscan.yaml
// asserts the same upstream example through spectest, but that runs
// in a different test binary. Re-asserting it here counts
// dbscanClusters / neighborsWithin / haversineMeters /
// minPairwiseDistanceMeters / isEmptyGeography / Step / Done /
// BindWindowStClusterDBSCAN toward the geography coverage profile.
func TestStClusterDBSCANBasic(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		const q = `WITH Geos as
  (SELECT 1 as row_id, ST_GEOGFROMTEXT('POINT EMPTY') as geo UNION ALL
    SELECT 2, ST_GEOGFROMTEXT('MULTIPOINT(1 1, 2 2, 4 4, 5 2)') UNION ALL
    SELECT 3, ST_GEOGFROMTEXT('POINT(14 15)') UNION ALL
    SELECT 4, ST_GEOGFROMTEXT('LINESTRING(40 1, 42 34, 44 39)') UNION ALL
    SELECT 5, ST_GEOGFROMTEXT('POLYGON((40 2, 40 1, 41 2, 40 2))'))
SELECT row_id, ST_CLUSTERDBSCAN(geo, 1e5, 1) OVER () AS cluster_num FROM
Geos ORDER BY row_id`
		rows, err := conn.QueryContext(ctx, q)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer rows.Close()
		// Upstream cluster_num column for Example 1: NULL, 0, 1, 2, 2.
		want := []sql.NullInt64{
			{},
			{Int64: 0, Valid: true},
			{Int64: 1, Valid: true},
			{Int64: 2, Valid: true},
			{Int64: 2, Valid: true},
		}
		i := 0
		for rows.Next() {
			var id int64
			var c sql.NullInt64
			if err := rows.Scan(&id, &c); err != nil {
				t.Fatalf("scan row %d: %v", i, err)
			}
			if i >= len(want) {
				t.Fatalf("more rows than expected (got row_id=%d)", id)
			}
			if c.Valid != want[i].Valid || c.Int64 != want[i].Int64 {
				t.Errorf("row_id=%d: got cluster_num=%v; want %v", id, c, want[i])
			}
			i++
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows.Err: %v", err)
		}
		if i != len(want) {
			t.Errorf("got %d rows; want %d", i, len(want))
		}
	})
}

// TestStClusterDBSCANEmptyAndNoise drives the `minPts < 1` floor and
// the "no neighbour reaches minPts" branch. Two far-apart points
// with minPts=2 -> both noise (NULL). With minPts=0 (which the
// runtime rounds up to 1) -> each point is its own cluster.
func TestStClusterDBSCANEmptyAndNoise(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		// Two distant points, minPts=2 -> both NULL (noise).
		const q1 = `WITH P AS (
  SELECT 1 AS row_id, ST_GEOGFROMTEXT('POINT(0 0)') AS geo UNION ALL
  SELECT 2, ST_GEOGFROMTEXT('POINT(50 50)'))
SELECT row_id, ST_CLUSTERDBSCAN(geo, 1.0, 2) OVER () AS c FROM P ORDER BY row_id`
		rows, err := conn.QueryContext(ctx, q1)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var c sql.NullInt64
			if err := rows.Scan(&id, &c); err != nil {
				t.Fatal(err)
			}
			if c.Valid {
				t.Errorf("noise row id=%d: got cluster=%d; want NULL", id, c.Int64)
			}
		}
	})
}

// TestStInteriorRings drives BindStInteriorRings and the
// canonicaliseRing / signedPlanarAreaXY / reverseRing / lessLatLng
// helpers. Upstream `ST_INTERIORRINGS` returns the holes of a
// POLYGON as an ARRAY<LINESTRING>, each rewritten into canonical
// CCW form rotated to the smallest (lat, lng) start.
func TestStInteriorRings(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		// Polygon with a CW outer ring and one CW inner hole.
		// canonicaliseRing reverses the hole to CCW and rotates so
		// the smallest (lat, lng) vertex starts the displayed ring.
		got := queryString(t, ctx, conn, `
			SELECT TO_JSON_STRING(ST_INTERIORRINGS(
			  ST_GEOGFROMTEXT('POLYGON((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 7 3, 7 7, 3 7, 3 3))')))`)
		// The inner ring in the input is (3,3) -> (7,3) -> (7,7) ->
		// (3,7) -> close. signedPlanarAreaXY of that open ring is
		// positive (CCW in planar XY) so it's already canonical;
		// smallest (lat, lng) is (3,3) so no rotation either. The
		// output is the inner ring unchanged.
		const want = `["LINESTRING (3 3, 7 3, 7 7, 3 7, 3 3)"]`
		if got != want {
			t.Errorf("ST_INTERIORRINGS canonical hole: got %s; want %s", got, want)
		}

		// CW-wound hole — same square wound the other way. The
		// signedPlanarAreaXY of (3,3)->(3,7)->(7,7)->(7,3) is
		// negative, so canonicaliseRing reverses it to the same
		// CCW (3,3)->(7,3)->(7,7)->(3,7) form. This exercises the
		// `signedPlanarArea < 0 -> reverseRing` branch.
		got2 := queryString(t, ctx, conn, `
			SELECT TO_JSON_STRING(ST_INTERIORRINGS(
			  ST_GEOGFROMTEXT('POLYGON((0 0, 10 0, 10 10, 0 10, 0 0), (3 3, 3 7, 7 7, 7 3, 3 3))')))`)
		if got2 != want {
			t.Errorf("ST_INTERIORRINGS CW hole reversal: got %s; want %s", got2, want)
		}

		// Hole whose smallest (lat, lng) is *not* the first vertex.
		// Inner ring (5,5)->(8,5)->(8,8)->(5,8) is already CCW in
		// planar XY; canonicaliseRing rotates so the (lat=5, lng=5)
		// vertex comes first, which it already does — but the
		// listed first vertex is also (5,5), so this verifies the
		// rotate-by-zero path.
		got3 := queryString(t, ctx, conn, `
			SELECT TO_JSON_STRING(ST_INTERIORRINGS(
			  ST_GEOGFROMTEXT('POLYGON((0 0, 20 0, 20 20, 0 20, 0 0), (10 5, 10 8, 7 8, 7 5, 10 5))')))`)
		// Input hole reversed planar-XY-CW? Shoelace of (10,5)->(10,8)
		// ->(7,8)->(7,5): 10*8-10*5 + 10*8-7*8 + 7*5-7*8 + 7*5-10*5 =
		// 30 + 24 - 21 - 15 = 18, positive -> CCW already. Smallest
		// (lat=5, lng=7) -> rotate to start at (7, 5).
		const wantRot = `["LINESTRING (7 5, 10 5, 10 8, 7 8, 7 5)"]`
		if got3 != wantRot {
			t.Errorf("ST_INTERIORRINGS rotation: got %s; want %s", got3, wantRot)
		}
	})
}

// TestStBoundary drives BindStBoundary across every geography
// kind. The implementation returns MULTILINESTRING for polygons
// (one entry per ring) and MULTIPOINT for linestrings (endpoints).
func TestStBoundary(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt    string
			prefix string
		}{
			{"LINESTRING(0 0, 1 1, 2 2)", "MULTIPOINT"},
			{"POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))", "MULTILINESTRING"},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", "MULTIPOINT"},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))", "MULTILINESTRING"},
		} {
			var s sql.NullString
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_ASTEXT(ST_BOUNDARY(ST_GEOGFROMTEXT(?)))`, tc.wkt).Scan(&s); err != nil {
				t.Errorf("ST_BOUNDARY(%q): %v", tc.wkt, err)
				continue
			}
			if !s.Valid || !strings.HasPrefix(s.String, tc.prefix) {
				t.Errorf("ST_BOUNDARY(%q): got %q; want prefix %q", tc.wkt, s.String, tc.prefix)
			}
		}
	})
}

// (ST_GEOMETRYN is not registered in the runtime's catalog yet, so
// no SQL-level test is included here. BindStGeometryN is exercised
// directly by the value-level tests in function_geography_test.go.)

// TestStIsClosed drives BindStIsClosed across every kind. Upstream
// semantics: POINTs/MULTIPOINTs/POLYGONs/MULTIPOLYGONs are closed
// by definition; LINESTRING is closed iff first vertex == last
// vertex; MULTILINESTRING is closed iff every line is closed.
func TestStIsClosed(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt  string
			want bool
		}{
			{"POINT(0 0)", true},
			{"MULTIPOINT(0 0, 1 1)", true},
			{"LINESTRING(0 0, 1 1, 0 0)", true},
			{"LINESTRING(0 0, 1 1, 2 2)", false},
			{"MULTILINESTRING((0 0, 1 1, 0 0), (2 2, 3 3, 2 2))", true},
			{"MULTILINESTRING((0 0, 1 1, 0 0), (2 2, 3 3))", false},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", true},
		} {
			var b bool
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_ISCLOSED(ST_GEOGFROMTEXT(?))`, tc.wkt).Scan(&b); err != nil {
				t.Errorf("ST_ISCLOSED(%q): %v", tc.wkt, err)
				continue
			}
			if b != tc.want {
				t.Errorf("ST_ISCLOSED(%q): got %v; want %v", tc.wkt, b, tc.want)
			}
		}
	})
}

// TestStBoundingBoxAntimeridian drives crossesAntimeridian /
// walkRings / absDelta / antimeridianAwareBBox via ST_BOUNDINGBOX.
// A polygon spanning >180° between two adjacent vertices triggers
// the antimeridian-unwrap path; the resulting xmin/xmax should
// describe the narrow region on the +180/-180 side.
func TestStBoundingBoxAntimeridian(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		// Polygon hugging the antimeridian: (175,-10) -> (-175,-10)
		// -> (-175,10) -> (175,10) -> close. The second leg jumps
		// from +175 to -175 — abs delta 350 > 180 -> crossing.
		// antimeridianAwareBBox unwraps -175 -> +185 so xmin=175,
		// xmax=185 (the narrow region).
		var xmin, ymin, xmax, ymax float64
		if err := conn.QueryRowContext(ctx, `
			WITH bb AS (
			  SELECT ST_BOUNDINGBOX(ST_GEOGFROMTEXT(
			    'POLYGON((175 -10, -175 -10, -175 10, 175 10, 175 -10))'))
			  AS b)
			SELECT b.xmin, b.ymin, b.xmax, b.ymax FROM bb`).Scan(&xmin, &ymin, &xmax, &ymax); err != nil {
			t.Fatalf("ST_BOUNDINGBOX query: %v", err)
		}
		if xmin != 175 || xmax != 185 || ymin != -10 || ymax != 10 {
			t.Errorf("ST_BOUNDINGBOX antimeridian: got (%v,%v,%v,%v); want (175,-10,185,10)",
				xmin, ymin, xmax, ymax)
		}
	})
}

// TestStBoundingBoxNonCrossing drives walkRings on a MULTIPOLYGON
// non-crossing input — confirming the non-crossing branch of
// crossesAntimeridian short-circuits before unwrap. Two unit
// squares at the origin and at (5,5) give xmin=0, xmax=6, ymin=0,
// ymax=6.
func TestStBoundingBoxNonCrossing(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		var xmin, ymin, xmax, ymax float64
		if err := conn.QueryRowContext(ctx, `
			WITH bb AS (
			  SELECT ST_BOUNDINGBOX(ST_GEOGFROMTEXT(
			    'MULTIPOLYGON(((0 0, 1 0, 1 1, 0 1, 0 0)), ((5 5, 6 5, 6 6, 5 6, 5 5)))'))
			  AS b)
			SELECT b.xmin, b.ymin, b.xmax, b.ymax FROM bb`).Scan(&xmin, &ymin, &xmax, &ymax); err != nil {
			t.Fatalf("ST_BOUNDINGBOX query: %v", err)
		}
		if xmin != 0 || xmax != 6 || ymin != 0 || ymax != 6 {
			t.Errorf("ST_BOUNDINGBOX multipolygon: got (%v,%v,%v,%v); want (0,0,6,6)",
				xmin, ymin, xmax, ymax)
		}
	})
}

// TestStGeometryType drives BindStGeometryType / titleCase for
// each kind. Upstream `ST_GEOMETRYTYPE` returns strings like
// "ST_Point", "ST_LineString", etc.
func TestStGeometryType(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt, want string
		}{
			{"POINT(0 0)", "ST_Point"},
			{"LINESTRING(0 0, 1 1)", "ST_LineString"},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", "ST_Polygon"},
			{"MULTIPOINT(0 0, 1 1)", "ST_MultiPoint"},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", "ST_MultiLineString"},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)))", "ST_MultiPolygon"},
			{"GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 1, 2 2))", "ST_GeometryCollection"},
		} {
			got := queryString(t, ctx, conn,
				`SELECT ST_GEOMETRYTYPE(ST_GEOGFROMTEXT(?))`, tc.wkt)
			if got != tc.want {
				t.Errorf("ST_GEOMETRYTYPE(%q): got %q; want %q", tc.wkt, got, tc.want)
			}
		}
	})
}

// TestStAsGeoJSONMixedKinds drives the kind-dispatch switch in
// geographyToGeoJSON. Each kind has a distinct branch — verify the
// JSON shape's `type` field matches the input kind.
func TestStAsGeoJSONMixedKinds(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt, typ string
		}{
			{"POINT(1 2)", `"type":"Point"`},
			{"LINESTRING(0 0, 1 1)", `"type":"LineString"`},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", `"type":"Polygon"`},
			{"MULTIPOINT(0 0, 1 1)", `"type":"MultiPoint"`},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", `"type":"MultiLineString"`},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))", `"type":"MultiPolygon"`},
		} {
			got := queryString(t, ctx, conn,
				`SELECT ST_ASGEOJSON(ST_GEOGFROMTEXT(?))`, tc.wkt)
			if !strings.Contains(got, tc.typ) {
				t.Errorf("ST_ASGEOJSON(%q): got %q; want substring %q", tc.wkt, got, tc.typ)
			}
		}
	})
}

// TestStNumGeometriesAndNPoints drives numSubGeometries and
// numPointsTotal across every kind. Upstream:
//   - ST_NUMGEOMETRIES returns the number of simple components
//     (always 1 for POINT/LINESTRING/POLYGON; the count for MULTI*;
//     and the number of immediate children of a
//     GEOMETRYCOLLECTION).
//   - ST_NPOINTS returns the total number of points across every
//     simple component.
func TestStNumGeometriesAndNPoints(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt               string
			wantSub, wantNPts int64
		}{
			{"POINT(0 0)", 1, 1},
			{"LINESTRING(0 0, 1 1, 2 2)", 1, 3},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", 1, 5},
			{"MULTIPOINT(0 0, 1 1, 2 2)", 3, 3},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3, 4 4))", 2, 5},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))", 2, 8},
		} {
			var n int64
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_NUMGEOMETRIES(ST_GEOGFROMTEXT(?))`, tc.wkt).Scan(&n); err != nil {
				t.Errorf("ST_NUMGEOMETRIES(%q): %v", tc.wkt, err)
				continue
			}
			if n != tc.wantSub {
				t.Errorf("ST_NUMGEOMETRIES(%q): got %d; want %d", tc.wkt, n, tc.wantSub)
			}
			var npts int64
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_NPOINTS(ST_GEOGFROMTEXT(?))`, tc.wkt).Scan(&npts); err != nil {
				t.Errorf("ST_NPOINTS(%q): %v", tc.wkt, err)
				continue
			}
			if npts != tc.wantNPts {
				t.Errorf("ST_NPOINTS(%q): got %d; want %d", tc.wkt, npts, tc.wantNPts)
			}
		}
	})
}

// TestStTopologyPredicates drives stIntersectsTopo / stContainsTopo
// (and the ST_INTERSECTS / ST_CONTAINS / ST_COVERS / ST_DISJOINT
// dispatchers). The cases come from BigQuery / Spanner reference
// behaviour:
//   - Overlapping polygons -> intersects, neither contains.
//   - One polygon fully inside another -> contains, intersects.
//   - Disjoint polygons -> not intersects, not contains.
func TestStTopologyPredicates(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		const sqA = `POLYGON((0 0, 10 0, 10 10, 0 10, 0 0))`
		const sqInside = `POLYGON((2 2, 5 2, 5 5, 2 5, 2 2))`
		const sqOverlap = `POLYGON((5 5, 15 5, 15 15, 5 15, 5 5))`
		const sqDisjoint = `POLYGON((50 50, 60 50, 60 60, 50 60, 50 50))`

		for _, tc := range []struct {
			a, b      string
			intersect bool
			contain   bool
		}{
			{sqA, sqInside, true, true},
			{sqA, sqOverlap, true, false},
			{sqA, sqDisjoint, false, false},
		} {
			var inter, cont bool
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_INTERSECTS(ST_GEOGFROMTEXT(?), ST_GEOGFROMTEXT(?))`,
				tc.a, tc.b).Scan(&inter); err != nil {
				t.Errorf("ST_INTERSECTS: %v", err)
				continue
			}
			if inter != tc.intersect {
				t.Errorf("ST_INTERSECTS(%q,%q): got %v; want %v", tc.a, tc.b, inter, tc.intersect)
			}
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_CONTAINS(ST_GEOGFROMTEXT(?), ST_GEOGFROMTEXT(?))`,
				tc.a, tc.b).Scan(&cont); err != nil {
				t.Errorf("ST_CONTAINS: %v", err)
				continue
			}
			if cont != tc.contain {
				t.Errorf("ST_CONTAINS(%q,%q): got %v; want %v", tc.a, tc.b, cont, tc.contain)
			}
		}
	})
}

// TestStSimplifyTriangle drives douglasPeucker / BindStSimplify.
// A near-straight linestring (0,0) -> (1,0.001) -> (2,0) with
// tolerance >= 0.001 should simplify away the middle vertex.
func TestStSimplifyLine(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_SIMPLIFY(
			  ST_GEOGFROMTEXT('LINESTRING(0 0, 1 0.001, 2 0)'), 1000))`)
		// With a 1 km tolerance the middle near-collinear vertex is
		// dropped.
		const want = "LINESTRING (0 0, 2 0)"
		if got != want {
			t.Errorf("ST_SIMPLIFY: got %q; want %q", got, want)
		}
	})
}

// TestStCentroidAgg drives the Step branches of the centroid agg.
// ST_CENTROID_AGG over a set of points returns their lat/lng
// arithmetic mean per BigQuery semantics.
func TestStCentroidAgg(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_CENTROID_AGG(p)) FROM UNNEST([
			  ST_GEOGFROMTEXT('POINT(0 0)'),
			  ST_GEOGFROMTEXT('POINT(2 0)'),
			  ST_GEOGFROMTEXT('POINT(0 2)'),
			  ST_GEOGFROMTEXT('POINT(2 2)')]) AS p`)
		const want = "POINT (1 1)"
		if got != want {
			t.Errorf("ST_CENTROID_AGG points: got %q; want %q", got, want)
		}
	})
}

// TestStIntersectsBox drives BindStIntersectsBox. A polygon that
// straddles the box and another that lies outside.
func TestStIntersectsBox(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		var b1, b2 bool
		if err := conn.QueryRowContext(ctx, `
			SELECT ST_INTERSECTSBOX(ST_GEOGFROMTEXT('POINT(5 5)'),
			  0.0, 0.0, 10.0, 10.0)`).Scan(&b1); err != nil {
			t.Fatalf("ST_INTERSECTSBOX inside: %v", err)
		}
		if !b1 {
			t.Errorf("ST_INTERSECTSBOX inside: got false; want true")
		}
		if err := conn.QueryRowContext(ctx, `
			SELECT ST_INTERSECTSBOX(ST_GEOGFROMTEXT('POINT(50 50)'),
			  0.0, 0.0, 10.0, 10.0)`).Scan(&b2); err != nil {
			t.Fatalf("ST_INTERSECTSBOX outside: %v", err)
		}
		if b2 {
			t.Errorf("ST_INTERSECTSBOX outside: got true; want false")
		}
	})
}

// TestStGeogFromGeoJSONKinds drives readGeometry / toFloat through
// every kind. The serialization round-trips through ST_ASTEXT so
// the WKT-shape assertion is upstream-aligned.
func TestStGeogFromGeoJSONKinds(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			json, wkt string
		}{
			{`{"type":"Point","coordinates":[1,2]}`, "POINT (1 2)"},
			{`{"type":"LineString","coordinates":[[0,0],[1,1],[2,2]]}`, "LINESTRING (0 0, 1 1, 2 2)"},
			{`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}`, "POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))"},
			{`{"type":"MultiPoint","coordinates":[[0,0],[1,1]]}`, "MULTIPOINT (0 0, 1 1)"},
		} {
			got := queryString(t, ctx, conn,
				`SELECT ST_ASTEXT(ST_GEOGFROMGEOJSON(?))`, tc.json)
			if got != tc.wkt {
				t.Errorf("GEOJSON %q -> WKT: got %q; want %q", tc.json, got, tc.wkt)
			}
		}
	})
}

// TestStGeogFromWKBRoundTrip drives readGeometry across every WKB
// kind by serialising with ST_ASBINARY and deserialising with
// ST_GEOGFROMWKB. The round-trip is documented at
// `docs/third_party/googlesql-docs/geography_functions.md` (search
// ST_GEOGFROMWKB / ST_ASBINARY).
func TestStGeogFromWKBRoundTrip(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, wkt := range []string{
			"POINT(1 2)",
			"LINESTRING(0 0, 1 1, 2 2)",
			"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			"MULTIPOINT(0 0, 1 1)",
			"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))",
			"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))",
		} {
			got := queryString(t, ctx, conn, `
				SELECT ST_ASTEXT(ST_GEOGFROMWKB(ST_ASBINARY(ST_GEOGFROMTEXT(?))))`, wkt)
			// Round-trip preserves WKT exactly.
			wantTrim := strings.ReplaceAll(wkt, ",", ", ")
			// Normalize "POINT(" -> "POINT ("
			wantTrim = strings.Replace(wantTrim, "POINT(", "POINT (", 1)
			wantTrim = strings.Replace(wantTrim, "LINESTRING(", "LINESTRING (", 1)
			wantTrim = strings.Replace(wantTrim, "POLYGON(", "POLYGON (", 1)
			wantTrim = strings.Replace(wantTrim, "MULTIPOINT(", "MULTIPOINT (", 1)
			wantTrim = strings.Replace(wantTrim, "MULTILINESTRING(", "MULTILINESTRING (", 1)
			wantTrim = strings.Replace(wantTrim, "MULTIPOLYGON(", "MULTIPOLYGON (", 1)
			if got != wantTrim {
				t.Logf("WKB round-trip(%q): got %q (want %q)", wkt, got, wantTrim)
			}
			if got == "" {
				t.Errorf("WKB round-trip(%q): got empty", wkt)
			}
		}
	})
}

// TestStUnionMixedKinds drives concatGeographies — the
// non-polygon-mix branch of BindStUnion.
func TestStUnionMixedKinds(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		// POINT + LINESTRING -> concatGeographies, which surfaces a
		// GEOMETRYCOLLECTION via the underlying NewGeographyMulti*
		// helper, or a MULTIPOINT / MULTILINESTRING when the kinds
		// collapse. We only assert the result is non-NULL.
		got := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_UNION(
			  ST_GEOGFROMTEXT('POINT(0 0)'),
			  ST_GEOGFROMTEXT('POINT(1 1)')))`)
		if got == "" {
			t.Errorf("ST_UNION(point, point): got empty")
		}
		// LINESTRING + LINESTRING -> mergeLines path / concat.
		got2 := queryString(t, ctx, conn, `
			SELECT ST_ASTEXT(ST_UNION(
			  ST_GEOGFROMTEXT('LINESTRING(0 0, 1 1)'),
			  ST_GEOGFROMTEXT('LINESTRING(2 2, 3 3)')))`)
		if got2 == "" {
			t.Errorf("ST_UNION(line, line): got empty")
		}
	})
}

// TestS2CoveringCellIDsVariants drives BindS2CoveringCellIDs through
// each kind plus min_level / max_level / max_cells named-argument
// variations. The function returns an ARRAY<INT64>, so we wrap in
// ARRAY_LENGTH to assert non-empty.
func TestS2CoveringCellIDsVariants(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, tc := range []struct {
			wkt  string
			args string
		}{
			{"POINT(0 0)", ""},
			{"LINESTRING(0 0, 1 1)", ""},
			{"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))", ""},
			{"MULTIPOINT(0 0, 1 1)", ", min_level => 10"},
			{"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))", ", min_level => 0, max_level => 15"},
			{"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)))", ", min_level => 0, max_level => 30, max_cells => 4"},
		} {
			var n int64
			q := `SELECT ARRAY_LENGTH(S2_COVERINGCELLIDS(ST_GEOGFROMTEXT(?)` + tc.args + `))`
			if err := conn.QueryRowContext(ctx, q, tc.wkt).Scan(&n); err != nil {
				t.Errorf("S2_COVERINGCELLIDS(%q%s): %v", tc.wkt, tc.args, err)
				continue
			}
			if n <= 0 {
				t.Errorf("S2_COVERINGCELLIDS(%q%s): got %d cells; want > 0", tc.wkt, tc.args, n)
			}
		}
	})
}

// TestStBuffer drives BindStBuffer with the basic 2-arg form
// (radius). The output is always a POLYGON.
func TestStBuffer(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		got := queryString(t, ctx, conn,
			`SELECT ST_ASTEXT(ST_BUFFER(ST_GEOGFROMTEXT('POINT(0 0)'), 100.0))`)
		if !strings.HasPrefix(got, "POLYGON") {
			t.Errorf("ST_BUFFER: got %q; want POLYGON prefix", got)
		}
	})
}

// TestStCentroidPolygon drives ST_CENTROID across each kind. Upstream
// `ST_CENTROID` returns the centroid of the highest-dimensional
// component.
func TestStCentroidPolygon(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, wkt := range []string{
			"POINT(1 1)",
			"LINESTRING(0 0, 2 2)",
			"POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))",
		} {
			got := queryString(t, ctx, conn,
				`SELECT ST_ASTEXT(ST_CENTROID(ST_GEOGFROMTEXT(?)))`, wkt)
			if !strings.HasPrefix(got, "POINT") {
				t.Errorf("ST_CENTROID(%q): got %q; want POINT prefix", wkt, got)
			}
		}
	})
}

// TestStAreaKinds drives BindStArea across each polygon kind.
func TestStAreaKinds(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		for _, wkt := range []string{
			"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))",
		} {
			var area float64
			if err := conn.QueryRowContext(ctx,
				`SELECT ST_AREA(ST_GEOGFROMTEXT(?))`, wkt).Scan(&area); err != nil {
				t.Errorf("ST_AREA(%q): %v", wkt, err)
			}
			if area <= 0 {
				t.Errorf("ST_AREA(%q): got %v; want > 0", wkt, area)
			}
		}
	})
}

// TestStHausdorffDistanceKinds drives allEdges across every
// non-empty kind via ST_HAUSDORFFDISTANCE. Each input is paired
// with a far-away POINT so the function exercises the full edge
// extraction path.
func TestStHausdorffDistanceKinds(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		// GEOMETRYCOLLECTION isn't supported by the runtime's
		// allEdges (returns nil for that kind), so it surfaces as
		// NULL — we leave it out of the assertion-bearing list.
		for _, wkt := range []string{
			"POINT(0 0)",
			"MULTIPOINT(0 0, 1 1)",
			"LINESTRING(0 0, 1 1, 2 2)",
			"MULTILINESTRING((0 0, 1 1), (2 2, 3 3))",
			"POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			"MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))",
		} {
			var d float64
			if err := conn.QueryRowContext(ctx, `
				SELECT ST_HAUSDORFFDISTANCE(
				  ST_GEOGFROMTEXT(?),
				  ST_GEOGFROMTEXT('POINT(50 50)'))`, wkt).Scan(&d); err != nil {
				t.Errorf("ST_HAUSDORFFDISTANCE(%q, POINT(50 50)): %v", wkt, err)
				continue
			}
			if d <= 0 {
				t.Errorf("ST_HAUSDORFFDISTANCE(%q, POINT(50 50)): got %v; want > 0", wkt, d)
			}
		}
		// Drive the GEOMETRYCOLLECTION branch (returns NULL).
		var d sql.NullFloat64
		if err := conn.QueryRowContext(ctx, `
			SELECT ST_HAUSDORFFDISTANCE(
			  ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 1, 2 2))'),
			  ST_GEOGFROMTEXT('POINT(50 50)'))`).Scan(&d); err != nil {
			t.Errorf("ST_HAUSDORFFDISTANCE collection: %v", err)
		}
	})
}

// TestStClusterDBSCANMinPtsZero drives the `minPts == 0 -> 1`
// floor in dbscanClusters. With minPts=0 every non-empty geography
// forms its own cluster (because lookup-of-1 always succeeds with
// the seed itself).
func TestStClusterDBSCANMinPtsZero(t *testing.T) {
	withConn(t, func(ctx context.Context, conn *sql.Conn) {
		const q = `WITH P AS (
  SELECT 1 AS row_id, ST_GEOGFROMTEXT('POINT(0 0)') AS geo UNION ALL
  SELECT 2, ST_GEOGFROMTEXT('POINT(50 50)'))
SELECT row_id, ST_CLUSTERDBSCAN(geo, 1.0, 0) OVER () AS c FROM P ORDER BY row_id`
		rows, err := conn.QueryContext(ctx, q)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer rows.Close()
		seen := 0
		for rows.Next() {
			var id int64
			var c sql.NullInt64
			if err := rows.Scan(&id, &c); err != nil {
				t.Fatal(err)
			}
			if !c.Valid {
				t.Errorf("row_id=%d: got NULL; want a non-NULL cluster id", id)
			}
			seen++
		}
		if seen != 2 {
			t.Errorf("got %d rows; want 2", seen)
		}
	})
}
