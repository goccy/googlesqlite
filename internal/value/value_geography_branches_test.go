package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestGeographyShapeMismatchEqual covers the "other is a different
// concrete tag" early-return in every geographyType.equal method.
// Each branch returns false when the right-hand side's underlying
// type doesn't match the receiver's tag — the existing tests touch
// only one such pairing (linestring vs polygon), so the per-tag
// branches were uncovered.
func TestGeographyShapeMismatchEqual(t *testing.T) {
	t.Parallel()

	point := value.NewGeographyPoint(1, 2)
	line := value.NewGeographyLineString([][2]float64{{0, 0}, {1, 1}})
	poly := value.NewGeographyPolygon([][][2]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}})
	mpoint := value.NewGeographyMultiPoint([][2]float64{{1, 2}})
	mline := value.NewGeographyMultiLineString([][][2]float64{{{1, 2}, {3, 4}}})
	mpoly := value.NewGeographyMultiPolygon([][][][2]float64{{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}}})
	coll := value.NewGeographyCollection([]*value.GeographyValue{point})
	fg, _ := value.GeographyFromWKT("fullglobe")

	// "point.equal(<not-point>)" branch.
	if ok, _ := point.EQ(line); ok {
		t.Fatal("point vs line should not be equal")
	}
	// "line.equal(<not-line>)" branch.
	if ok, _ := line.EQ(poly); ok {
		t.Fatal("line vs poly should not be equal")
	}
	// "poly.equal(<not-poly>)" branch.
	if ok, _ := poly.EQ(mpoint); ok {
		t.Fatal("poly vs mpoint should not be equal")
	}
	// "mpoint.equal(<not-mpoint>)" branch.
	if ok, _ := mpoint.EQ(mline); ok {
		t.Fatal("mpoint vs mline should not be equal")
	}
	// "mline.equal(<not-mline>)" branch.
	if ok, _ := mline.EQ(mpoly); ok {
		t.Fatal("mline vs mpoly should not be equal")
	}
	// "mpoly.equal(<not-mpoly>)" branch.
	if ok, _ := mpoly.EQ(coll); ok {
		t.Fatal("mpoly vs coll should not be equal")
	}
	// "collection.equal(<not-collection>)" branch.
	if ok, _ := coll.EQ(point); ok {
		t.Fatal("coll vs point should not be equal")
	}
	// "fullglobe.equal(<not-fullglobe>)" branch.
	if ok, _ := fg.EQ(point); ok {
		t.Fatal("fullglobe vs point should not be equal")
	}
}

