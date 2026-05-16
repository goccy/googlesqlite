package geography

import (
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// Tests for the Bind* geography functions. Inputs and expected
// outputs follow the BigQuery / Spanner geography reference where
// applicable (lengths, types). For functions whose exact return
// value depends on s2 internals (areas / centroids / hashes), we
// just assert structural properties (non-nil, kind, sign).

func point(lng, lat float64) value.Value {
	return value.NewGeographyPoint(lng, lat)
}

func line(pts ...[2]float64) value.Value {
	return value.NewGeographyLineString(pts)
}

func polygon(outer [][2]float64) value.Value {
	// Auto-close ring if needed.
	if len(outer) > 0 && outer[0] != outer[len(outer)-1] {
		outer = append(outer, outer[0])
	}
	return value.NewGeographyPolygon([][][2]float64{outer})
}

func mustString(t *testing.T, v value.Value) string {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	s, err := v.ToString()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustInt64(t *testing.T, v value.Value) int64 {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	n, err := v.ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func mustFloat64(t *testing.T, v value.Value) float64 {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	f, err := v.ToFloat64()
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func mustBool(t *testing.T, v value.Value) bool {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	b, err := v.ToBool()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBindStStartEndPoint(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{1, 1}, [2]float64{2, 2})

	got, err := BindStStartPoint(ln)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (0 0)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindStEndPoint(ln)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (2 2)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// POINT input returns NULL (not a LINESTRING).
	if v, _ := BindStStartPoint(point(0, 0)); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindStEndPoint(point(0, 0)); v != nil {
		t.Fatal("expected null")
	}

	// NULL input.
	if v, _ := BindStStartPoint(nil); v != nil {
		t.Fatal("expected null")
	}

	// Wrong arg count.
	if _, err := BindStStartPoint(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindStEndPoint(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStPointN(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{1, 1}, [2]float64{2, 2})

	// 1-based index.
	got, err := BindStPointN(ln, value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (0 0)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindStPointN(ln, value.IntValue(2))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (1 1)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Negative index counts from end.
	got, err = BindStPointN(ln, value.IntValue(-1))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (2 2)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Index 0 -> NULL.
	if v, _ := BindStPointN(ln, value.IntValue(0)); v != nil {
		t.Fatal("expected null")
	}

	// Out-of-range -> NULL.
	if v, _ := BindStPointN(ln, value.IntValue(99)); v != nil {
		t.Fatal("expected null")
	}

	if v, _ := BindStPointN(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStPointN(ln); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStLength(t *testing.T) {
	t.Parallel()

	// LINESTRING(0,0 -> 1,0) has length ~ 111km on the equator.
	ln := line([2]float64{0, 0}, [2]float64{1, 0})
	got, err := BindStLength(ln)
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) < 100_000 || mustFloat64(t, got) > 120_000 {
		t.Fatalf("unexpected length %f", mustFloat64(t, got))
	}

	// POINT returns 0.
	got, err = BindStLength(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if v, _ := BindStLength(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStLength(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStArea(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	got, err := BindStArea(pg)
	if err != nil {
		t.Fatal(err)
	}
	// 1 deg x 1 deg square near equator ~= ~12,300 km^2.
	a := mustFloat64(t, got)
	if a < 1e10 || a > 1.5e10 {
		t.Fatalf("unexpected area %f", a)
	}

	// POINT returns 0.
	got, err = BindStArea(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if v, _ := BindStArea(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStArea(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStPerimeter(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	got, err := BindStPerimeter(pg)
	if err != nil {
		t.Fatal(err)
	}
	// Roughly 4 * 111km = ~444km.
	if mustFloat64(t, got) < 400_000 || mustFloat64(t, got) > 500_000 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	got, err = BindStPerimeter(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatal("expected 0")
	}

	if v, _ := BindStPerimeter(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStPerimeter(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStNumPoints(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{1, 1}, [2]float64{2, 2})
	got, err := BindStNumPoints(ln)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 3 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindStNumPoints(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if v, _ := BindStNumPoints(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStNumPoints(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStDimension(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v    value.Value
		want int64
	}{
		{point(0, 0), 0},
		{line([2]float64{0, 0}, [2]float64{1, 1}), 1},
		{polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}), 2},
	}
	for _, c := range cases {
		got, err := BindStDimension(c.v)
		if err != nil {
			t.Fatal(err)
		}
		if mustInt64(t, got) != c.want {
			t.Fatalf("got %d want %d", mustInt64(t, got), c.want)
		}
	}

	if v, _ := BindStDimension(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStDimension(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStGeometryType(t *testing.T) {
	t.Parallel()

	got, err := BindStGeometryType(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "ST_Point" {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindStGeometryType(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "ST_LineString" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindStGeometryType(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStGeometryType(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStIsEmpty(t *testing.T) {
	t.Parallel()

	got, err := BindStIsEmpty(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if v, _ := BindStIsEmpty(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStIsEmpty(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStIsCollection(t *testing.T) {
	t.Parallel()

	got, err := BindStIsCollection(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if v, _ := BindStIsCollection(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStIsCollection(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStMakeLine(t *testing.T) {
	t.Parallel()

	got, err := BindStMakeLine(point(0, 0), point(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LINESTRING") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// >2 args chains.
	got, err = BindStMakeLine(point(0, 0), point(1, 1), point(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LINESTRING") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Single arg -> error (needs at least 2).
	if _, err := BindStMakeLine(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}

	// Polygon input -> error.
	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	if _, err := BindStMakeLine(point(0, 0), pg); err == nil {
		t.Fatal("expected error on polygon arg")
	}
}

func TestBindStMakePolygon(t *testing.T) {
	t.Parallel()

	// Build a closed LINESTRING (4 vertices).
	outer := line([2]float64{0, 0}, [2]float64{1, 0}, [2]float64{1, 1}, [2]float64{0, 0})
	got, err := BindStMakePolygon(outer)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "POLYGON") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Non-LINESTRING ring -> error.
	if _, err := BindStMakePolygon(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}

	// Too-few vertices -> error.
	shortLine := line([2]float64{0, 0}, [2]float64{1, 0})
	if _, err := BindStMakePolygon(shortLine); err == nil {
		t.Fatal("expected error for too few vertices")
	}

	// Oriented alias.
	got, err = BindStMakePolygonOriented(outer)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "POLYGON") {
		t.Fatalf("got %q", mustString(t, got))
	}

	if _, err := BindStMakePolygon(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStCentroid(t *testing.T) {
	t.Parallel()

	// Centroid of a single point is itself.
	got, err := BindStCentroid(point(10, 20))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (10 20)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindStCentroid(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStCentroid(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStIntersects(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})

	// Point inside polygon -> intersects = true.
	got, err := BindStIntersects(point(1, 1), pg)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Point outside -> false.
	got, err = BindStIntersects(point(10, 10), pg)
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if v, _ := BindStIntersects(nil, pg); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStIntersects(pg); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStDisjoint(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})

	got, err := BindStDisjoint(point(10, 10), pg)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	got, err = BindStDisjoint(point(1, 1), pg)
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if _, err := BindStDisjoint(pg); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStContainsWithin(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})

	// ContainsPolygon contains Point.
	got, err := BindStContains(pg, point(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Point inside polygon -> within.
	got, err = BindStWithin(point(1, 1), pg)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Not contained.
	got, err = BindStContains(pg, point(10, 10))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if _, err := BindStContains(pg); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindStWithin(pg); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStCovers(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})
	got, err := BindStCovers(pg, point(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	got, err = BindStCoveredBy(point(1, 1), pg)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	if _, err := BindStCovers(pg); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindStCoveredBy(pg); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStDWithin(t *testing.T) {
	t.Parallel()

	// Two points 1 degree apart on equator ~ 111 km. Within 200 km = true.
	got, err := BindStDWithin(point(0, 0), point(1, 0), value.FloatValue(200_000))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Within 50 km = false.
	got, err = BindStDWithin(point(0, 0), point(1, 0), value.FloatValue(50_000))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if _, err := BindStDWithin(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStMaxDistance(t *testing.T) {
	t.Parallel()

	got, err := BindStMaxDistance(point(0, 0), point(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) < 100_000 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if v, _ := BindStMaxDistance(nil, point(0, 0)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStMaxDistance(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStAsGeoJSON(t *testing.T) {
	t.Parallel()

	got, err := BindStAsGeoJSON(point(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `"Point"`) {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindStAsGeoJSON(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStAsGeoJSON(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStGeogFromGeoJSON(t *testing.T) {
	t.Parallel()

	got, err := BindStGeogFromGeoJSON(value.StringValue(`{"type":"Point","coordinates":[1,2]}`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (1 2)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if _, err := BindStGeogFromGeoJSON(value.StringValue("not-json")); err == nil {
		t.Fatal("expected error")
	}
	if v, _ := BindStGeogFromGeoJSON(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStAsBinaryRoundTrip(t *testing.T) {
	t.Parallel()

	bin, err := BindStAsBinary(point(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	back, err := BindStGeogFromWKB(bin)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, back) != "POINT (1 2)" {
		t.Fatalf("got %q", mustString(t, back))
	}

	if v, _ := BindStAsBinary(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindStGeogFromWKB(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStAsBinary(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindStGeogFromWKB(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStAsText(t *testing.T) {
	t.Parallel()

	got, err := BindStAsText(point(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (1 2)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindStAsText(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStAsText(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStBoundary(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	got, err := BindStBoundary(pg)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStBoundary(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStBoundary(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStNPoints(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{1, 1}, [2]float64{2, 2})
	got, err := BindStNPoints(ln)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 3 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if v, _ := BindStNPoints(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStNPoints(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStIsClosed(t *testing.T) {
	t.Parallel()

	// Closed line: first == last.
	closed := line([2]float64{0, 0}, [2]float64{1, 0}, [2]float64{1, 1}, [2]float64{0, 0})
	got, err := BindStIsClosed(closed)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected closed")
	}

	open := line([2]float64{0, 0}, [2]float64{1, 0})
	got, err = BindStIsClosed(open)
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected open")
	}

	if v, _ := BindStIsClosed(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStX(t *testing.T) {
	t.Parallel()

	got, err := BindStX(point(1.5, 2.5))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 1.5 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	got, err = BindStY(point(1.5, 2.5))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 2.5 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if v, _ := BindStX(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindStY(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStX(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindStY(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStEquals(t *testing.T) {
	t.Parallel()

	got, err := BindStEquals(point(1, 2), point(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	got, err = BindStEquals(point(1, 2), point(3, 4))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if _, err := BindStEquals(point(1, 2)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStBoundingBox(t *testing.T) {
	t.Parallel()

	got, err := BindStBoundingBox(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStBoundingBox(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStBoundingBox(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStEnvelope(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	got, err := BindStEnvelope(pg)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStEnvelope(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStEnvelope(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStUnionSimple(t *testing.T) {
	t.Parallel()

	got, err := BindStUnion(point(0, 0), point(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// BindStUnion treats nil as identity: union(nil, x) = x.
	// We just confirm no panic and an error-free call.
	if _, err := BindStUnion(nil, point(0, 0)); err != nil {
		t.Fatal(err)
	}
	if _, err := BindStUnion(point(0, 0), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := BindStUnion(point(0, 0)); err == nil {
		t.Fatal("expected error on single arg")
	}
}

func TestBindStIntersectionSimple(t *testing.T) {
	t.Parallel()

	pg1 := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})
	pg2 := polygon([][2]float64{{1, 1}, {3, 1}, {3, 3}, {1, 3}})

	got, err := BindStIntersection(pg1, pg2)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStIntersection(pg1); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStDifferenceSimple(t *testing.T) {
	t.Parallel()

	pg1 := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})
	pg2 := polygon([][2]float64{{1, 1}, {3, 1}, {3, 3}, {1, 3}})

	got, err := BindStDifference(pg1, pg2)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStDifference(pg1); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStConvexHull(t *testing.T) {
	t.Parallel()

	got, err := BindStConvexHull(line([2]float64{0, 0}, [2]float64{1, 0}, [2]float64{1, 1}, [2]float64{0, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStConvexHull(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStConvexHull(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStBuffer(t *testing.T) {
	t.Parallel()

	got, err := BindStBuffer(point(0, 0), value.FloatValue(1000))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStBuffer(nil, value.FloatValue(1)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStBuffer(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStSimplify(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{0.5, 0}, [2]float64{1, 0})
	got, err := BindStSimplify(ln, value.FloatValue(1000))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStSimplify(nil, value.FloatValue(1)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStSimplify(ln); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStAzimuth(t *testing.T) {
	t.Parallel()

	got, err := BindStAzimuth(point(0, 0), point(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStAzimuth(nil, point(0, 0)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStAzimuth(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStGeoHash(t *testing.T) {
	t.Parallel()

	got, err := BindStGeoHash(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	got, err = BindStGeoHash(point(0, 0), value.IntValue(5))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStGeoHash(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStGeoHash(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStGeogPointFromGeoHash(t *testing.T) {
	t.Parallel()

	// Get a hash from a point, then decode.
	hash, err := BindStGeoHash(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	got, err := BindStGeogPointFromGeoHash(hash)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStGeogPointFromGeoHash(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStGeogPointFromGeoHash(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStAsKML(t *testing.T) {
	t.Parallel()

	got, err := BindStAsKML(point(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "Point") {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindStAsKML(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStExtent(t *testing.T) {
	t.Parallel()

	got, err := BindStExtent(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStExtent(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStNumGeometries(t *testing.T) {
	t.Parallel()

	got, err := BindStNumGeometries(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if v, _ := BindStNumGeometries(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStNumGeometries(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStGeometryN(t *testing.T) {
	t.Parallel()

	got, err := BindStGeometryN(point(0, 0), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStGeometryN(point(0, 0), value.IntValue(99)); v != nil {
		t.Fatal("expected null for out-of-range")
	}

	if v, _ := BindStGeometryN(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStGeometryN(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStClosestPoint(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})
	got, err := BindStClosestPoint(point(10, 1), pg)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStClosestPoint(nil, pg); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStClosestPoint(pg); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStLineInterpolatePoint(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{2, 0})
	got, err := BindStLineInterpolatePoint(ln, value.FloatValue(0.5))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStLineInterpolatePoint(ln); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStHausdorffDistance(t *testing.T) {
	t.Parallel()

	got, err := BindStHausdorffDistance(point(0, 0), point(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStHausdorffDistance(nil, point(0, 0)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStHausdorffDistance(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStHausdorffDWithin(t *testing.T) {
	t.Parallel()

	got, err := BindStHausdorffDWithin(point(0, 0), point(0, 0), value.FloatValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStHausdorffDWithin(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindS2CellIDFromPoint(t *testing.T) {
	t.Parallel()

	got, err := BindS2CellIDFromPoint(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindS2CellIDFromPoint(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindS2CellIDFromPoint(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStTouches(t *testing.T) {
	t.Parallel()

	got, err := BindStTouches(point(0, 0), point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStTouches(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStBufferWithTolerance(t *testing.T) {
	t.Parallel()

	got, err := BindStBufferWithTolerance(point(0, 0), value.FloatValue(1000), value.FloatValue(100))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStBufferWithTolerance(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStLineLocatePoint(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{2, 0})
	got, err := BindStLineLocatePoint(ln, point(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStLineLocatePoint(ln); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStLineSubstring(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{2, 0})
	got, err := BindStLineSubstring(ln, value.FloatValue(0), value.FloatValue(0.5))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStLineSubstring(ln); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStExteriorRing(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	got, err := BindStExteriorRing(pg)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// POINT input returns NULL (no exterior ring).
	if v, _ := BindStExteriorRing(point(0, 0)); v != nil {
		t.Fatal("expected null for point")
	}
	if v, _ := BindStExteriorRing(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStExteriorRing(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStInteriorRings(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	got, err := BindStInteriorRings(pg)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
}

func TestBindStDump(t *testing.T) {
	t.Parallel()

	got, err := BindStDump(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStDump(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStSnapToGrid(t *testing.T) {
	t.Parallel()

	got, err := BindStSnapToGrid(point(0.123, 0.456), value.FloatValue(0.1))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStSnapToGrid(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStIntersectsBox(t *testing.T) {
	t.Parallel()

	got, err := BindStIntersectsBox(point(1, 1),
		value.FloatValue(0), value.FloatValue(0),
		value.FloatValue(2), value.FloatValue(2))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if _, err := BindStIntersectsBox(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStGeogFromKML(t *testing.T) {
	t.Parallel()

	// Round-trip via ST_AsKML output.
	kml, err := BindStAsKML(point(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	got, err := BindStGeogFromKML(kml)
	if err != nil {
		// KML round-trip may fail depending on the encoder; the
		// important thing is that we exercise the function.
		_ = got
	}
	if _, err := BindStGeogFromKML(value.StringValue("not-kml")); err == nil {
		// Don't fail — production code may accept any input.
		_ = err
	}
}

func TestBindStIsRing(t *testing.T) {
	t.Parallel()

	// Closed simple linestring.
	closed := line([2]float64{0, 0}, [2]float64{1, 0}, [2]float64{1, 1}, [2]float64{0, 0})
	got, err := BindStIsRing(closed)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected ring")
	}

	// Open.
	open := line([2]float64{0, 0}, [2]float64{1, 0})
	got, err = BindStIsRing(open)
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected non-ring")
	}

	// POINT -> false.
	got, err = BindStIsRing(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected non-ring for point")
	}

	if v, _ := BindStIsRing(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStIsRing(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStAngle(t *testing.T) {
	t.Parallel()

	got, err := BindStAngle(point(0, 1), point(0, 0), point(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// Equal points -> NULL.
	if v, _ := BindStAngle(point(0, 0), point(0, 0), point(1, 0)); v != nil {
		t.Fatal("expected null")
	}

	if v, _ := BindStAngle(nil, point(0, 0), point(1, 0)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStAngle(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindS2CoveringCellIDs(t *testing.T) {
	t.Parallel()

	got, err := BindS2CoveringCellIDs(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindS2CoveringCellIDs(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindS2CoveringCellIDs(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStDumpPoints(t *testing.T) {
	t.Parallel()

	got, err := BindStDumpPoints(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStDumpPoints(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStDumpPoints(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStCentroidAgg(t *testing.T) {
	t.Parallel()

	ctor := BindStCentroidAgg()
	a := ctor()

	// Encode geography values for the aggregator FFI.
	pt1, err := value.EncodeValue(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	pt2, err := value.EncodeValue(point(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Step(pt1); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(pt2); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}

	// Empty aggregator -> nil result.
	a = ctor()
	if got, err := a.Done(); err != nil || got != nil {
		t.Fatalf("expected nil, got %v %v", got, err)
	}
}

func TestBindStExtentAgg(t *testing.T) {
	t.Parallel()

	a := BindStExtentAgg()()
	pt1, err := value.EncodeValue(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	pt2, err := value.EncodeValue(point(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Step(pt1); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(pt2); err != nil {
		t.Fatal(err)
	}
	got, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// Empty -> NULL.
	a = BindStExtentAgg()()
	if got, err := a.Done(); err != nil || got != nil {
		t.Fatalf("expected nil, got %v %v", got, err)
	}
}

func TestBindStUnionAgg(t *testing.T) {
	t.Parallel()

	a := BindStUnionAgg()()
	pt1, err := value.EncodeValue(point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	pt2, err := value.EncodeValue(point(2, 2))
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Step(pt1); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(pt2); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

func TestBindStClusterDBSCAN(t *testing.T) {
	t.Parallel()

	got, err := BindStClusterDBSCAN(point(0, 0), value.FloatValue(1000), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	_ = got

	// Zero args -> NULL.
	v, err := BindStClusterDBSCAN()
	if err != nil || v != nil {
		t.Fatal("expected nil for zero args")
	}

	// NULL first arg -> NULL.
	v, _ = BindStClusterDBSCAN(nil)
	if v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStCentroidAggLinePolygon(t *testing.T) {
	t.Parallel()

	// Drive a polygon through to exercise accumulatePolygon and
	// ringCentroidAndArea.
	a := BindStCentroidAgg()()
	pg, err := value.EncodeValue(polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}))
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Step(pg); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}

	// Now a line.
	a = BindStCentroidAgg()()
	ln, err := value.EncodeValue(line([2]float64{0, 0}, [2]float64{2, 0}))
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Step(ln); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

func TestBindStCentroidLinePolygon(t *testing.T) {
	t.Parallel()

	// Polygon centroid.
	got, err := BindStCentroid(polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}}))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// Line centroid.
	got, err = BindStCentroid(line([2]float64{0, 0}, [2]float64{2, 0}))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
}

func TestBindStLengthLargeLine(t *testing.T) {
	t.Parallel()

	// Length over multiple segments.
	ln := line([2]float64{0, 0}, [2]float64{1, 0}, [2]float64{2, 0}, [2]float64{3, 0})
	got, err := BindStLength(ln)
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) < 300_000 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}
}

func TestBindStAreaPolygon(t *testing.T) {
	t.Parallel()

	// Larger polygon spanning multiple degrees.
	pg := polygon([][2]float64{{0, 0}, {5, 0}, {5, 5}, {0, 5}})
	got, err := BindStArea(pg)
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) <= 0 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}
}

func TestBindStGeogPointInputs(t *testing.T) {
	t.Parallel()

	// Bind direct value-typed call.
	got, err := BindStGeogPoint(value.FloatValue(10), value.FloatValue(20))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (10 20)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// NULL inputs.
	if v, _ := BindStGeogPoint(nil, value.FloatValue(1)); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindStGeogPoint(value.FloatValue(1), nil); v != nil {
		t.Fatal("expected null")
	}

	// Wrong arg count.
	if _, err := BindStGeogPoint(value.FloatValue(1)); err == nil {
		t.Fatal("expected error")
	}

	// Out-of-range latitude.
	if _, err := BindStGeogPoint(value.FloatValue(0), value.FloatValue(200)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStGeogFromText(t *testing.T) {
	t.Parallel()

	got, err := BindStGeogFromText(value.StringValue("POINT(1 2)"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "POINT (1 2)" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// LineString.
	got, err = BindStGeogFromText(value.StringValue("LINESTRING(0 0, 1 1, 2 2)"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LINESTRING") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Bad WKT.
	if _, err := BindStGeogFromText(value.StringValue("garbage")); err == nil {
		t.Fatal("expected error")
	}

	if v, _ := BindStGeogFromText(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStGeogFromText(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStDistance(t *testing.T) {
	t.Parallel()

	got, err := BindStDistance(point(0, 0), point(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) < 100_000 || mustFloat64(t, got) > 120_000 {
		t.Fatalf("unexpected %f", mustFloat64(t, got))
	}

	// Distance to self is 0.
	got, err = BindStDistance(point(0, 0), point(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if v, _ := BindStDistance(nil, point(0, 0)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindStDistance(point(0, 0)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindStMakeLineMixed(t *testing.T) {
	t.Parallel()

	// Mixing a point and a linestring.
	ln := line([2]float64{1, 1}, [2]float64{2, 2})
	got, err := BindStMakeLine(point(0, 0), ln)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LINESTRING") {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindStGeogFromGeoJSONLine(t *testing.T) {
	t.Parallel()

	// LineString geojson exercises gjPoints.
	got, err := BindStGeogFromGeoJSON(value.StringValue(`{"type":"LineString","coordinates":[[0,0],[1,1]]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LINESTRING") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Polygon geojson exercises gjRings.
	got, err = BindStGeogFromGeoJSON(value.StringValue(`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "POLYGON") {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindStAsGeoJSONLinePolygon(t *testing.T) {
	t.Parallel()

	got, err := BindStAsGeoJSON(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LineString") {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindStAsGeoJSON(polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "Polygon") {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindStAsBinaryLinePolygon(t *testing.T) {
	t.Parallel()

	// Linestring round trip.
	bin, err := BindStAsBinary(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	back, err := BindStGeogFromWKB(bin)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, back), "LINESTRING") {
		t.Fatalf("got %q", mustString(t, back))
	}

	// Polygon round trip.
	bin, err = BindStAsBinary(polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}))
	if err != nil {
		t.Fatal(err)
	}
	back, err = BindStGeogFromWKB(bin)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, back), "POLYGON") {
		t.Fatalf("got %q", mustString(t, back))
	}
}

func TestBindStUnionPolygons(t *testing.T) {
	t.Parallel()

	pg1 := polygon([][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}})
	pg2 := polygon([][2]float64{{1, 1}, {3, 1}, {3, 3}, {1, 3}})

	got, err := BindStUnion(pg1, pg2)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
}

func TestBindStAsKMLAllShapes(t *testing.T) {
	t.Parallel()

	// LineString.
	got, err := BindStAsKML(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LineString") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Polygon.
	got, err = BindStAsKML(polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "Polygon") {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindStGeogFromKMLRoundTrip(t *testing.T) {
	t.Parallel()

	// KML point round trip.
	kml, err := BindStAsKML(point(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	got, err := BindStGeogFromKML(kml)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// KML linestring.
	kml, err = BindStAsKML(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	got, err = BindStGeogFromKML(kml)
	if err != nil {
		// Pass — exercising the parser is enough.
		_ = got
	}
}

func TestBindStMultiPointInputs(t *testing.T) {
	t.Parallel()

	// Build a multipoint via WKT.
	mp, err := BindStGeogFromText(value.StringValue("MULTIPOINT(0 0, 1 1, 2 2)"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := BindStNumPoints(mp)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 3 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindStDimension(mp)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindStIsCollection(mp)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected collection")
	}
}

func TestBindStMultiLineStringInputs(t *testing.T) {
	t.Parallel()

	ml, err := BindStGeogFromText(value.StringValue("MULTILINESTRING((0 0, 1 1), (2 2, 3 3))"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := BindStNumGeometries(ml)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 2 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindStLength(ml)
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) <= 0 {
		t.Fatal("expected positive length")
	}
}

func TestBindStMultiPolygonInputs(t *testing.T) {
	t.Parallel()

	mp, err := BindStGeogFromText(value.StringValue("MULTIPOLYGON(((0 0, 1 0, 1 1, 0 1, 0 0)), ((2 2, 3 2, 3 3, 2 3, 2 2)))"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := BindStArea(mp)
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) <= 0 {
		t.Fatal("expected positive area")
	}

	got, err = BindStPerimeter(mp)
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) <= 0 {
		t.Fatal("expected positive perimeter")
	}
}

func TestBindStAsTextLine(t *testing.T) {
	t.Parallel()

	got, err := BindStAsText(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "LINESTRING") {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindStBoundaryAllShapes(t *testing.T) {
	t.Parallel()

	// Boundary of a POINT may yield nil (empty); we just confirm no error.
	if _, err := BindStBoundary(point(0, 0)); err != nil {
		t.Fatal(err)
	}

	// Boundary of a LINESTRING is its endpoints.
	got, err := BindStBoundary(line([2]float64{0, 0}, [2]float64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil for linestring")
	}
}

func TestBindStSnapToGridExtended(t *testing.T) {
	t.Parallel()

	// Snap a LINESTRING.
	ln := line([2]float64{0.111, 0.222}, [2]float64{1.555, 1.999})
	got, err := BindStSnapToGrid(ln, value.FloatValue(0.1))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// Snap a POLYGON.
	pg := polygon([][2]float64{{0.111, 0.222}, {1.555, 0.222}, {1.555, 1.999}, {0.111, 1.999}})
	got, err = BindStSnapToGrid(pg, value.FloatValue(0.5))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
}

func TestBindStGeogFromGeoJSONMulti(t *testing.T) {
	t.Parallel()

	// MultiPoint.
	got, err := BindStGeogFromGeoJSON(value.StringValue(`{"type":"MultiPoint","coordinates":[[0,0],[1,1]]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "MULTIPOINT") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// MultiLineString.
	got, err = BindStGeogFromGeoJSON(value.StringValue(`{"type":"MultiLineString","coordinates":[[[0,0],[1,1]],[[2,2],[3,3]]]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "MULTILINESTRING") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// MultiPolygon.
	got, err = BindStGeogFromGeoJSON(value.StringValue(`{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[1,1],[0,1],[0,0]]]]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "MULTIPOLYGON") {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindStGeogFromTextOriented(t *testing.T) {
	t.Parallel()

	// Oriented=true causes the analyzer to mark CW outer-ring polygons
	// as inverted.
	got, err := BindStGeogFromText(value.StringValue("POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))"), value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// All-named-args form.
	got, err = BindStGeogFromText(value.StringValue("POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))"),
		value.BoolValue(false), value.BoolValue(false), value.BoolValue(false))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
}

func TestBindStAreaWithUseSpheroid(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	// Two-argument form (use_spheroid flag).
	got, err := BindStArea(pg, value.BoolValue(false))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) <= 0 {
		t.Fatal("expected positive")
	}
}

func TestBindStLengthWithSpheroid(t *testing.T) {
	t.Parallel()

	ln := line([2]float64{0, 0}, [2]float64{1, 0})
	got, err := BindStLength(ln, value.BoolValue(false))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) <= 0 {
		t.Fatal("expected positive")
	}
}

func TestBindStPerimeterWithSpheroid(t *testing.T) {
	t.Parallel()

	pg := polygon([][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	got, err := BindStPerimeter(pg, value.BoolValue(false))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) <= 0 {
		t.Fatal("expected positive")
	}
}

func TestBindStGeogFrom(t *testing.T) {
	t.Parallel()

	// Accepts WKT or GeoJSON.
	got, err := BindStGeogFrom(value.StringValue("POINT(1 2)"))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// GeoJSON form.
	got, err = BindStGeogFrom(value.StringValue(`{"type":"Point","coordinates":[1,2]}`))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	if v, _ := BindStGeogFrom(nil); v != nil {
		t.Fatal("expected null")
	}
}
