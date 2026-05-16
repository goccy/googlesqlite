// Direct value-level tests for geography Bind* helpers that don't
// have a SQL surface, plus extra branches of helpers that the SQL-
// level tests can't easily reach. These tests live in the `geography`
// (not `geography_test`) package so they can poke unexported state.
//
// Test expectations match the BigQuery / Spanner geography reference
// (see docs/third_party/googlesql-docs/geography_functions.md).

package geography

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestBindStGeometryNDirect exercises BindStGeometryN. It isn't
// reachable from the public SQL dialect yet, but the binder is
// part of the geography surface. Each kind has its own branch in
// the switch.
func TestBindStGeometryNDirect(t *testing.T) {
	t.Parallel()
	pt, _ := value.GeographyFromWKT("POINT(1 2)")
	mp, _ := value.GeographyFromWKT("MULTIPOINT(0 0, 1 1)")
	ml, _ := value.GeographyFromWKT("MULTILINESTRING((0 0, 1 1), (2 2, 3 3))")
	mpoly, _ := value.GeographyFromWKT("MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)), ((5 5, 6 5, 6 6, 5 5)))")

	for _, tc := range []struct {
		desc    string
		g       value.Value
		n       int64
		nilWant bool
	}{
		{"POINT n=1", pt, 1, false},
		{"POINT n=2", pt, 2, true}, // out-of-range
		{"MULTIPOINT n=2", mp, 2, false},
		{"MULTIPOINT n=99", mp, 99, true},
		{"MULTILINESTRING n=1", ml, 1, false},
		{"MULTIPOLYGON n=2", mpoly, 2, false},
		{"n=0", pt, 0, true},
	} {
		got, err := BindStGeometryN(tc.g, value.IntValue(tc.n))
		if err != nil {
			t.Errorf("%s: %v", tc.desc, err)
			continue
		}
		if tc.nilWant && got != nil {
			t.Errorf("%s: got %v; want NIL", tc.desc, got)
		}
		if !tc.nilWant && got == nil {
			t.Errorf("%s: got NIL; want non-nil", tc.desc)
		}
	}
}

// TestBindStInteriorRingsEmpty drives the early-return branch of
// BindStInteriorRings (no inner rings -> empty array).
func TestBindStInteriorRingsEmpty(t *testing.T) {
	t.Parallel()
	g, _ := value.GeographyFromWKT("POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))")
	got, err := BindStInteriorRings(g)
	if err != nil {
		t.Fatalf("ST_INTERIORRINGS no holes: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 0 {
		t.Errorf("ST_INTERIORRINGS no holes: got %d; want 0", len(arr.Values))
	}
}

// TestStDumpDimensionMinusOne drives the catch-all branch of
// stDumpDimension (e.g. MULTIPOINT returns -1 because the explode
// would have already burst it into POINT components).
func TestStDumpDimensionMinusOne(t *testing.T) {
	t.Parallel()
	g, _ := value.GeographyFromWKT("MULTIPOINT(0 0, 1 1)")
	if stDumpDimension(g) != -1 {
		t.Errorf("stDumpDimension(MULTIPOINT) = %d; want -1", stDumpDimension(g))
	}
}

// TestStDumpPointsExplodeEmpty drives stDumpPointsExplode on an
// empty geometry collection (catch-all branch returning empty).
func TestStDumpPointsExplodeEmpty(t *testing.T) {
	t.Parallel()
	g, _ := value.GeographyFromWKT("GEOMETRYCOLLECTION EMPTY")
	pts := stDumpPointsExplode(g)
	if len(pts) != 0 {
		t.Errorf("stDumpPointsExplode(EMPTY) = %d points; want 0", len(pts))
	}
}

// TestApplyOrientationNonPolygon drives the non-POLYGON early-return
// of applyOrientation.
func TestApplyOrientationNonPolygon(t *testing.T) {
	t.Parallel()
	g, _ := value.GeographyFromWKT("LINESTRING(0 0, 1 1)")
	applyOrientation(g) // must not panic
}

// TestUnwrapLongitudesEmpty drives the len==0 early-return.
func TestUnwrapLongitudesEmpty(t *testing.T) {
	t.Parallel()
	out := unwrapLongitudes(nil)
	if len(out) != 0 {
		t.Errorf("unwrapLongitudes(nil): got %d points; want 0", len(out))
	}
}

// TestSignedPlanarAreaXYShortRing drives the n<3 early-return.
func TestSignedPlanarAreaXYShortRing(t *testing.T) {
	t.Parallel()
	if a := signedPlanarAreaXY([][2]float64{{0, 0}, {1, 1}}); a != 0 {
		t.Errorf("signedPlanarAreaXY short ring: got %v; want 0", a)
	}
}

// TestCanonicaliseRingShort drives the len < 4 early-return.
func TestCanonicaliseRingShort(t *testing.T) {
	t.Parallel()
	in := [][2]float64{{0, 0}, {1, 1}}
	out := canonicaliseRing(in)
	if len(out) != 2 {
		t.Errorf("canonicaliseRing short ring: got %d points; want 2", len(out))
	}
}

// TestExtractCoordsBetweenMissing drives the missing-open / missing-
// close branches of extractCoordsBetween.
func TestExtractCoordsBetweenMissing(t *testing.T) {
	t.Parallel()
	if out := extractCoordsBetween("no markers here", "<x>", "</x>"); out != nil {
		t.Errorf("extractCoordsBetween no marker: got %v; want nil", out)
	}
	if out := extractCoordsBetween("<x>oops", "<x>", "</x>"); out != nil {
		t.Errorf("extractCoordsBetween no close: got %v; want nil", out)
	}
}

// TestExtractAllCoordsBetweenMissing drives the early-return branch.
func TestExtractAllCoordsBetweenMissing(t *testing.T) {
	t.Parallel()
	out := extractAllCoordsBetween("no marker", "<x>", "</x>")
	if len(out) != 0 {
		t.Errorf("extractAllCoordsBetween no marker: got %v; want empty", out)
	}
}
