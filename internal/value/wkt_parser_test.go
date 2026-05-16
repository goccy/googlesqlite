package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestWKTParserBranches covers parser error paths and the
// multi-element / nested-paren branches of readCoordRings,
// readPolygonList, readCollection, and readMultiPointCoords.
func TestWKTParserBranches(t *testing.T) {
	t.Parallel()

	t.Run("trailing garbage rejected", func(t *testing.T) {
		if _, err := value.GeographyFromWKT("POINT (1 2) extra"); err == nil {
			t.Fatal("expected trailing-content error")
		}
	})

	t.Run("empty input rejected", func(t *testing.T) {
		if _, err := value.GeographyFromWKT(""); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unknown tag rejected", func(t *testing.T) {
		if _, err := value.GeographyFromWKT("ZONK (1 2)"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("MULTIPOINT both syntactic forms", func(t *testing.T) {
		// flat coords
		flat, err := value.GeographyFromWKT("MULTIPOINT (1 2, 3 4)")
		if err != nil {
			t.Fatal(err)
		}
		if flat.Kind() != "MULTIPOINT" {
			t.Fatalf("Kind: %s", flat.Kind())
		}
		// parenthesised inner groups
		paren, err := value.GeographyFromWKT("MULTIPOINT ((1 2), (3 4))")
		if err != nil {
			t.Fatal(err)
		}
		if paren.Kind() != "MULTIPOINT" {
			t.Fatalf("Kind: %s", paren.Kind())
		}
	})

	t.Run("MULTIPOINT inner group with multiple coords rejected", func(t *testing.T) {
		if _, err := value.GeographyFromWKT("MULTIPOINT ((1 2, 3 4))"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("POLYGON with multiple rings", func(t *testing.T) {
		g, err := value.GeographyFromWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (2 2, 4 2, 4 4, 2 4, 2 2))")
		if err != nil {
			t.Fatal(err)
		}
		rings, _ := g.PolygonRings()
		if len(rings) != 2 {
			t.Fatalf("rings: %d", len(rings))
		}
	})

	t.Run("MULTIPOLYGON with multiple polygons", func(t *testing.T) {
		g, err := value.GeographyFromWKT("MULTIPOLYGON (((0 0, 1 0, 1 1, 0 1, 0 0)), ((2 2, 3 2, 3 3, 2 3, 2 2)))")
		if err != nil {
			t.Fatal(err)
		}
		polys, _ := g.MultiPolygonPolys()
		if len(polys) != 2 {
			t.Fatalf("polys: %d", len(polys))
		}
	})

	t.Run("GEOMETRYCOLLECTION with multiple parts", func(t *testing.T) {
		g, err := value.GeographyFromWKT("GEOMETRYCOLLECTION (POINT (1 2), LINESTRING (3 4, 5 6))")
		if err != nil {
			t.Fatal(err)
		}
		parts, _ := g.CollectionParts()
		if len(parts) != 2 {
			t.Fatalf("parts: %d", len(parts))
		}
	})

	t.Run("malformed nested geometry rejected", func(t *testing.T) {
		// Missing close paren in inner POINT
		if _, err := value.GeographyFromWKT("GEOMETRYCOLLECTION (POINT (1 2, LINESTRING (3 4, 5 6))"); err == nil {
			t.Fatal("expected error")
		}
		// Missing close paren in MULTIPOLYGON
		if _, err := value.GeographyFromWKT("MULTIPOLYGON (((0 0, 1 0, 1 1, 0 1, 0 0)"); err == nil {
			t.Fatal("expected error")
		}
		// Missing close paren in POLYGON
		if _, err := value.GeographyFromWKT("POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0)"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("MULTIPOINT missing close paren rejected", func(t *testing.T) {
		if _, err := value.GeographyFromWKT("MULTIPOINT (1 2"); err == nil {
			t.Fatal("expected error")
		}
		if _, err := value.GeographyFromWKT("MULTIPOINT ((1 2"); err == nil {
			t.Fatal("expected error")
		}
	})
}