// TestGeographyEqualLengthMismatch covers the length-mismatch
// early-return for every collection-shaped geographyType.equal.
func TestGeographyEqualLengthMismatch(t *testing.T) {
	t.Parallel()

	// Polygon: same kind, different ring count.
	a := value.NewGeographyPolygon([][][2]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}})
	b := value.NewGeographyPolygon([][][2]float64{
		{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}},
		{{0.2, 0.2}, {0.3, 0.2}, {0.3, 0.3}, {0.2, 0.3}, {0.2, 0.2}},
	})
	if ok, _ := a.EQ(b); ok {
		t.Fatal("polygons with different ring counts should not be equal")
	}

	// MultiPoint: same kind, different point count.
	mpa := value.NewGeographyMultiPoint([][2]float64{{1, 2}})
	mpb := value.NewGeographyMultiPoint([][2]float64{{1, 2}, {3, 4}})
	if ok, _ := mpa.EQ(mpb); ok {
		t.Fatal("multipoints with different counts should not be equal")
	}

	// MultiLineString: same kind, different line count.
	mla := value.NewGeographyMultiLineString([][][2]float64{{{1, 2}, {3, 4}}})
	mlb := value.NewGeographyMultiLineString([][][2]float64{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}})
	if ok, _ := mla.EQ(mlb); ok {
		t.Fatal("multilinestrings with different counts should not be equal")
	}

	// MultiPolygon: same kind, different polygon count.
	mpga := value.NewGeographyMultiPolygon([][][][2]float64{{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}}})
	mpgb := value.NewGeographyMultiPolygon([][][][2]float64{
		{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}},
		{{{2, 2}, {3, 2}, {3, 3}, {2, 3}, {2, 2}}},
	})
	if ok, _ := mpga.EQ(mpgb); ok {
		t.Fatal("multipolygons with different counts should not be equal")
	}

	// MultiPolygon: same polygon count but inner-ring count differs
	// — exercises the inner `len(g.polys[i]) != len(other.polys[i])`
	// branch.
	mpgc := value.NewGeographyMultiPolygon([][][][2]float64{
		{
			{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}},
			{{0.2, 0.2}, {0.3, 0.2}, {0.3, 0.3}, {0.2, 0.3}, {0.2, 0.2}},
		},
	})
	if ok, _ := mpga.EQ(mpgc); ok {
		t.Fatal("multipolygons with different inner-ring counts should not be equal")
	}

	// Collection: same kind, different part count.
	col1 := value.NewGeographyCollection([]*value.GeographyValue{value.NewGeographyPoint(1, 2)})
	col2 := value.NewGeographyCollection([]*value.GeographyValue{
		value.NewGeographyPoint(1, 2),
		value.NewGeographyPoint(3, 4),
	})
	if ok, _ := col1.EQ(col2); ok {
		t.Fatal("collections with different part counts should not be equal")
	}

	// Collection: same kind, same count, parts differ — exercises the
	// per-part equal short-circuit.
	col3 := value.NewGeographyCollection([]*value.GeographyValue{value.NewGeographyPoint(9, 9)})
	if ok, _ := col1.EQ(col3); ok {
		t.Fatal("collections with differing parts should not be equal")
	}
}

// TestGeographyValueToWKTNilReceiver hits the early-return on
// GeographyValue.ToWKT when the receiver (or its inner g) is nil.
func TestGeographyValueToWKTNilReceiver(t *testing.T) {
	t.Parallel()

	var nilG *value.GeographyValue
	wkt, err := nilG.ToWKT()
	if err != nil {
		t.Fatal(err)
	}
	if wkt != "" {
		t.Fatalf("expected empty wkt, got %q", wkt)
	}
}

// TestCollectionIsEmptyBranches exercises the IsEmpty loop in the
// geography collection arm: it returns true iff every part is empty,
// false otherwise.
func TestCollectionIsEmptyBranches(t *testing.T) {
	t.Parallel()

	// Collection with one empty part -> IsEmpty true.
	emptyPoint, _ := value.GeographyFromWKT("POINT EMPTY")
	col := value.NewGeographyCollection([]*value.GeographyValue{emptyPoint})
	if !col.IsEmpty() {
		t.Fatal("collection of empty parts should be empty")
	}
	// Collection with one non-empty part -> IsEmpty false.
	notEmpty := value.NewGeographyPoint(1, 2)
	col2 := value.NewGeographyCollection([]*value.GeographyValue{notEmpty})
	if col2.IsEmpty() {
		t.Fatal("collection of non-empty parts should not be empty")
	}
	// Mixed: one empty, one non-empty -> IsEmpty false.
	col3 := value.NewGeographyCollection([]*value.GeographyValue{emptyPoint, notEmpty})
	if col3.IsEmpty() {
		t.Fatal("collection containing a non-empty part should not be empty")
	}
}

// TestGeographyPointEmptyToWKT covers the POINT EMPTY branch on
// geographyPoint.ToWKT.
func TestGeographyPointEmptyToWKT(t *testing.T) {
	t.Parallel()

	g, err := value.GeographyFromWKT("POINT EMPTY")
	if err != nil {
		t.Fatal(err)
	}
	wkt, err := g.ToWKT()
	if err != nil {
		t.Fatal(err)
	}
	if wkt != "POINT EMPTY" {
		t.Fatalf("expected POINT EMPTY, got %q", wkt)
	}
}
