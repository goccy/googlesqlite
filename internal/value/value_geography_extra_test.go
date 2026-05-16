package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestGeographyConstructorsAndAccessors exercises the constructors,
// accessors, EQ, IsEmpty / Kind, and the all-unsupported scalar
// operations on GeographyValue.
func TestGeographyConstructorsAndAccessors(t *testing.T) {
	t.Parallel()

	t.Run("Point round-trip and accessors", func(t *testing.T) {
		g := value.NewGeographyPoint(1, 2)
		if g.Kind() != "POINT" {
			t.Fatalf("Kind: %s", g.Kind())
		}
		wkt, err := g.ToWKT()
		if err != nil {
			t.Fatal(err)
		}
		if wkt != "POINT (1 2)" {
			t.Fatalf("WKT: %s", wkt)
		}
		lon, lat, ok := g.PointCoordinates()
		if !ok || lon != 1 || lat != 2 {
			t.Fatalf("PointCoordinates: %f %f %v", lon, lat, ok)
		}
		if g.IsEmpty() {
			t.Fatal("point not empty")
		}
	})

	t.Run("LineString round-trip and accessors", func(t *testing.T) {
		pts := [][2]float64{{1, 2}, {3, 4}}
		g := value.NewGeographyLineString(pts)
		if g.Kind() != "LINESTRING" {
			t.Fatalf("Kind: %s", g.Kind())
		}
		out, ok := g.LineStringPoints()
		if !ok || len(out) != 2 {
			t.Fatalf("LineStringPoints: %v / ok=%v", out, ok)
		}
		wkt, err := g.ToWKT()
		if err != nil {
			t.Fatal(err)
		}
		if wkt != "LINESTRING (1 2, 3 4)" {
			t.Fatalf("WKT: %s", wkt)
		}
	})

	t.Run("Polygon round-trip", func(t *testing.T) {
		g := value.NewGeographyPolygon([][][2]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}})
		if g.Kind() != "POLYGON" {
			t.Fatalf("Kind: %s", g.Kind())
		}
		rings, ok := g.PolygonRings()
		if !ok || len(rings) != 1 {
			t.Fatalf("PolygonRings: %v / ok=%v", rings, ok)
		}
		wkt, err := g.ToWKT()
		if err != nil {
			t.Fatal(err)
		}
		if wkt != "POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))" {
			t.Fatalf("WKT: %s", wkt)
		}
		// POLYGON EMPTY branch
		empty := value.NewGeographyPolygon(nil)
		wkt, _ = empty.ToWKT()
		if wkt != "POLYGON EMPTY" {
			t.Fatalf("empty: %s", wkt)
		}
	})

	t.Run("MultiPoint / MultiLineString / MultiPolygon accessors", func(t *testing.T) {
		mp := value.NewGeographyMultiPoint([][2]float64{{1, 2}, {3, 4}})
		if mp.Kind() != "MULTIPOINT" {
			t.Fatalf("Kind: %s", mp.Kind())
		}
		if pts, ok := mp.MultiPointPoints(); !ok || len(pts) != 2 {
			t.Fatalf("MultiPointPoints: %v / ok=%v", pts, ok)
		}

		mls := value.NewGeographyMultiLineString([][][2]float64{{{1, 2}, {3, 4}}})
		if mls.Kind() != "MULTILINESTRING" {
			t.Fatalf("Kind: %s", mls.Kind())
		}
		if lines, ok := mls.MultiLineStringLines(); !ok || len(lines) != 1 {
			t.Fatalf("MultiLineStringLines: %v / ok=%v", lines, ok)
		}

		mpg := value.NewGeographyMultiPolygon([][][][2]float64{{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}}})
		if mpg.Kind() != "MULTIPOLYGON" {
			t.Fatalf("Kind: %s", mpg.Kind())
		}
		if polys, ok := mpg.MultiPolygonPolys(); !ok || len(polys) != 1 {
			t.Fatalf("MultiPolygonPolys: %v / ok=%v", polys, ok)
		}
	})

	t.Run("Collection accessors", func(t *testing.T) {
		c := value.NewGeographyCollection([]*value.GeographyValue{
			value.NewGeographyPoint(1, 2),
		})
		if c.Kind() != "GEOMETRYCOLLECTION" {
			t.Fatalf("Kind: %s", c.Kind())
		}
		parts, ok := c.CollectionParts()
		if !ok || len(parts) != 1 {
			t.Fatalf("CollectionParts: %v / ok=%v", parts, ok)
		}
	})

	t.Run("accessor type mismatches return false", func(t *testing.T) {
		p := value.NewGeographyPoint(1, 2)
		if _, _, ok := p.PointCoordinates(); !ok {
			t.Fatal("expected ok for point")
		}
		if _, ok := p.LineStringPoints(); ok {
			t.Fatal("expected !ok")
		}
		if _, ok := p.PolygonRings(); ok {
			t.Fatal("expected !ok")
		}
		if _, ok := p.MultiPointPoints(); ok {
			t.Fatal("expected !ok")
		}
		if _, ok := p.MultiLineStringLines(); ok {
			t.Fatal("expected !ok")
		}
		if _, ok := p.MultiPolygonPolys(); ok {
			t.Fatal("expected !ok")
		}
		if _, ok := p.CollectionParts(); ok {
			t.Fatal("expected !ok")
		}
	})

	t.Run("nil-receiver accessors", func(t *testing.T) {
		var g *value.GeographyValue
		if got := g.Kind(); got != "" {
			t.Fatalf("nil Kind: %s", got)
		}
		if !g.IsEmpty() {
			t.Fatal("nil should be empty")
		}
		if _, _, ok := g.PointCoordinates(); ok {
			t.Fatal("nil PointCoordinates ok")
		}
		if _, ok := g.LineStringPoints(); ok {
			t.Fatal("nil LineStringPoints ok")
		}
		if _, ok := g.PolygonRings(); ok {
			t.Fatal("nil PolygonRings ok")
		}
		if _, ok := g.MultiPointPoints(); ok {
			t.Fatal("nil MultiPointPoints ok")
		}
		if _, ok := g.MultiLineStringLines(); ok {
			t.Fatal("nil MultiLineStringLines ok")
		}
		if _, ok := g.MultiPolygonPolys(); ok {
			t.Fatal("nil MultiPolygonPolys ok")
		}
		if _, ok := g.CollectionParts(); ok {
			t.Fatal("nil CollectionParts ok")
		}
	})

	t.Run("IsEmpty branches", func(t *testing.T) {
		// POINT EMPTY (parsed via WKT)
		pe, _ := value.GeographyFromWKT("POINT EMPTY")
		if !pe.IsEmpty() {
			t.Fatal("POINT EMPTY")
		}
		// LINESTRING EMPTY
		le, _ := value.GeographyFromWKT("LINESTRING EMPTY")
		if !le.IsEmpty() {
			t.Fatal("LINESTRING EMPTY")
		}
		// POLYGON EMPTY
		poe, _ := value.GeographyFromWKT("POLYGON EMPTY")
		if !poe.IsEmpty() {
			t.Fatal("POLYGON EMPTY")
		}
		// MULTIPOINT EMPTY
		mpe, _ := value.GeographyFromWKT("MULTIPOINT EMPTY")
		if !mpe.IsEmpty() {
			t.Fatal("MULTIPOINT EMPTY")
		}
		// MULTILINESTRING EMPTY
		mle, _ := value.GeographyFromWKT("MULTILINESTRING EMPTY")
		if !mle.IsEmpty() {
			t.Fatal("MULTILINESTRING EMPTY")
		}
		// MULTIPOLYGON EMPTY
		mpge, _ := value.GeographyFromWKT("MULTIPOLYGON EMPTY")
		if !mpge.IsEmpty() {
			t.Fatal("MULTIPOLYGON EMPTY")
		}
		// GEOMETRYCOLLECTION EMPTY
		gce, _ := value.GeographyFromWKT("GEOMETRYCOLLECTION EMPTY")
		if !gce.IsEmpty() {
			t.Fatal("GEOMETRYCOLLECTION EMPTY")
		}
		// fullglobe is NOT empty
		fg, _ := value.GeographyFromWKT("fullglobe")
		if fg.IsEmpty() {
			t.Fatal("fullglobe should not be empty")
		}
	})

	t.Run("EQ", func(t *testing.T) {
		a := value.NewGeographyPoint(1, 2)
		b := value.NewGeographyPoint(1, 2)
		c := value.NewGeographyPoint(3, 4)
		if ok, _ := a.EQ(b); !ok {
			t.Fatal("equal")
		}
		if ok, _ := a.EQ(c); ok {
			t.Fatal("not equal")
		}
		// rhs not a Geography returns error
		if _, err := a.EQ(value.IntValue(0)); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("EqualPoint", func(t *testing.T) {
		a := value.NewGeographyPoint(1, 2)
		b := value.NewGeographyPoint(1, 2)
		if !a.EqualPoint(b) {
			t.Fatal("equal points")
		}
		// non-point comparison
		ls := value.NewGeographyLineString([][2]float64{{0, 0}, {1, 1}})
		if a.EqualPoint(ls) {
			t.Fatal("point vs linestring")
		}
		// nil receiver
		var nilG *value.GeographyValue
		if nilG.EqualPoint(a) {
			t.Fatal("nil receiver")
		}
		if a.EqualPoint(nil) {
			t.Fatal("nil other")
		}
	})

	t.Run("scalar operations unsupported", func(t *testing.T) {
		g := value.NewGeographyPoint(1, 2)
		if _, err := g.Add(g); err == nil {
			t.Fatal("Add")
		}
		if _, err := g.Sub(g); err == nil {
			t.Fatal("Sub")
		}
		if _, err := g.Mul(g); err == nil {
			t.Fatal("Mul")
		}
		if _, err := g.Div(g); err == nil {
			t.Fatal("Div")
		}
		if _, err := g.GT(g); err == nil {
			t.Fatal("GT")
		}
		if _, err := g.GTE(g); err == nil {
			t.Fatal("GTE")
		}
		if _, err := g.LT(g); err == nil {
			t.Fatal("LT")
		}
		if _, err := g.LTE(g); err == nil {
			t.Fatal("LTE")
		}
		if _, err := g.ToInt64(); err == nil {
			t.Fatal("ToInt64")
		}
		if _, err := g.ToFloat64(); err == nil {
			t.Fatal("ToFloat64")
		}
		if _, err := g.ToBool(); err == nil {
			t.Fatal("ToBool")
		}
		if _, err := g.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := g.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
		if _, err := g.ToTime(); err == nil {
			t.Fatal("ToTime")
		}
		if _, err := g.ToRat(); err == nil {
			t.Fatal("ToRat")
		}
	})

	t.Run("ToString/ToBytes/ToJSON/Format/Interface", func(t *testing.T) {
		g := value.NewGeographyPoint(1, 2)
		s, _ := g.ToString()
		if s != "POINT (1 2)" {
			t.Fatalf("ToString: %s", s)
		}
		b, _ := g.ToBytes()
		if string(b) != "POINT (1 2)" {
			t.Fatalf("ToBytes: %s", b)
		}
		j, _ := g.ToJSON()
		if j != `"POINT (1 2)"` {
			t.Fatalf("ToJSON: %s", j)
		}
		if got := g.Format('t'); got != "POINT (1 2)" {
			t.Fatalf("Format t: %s", got)
		}
		if got := g.Format('T'); got != `GEOGRAPHY "POINT (1 2)"` {
			t.Fatalf("Format T: %s", got)
		}
		if got := g.Format('x'); got != "POINT (1 2)" {
			t.Fatalf("Format default: %s", got)
		}
		if got, ok := g.Interface().(string); !ok || got != "POINT (1 2)" {
			t.Fatalf("Interface: %v (%T)", g.Interface(), g.Interface())
		}
		// String() returns the WKT.
		str, _ := g.String()
		if str != "POINT (1 2)" {
			t.Fatalf("String: %s", str)
		}
	})

	t.Run("MarkInverted / Inverted", func(t *testing.T) {
		g := value.NewGeographyPolygon([][][2]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}})
		if g.Inverted() {
			t.Fatal("default should not be inverted")
		}
		g.MarkInverted()
		if !g.Inverted() {
			t.Fatal("should be inverted after MarkInverted")
		}
		// MarkInverted on non-polygon is a no-op (no panic).
		p := value.NewGeographyPoint(1, 2)
		p.MarkInverted()
		if p.Inverted() {
			t.Fatal("point Inverted should be false")
		}
	})

	t.Run("EQ across geometry kinds", func(t *testing.T) {
		// LINESTRING vs LINESTRING — exercises geographyLineString.equal.
		a := value.NewGeographyLineString([][2]float64{{1, 2}, {3, 4}})
		b := value.NewGeographyLineString([][2]float64{{1, 2}, {3, 4}})
		if ok, _ := a.EQ(b); !ok {
			t.Fatal("linestrings equal")
		}
		c := value.NewGeographyLineString([][2]float64{{1, 2}, {3, 5}})
		if ok, _ := a.EQ(c); ok {
			t.Fatal("differing linestrings should not be equal")
		}
		// POLYGON vs POLYGON
		pa := value.NewGeographyPolygon([][][2]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}})
		pb := value.NewGeographyPolygon([][][2]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}})
		if ok, _ := pa.EQ(pb); !ok {
			t.Fatal("polygons equal")
		}
		// MULTI* equality
		mpa := value.NewGeographyMultiPoint([][2]float64{{1, 2}})
		mpb := value.NewGeographyMultiPoint([][2]float64{{1, 2}})
		if ok, _ := mpa.EQ(mpb); !ok {
			t.Fatal("multipoints equal")
		}
		mla := value.NewGeographyMultiLineString([][][2]float64{{{1, 2}, {3, 4}}})
		mlb := value.NewGeographyMultiLineString([][][2]float64{{{1, 2}, {3, 4}}})
		if ok, _ := mla.EQ(mlb); !ok {
			t.Fatal("multilinestrings equal")
		}
		mpga := value.NewGeographyMultiPolygon([][][][2]float64{{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}}})
		mpgb := value.NewGeographyMultiPolygon([][][][2]float64{{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}}})
		if ok, _ := mpga.EQ(mpgb); !ok {
			t.Fatal("multipolygons equal")
		}
		// Collection equality
		cola := value.NewGeographyCollection([]*value.GeographyValue{value.NewGeographyPoint(1, 2)})
		colb := value.NewGeographyCollection([]*value.GeographyValue{value.NewGeographyPoint(1, 2)})
		if ok, _ := cola.EQ(colb); !ok {
			t.Fatal("collections equal")
		}
		// fullglobe equality
		fga, _ := value.GeographyFromWKT("fullglobe")
		fgb, _ := value.GeographyFromWKT("fullglobe")
		if ok, _ := fga.EQ(fgb); !ok {
			t.Fatal("fullglobes equal")
		}
		// Different-kind not equal.
		if ok, _ := a.EQ(pa); ok {
			t.Fatal("line not equal to polygon")
		}
	})

	t.Run("DistanceTo", func(t *testing.T) {
		a := value.NewGeographyPoint(0, 0)
		b := value.NewGeographyPoint(0, 0)
		d, err := a.DistanceTo(b)
		if err != nil {
			t.Fatal(err)
		}
		if d != 0 {
			t.Fatalf("self distance: %f", d)
		}
		c := value.NewGeographyPoint(0, 1)
		d, err = a.DistanceTo(c)
		if err != nil {
			t.Fatal(err)
		}
		// 1 degree of latitude ~ 111 km.
		if d < 100_000 || d > 120_000 {
			t.Fatalf("expected ~111000 m, got %f", d)
		}
	})

	t.Run("DistanceTo errors", func(t *testing.T) {
		var nilG *value.GeographyValue
		if _, err := nilG.DistanceTo(nilG); err == nil {
			t.Fatal("nil should error")
		}
		ls := value.NewGeographyLineString([][2]float64{{0, 0}, {1, 1}})
		p := value.NewGeographyPoint(0, 0)
		if _, err := ls.DistanceTo(p); err == nil {
			t.Fatal("non-point lhs should error")
		}
		if _, err := p.DistanceTo(ls); err == nil {
			t.Fatal("non-point rhs should error")
		}
	})

	t.Run("WKT MULTIPOINT/MULTILINESTRING/MULTIPOLYGON/COLLECTION/fullglobe", func(t *testing.T) {
		cases := []string{
			"MULTIPOINT ((1 2), (3 4))",
			"MULTILINESTRING ((1 2, 3 4))",
			"MULTIPOLYGON (((0 0, 1 0, 1 1, 0 1, 0 0)))",
			"GEOMETRYCOLLECTION (POINT (1 2))",
			"fullglobe",
		}
		for _, c := range cases {
			t.Run(c, func(t *testing.T) {
				g, err := value.GeographyFromWKT(c)
				if err != nil {
					t.Fatal(err)
				}
				if g.Kind() == "" {
					t.Fatalf("empty Kind for %q", c)
				}
				// round-trip
				wkt, err := g.ToWKT()
				if err != nil {
					t.Fatal(err)
				}
				if wkt == "" {
					t.Fatal("empty WKT")
				}
			})
		}
	})
}
