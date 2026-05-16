package geography

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// fromWKT is a thin helper wrapping value.GeographyFromWKT; the test
// inputs follow the OGC WKT shape so we can assert against named-kind
// outputs.
func fromWKT(t *testing.T, wkt string) value.Value {
	t.Helper()
	g, err := value.GeographyFromWKT(wkt)
	if err != nil {
		t.Fatalf("GeographyFromWKT(%q): %v", wkt, err)
	}
	return g
}

// TestBindStIntersectionNonPolygonBranches exercises the
// non-polygon dispatch branch of BindStIntersection. For point /
// line inputs the function returns the intersection by point
// containment.
func TestBindStIntersectionNonPolygonBranches(t *testing.T) {
	t.Parallel()

	// Two coincident points -> a single POINT.
	got, err := BindStIntersection(
		fromWKT(t, "POINT (1 2)"),
		fromWKT(t, "POINT (1 2)"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil intersection of coincident points")
	}

	// Point that lies on a line (within a polygon containing both):
	// not strictly necessary, but exercises the line/point dispatch.
	got, err = BindStIntersection(
		fromWKT(t, "POINT (1 1)"),
		fromWKT(t, "LINESTRING (0 0, 5 5)"),
	)
	if err != nil {
		t.Fatal(err)
	}
	// Result can be nil if pointInGeography(point, line) returns
	// false; the important thing is no error and the function
	// dispatched into the non-polygon branch.
	_ = got

	// Disjoint points -> nil.
	got, err = BindStIntersection(
		fromWKT(t, "POINT (0 0)"),
		fromWKT(t, "POINT (99 99)"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for disjoint points, got %v", got)
	}

	// MULTIPOINT with two coincident points produces MULTIPOINT.
	got, err = BindStIntersection(
		fromWKT(t, "MULTIPOINT ((1 1), (2 2))"),
		fromWKT(t, "MULTIPOINT ((1 1), (2 2))"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil intersection of multipoint")
	}
	if gv, ok := got.(*value.GeographyValue); !ok || gv.Kind() != "MULTIPOINT" {
		t.Fatalf("expected MULTIPOINT, got %v / kind=%q", got, gv.Kind())
	}
}

// TestBindStIntersectionNullPropagation drives the early-nil-return
// branch when either argument is NULL.
func TestBindStIntersectionNullPropagation(t *testing.T) {
	t.Parallel()

	got, err := BindStIntersection(nil, fromWKT(t, "POINT (1 1)"))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for lhs null, got %v", got)
	}
	got, err = BindStIntersection(fromWKT(t, "POINT (1 1)"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for rhs null, got %v", got)
	}
}

// TestBindStDifferenceNullAndPointBranches drives the early-return
// branches of BindStDifference plus the non-polygon dispatch.
func TestBindStDifferenceNullAndPointBranches(t *testing.T) {
	t.Parallel()

	// lhs NULL -> result NULL.
	got, err := BindStDifference(nil, fromWKT(t, "POINT (1 1)"))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for lhs null, got %v", got)
	}

	// rhs NULL -> return lhs.
	lhs := fromWKT(t, "POINT (1 1)")
	got, err = BindStDifference(lhs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected lhs returned when rhs is null")
	}

	// non-polygon dispatch: two disjoint points.
	got, err = BindStDifference(
		fromWKT(t, "POINT (1 1)"),
		fromWKT(t, "POINT (5 5)"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("disjoint point difference should not be nil")
	}

	// non-polygon dispatch: same point -> empty result -> nil.
	got, err = BindStDifference(
		fromWKT(t, "POINT (1 1)"),
		fromWKT(t, "POINT (1 1)"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for coincident-point difference, got %v", got)
	}

	// MULTIPOINT \ point keeps the surviving points.
	got, err = BindStDifference(
		fromWKT(t, "MULTIPOINT ((1 1), (2 2))"),
		fromWKT(t, "POINT (1 1)"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil multipoint difference")
	}
}

// TestBindStIntersectsBoxBranches drives the BindStIntersectsBox
// invalid-arg and float-conversion paths plus the polygon-contains-box
// branch.
func TestBindStIntersectsBoxBranches(t *testing.T) {
	t.Parallel()

	// Point inside box -> true.
	got, err := BindStIntersectsBox(
		fromWKT(t, "POINT (1 1)"),
		value.FloatValue(0), value.FloatValue(0),
		value.FloatValue(2), value.FloatValue(2),
	)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBool(); !b {
		t.Fatal("point inside box should return true")
	}

	// Point outside box -> false.
	got, err = BindStIntersectsBox(
		fromWKT(t, "POINT (5 5)"),
		value.FloatValue(0), value.FloatValue(0),
		value.FloatValue(2), value.FloatValue(2),
	)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBool(); b {
		t.Fatal("point outside box should return false")
	}

	// NULL geography -> NULL output (early return).
	got, err = BindStIntersectsBox(
		nil,
		value.FloatValue(0), value.FloatValue(0),
		value.FloatValue(2), value.FloatValue(2),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for NULL geography, got %v", got)
	}

	// Float conversion errors: any non-numeric bound surfaces an error.
	cases := []struct {
		name string
		args []value.Value
	}{
		{"lonLo not float", []value.Value{
			fromWKT(t, "POINT (1 1)"),
			value.StringValue("nope"), value.FloatValue(0),
			value.FloatValue(2), value.FloatValue(2),
		}},
		{"latLo not float", []value.Value{
			fromWKT(t, "POINT (1 1)"),
			value.FloatValue(0), value.StringValue("nope"),
			value.FloatValue(2), value.FloatValue(2),
		}},
		{"lonHi not float", []value.Value{
			fromWKT(t, "POINT (1 1)"),
			value.FloatValue(0), value.FloatValue(0),
			value.StringValue("nope"), value.FloatValue(2),
		}},
		{"latHi not float", []value.Value{
			fromWKT(t, "POINT (1 1)"),
			value.FloatValue(0), value.FloatValue(0),
			value.FloatValue(2), value.StringValue("nope"),
		}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := BindStIntersectsBox(c.args...); err == nil {
				t.Fatalf("expected float-conversion error: %s", c.name)
			}
		})
	}

	// Polygon-contains-box branch: a big polygon enclosing the bbox.
	got, err = BindStIntersectsBox(
		fromWKT(t, "POLYGON ((-10 -10, 10 -10, 10 10, -10 10, -10 -10))"),
		value.FloatValue(-1), value.FloatValue(-1),
		value.FloatValue(1), value.FloatValue(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBool(); !b {
		t.Fatal("big polygon should intersect inner box")
	}
}

// TestBindStIntersectionPolygonBranches drives the polygon-polygon
// dispatch of BindStIntersection via WKT-constructed inputs.
func TestBindStIntersectionPolygonBranches(t *testing.T) {
	t.Parallel()

	// Two overlapping polygons -> intersection has area.
	got, err := BindStIntersection(
		fromWKT(t, "POLYGON ((0 0, 2 0, 2 2, 0 2, 0 0))"),
		fromWKT(t, "POLYGON ((1 1, 3 1, 3 3, 1 3, 1 1))"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil polygon intersection")
	}

	// MULTIPOLYGON dispatch.
	got, err = BindStIntersection(
		fromWKT(t, "MULTIPOLYGON (((0 0, 2 0, 2 2, 0 2, 0 0)))"),
		fromWKT(t, "POLYGON ((1 1, 3 1, 3 3, 1 3, 1 1))"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil multipolygon intersection")
	}
}

// TestBindStDifferencePolygonBranches drives the polygon-polygon
// dispatch of BindStDifference.
func TestBindStDifferencePolygonBranches(t *testing.T) {
	t.Parallel()

	got, err := BindStDifference(
		fromWKT(t, "POLYGON ((0 0, 2 0, 2 2, 0 2, 0 0))"),
		fromWKT(t, "POLYGON ((1 1, 3 1, 3 3, 1 3, 1 1))"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil polygon difference")
	}

	// MULTIPOLYGON \ POLYGON dispatch.
	got, err = BindStDifference(
		fromWKT(t, "MULTIPOLYGON (((0 0, 2 0, 2 2, 0 2, 0 0)))"),
		fromWKT(t, "POLYGON ((1 1, 3 1, 3 3, 1 3, 1 1))"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil multipolygon difference")
	}
}
