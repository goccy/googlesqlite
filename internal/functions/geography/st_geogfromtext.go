package geography

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func ST_GEOGFROMTEXT(wkt string) (value.Value, error) {
	res, err := value.GeographyFromWKT(wkt)
	if err != nil {
		return nil, fmt.Errorf("ST_GEOGFROMTEXT failed: %w", err)
	}

	return res, nil
}

// BindStGeogFromText accepts:
//
//	ST_GEOGFROMTEXT(wkt_string)
//	ST_GEOGFROMTEXT(wkt_string, oriented)
//	ST_GEOGFROMTEXT(wkt_string [, oriented => bool] [, planar => bool]
//	                            [, make_valid => bool])
//
// The analyzer materialises every declared named argument as a
// positional after the leading WKT, so the runtime can see up to
// four args here. When `oriented` is TRUE and the parsed polygon's
// outer ring is clockwise (signed planar area on antimeridian-
// unwrapped vertices), the geography is marked as inverted — the
// upstream convention is that the small interior region is the
// *exterior* of the ring, and ST_BOUNDINGBOX treats that as the
// whole globe.
//
// The 4-argument form positions named arguments in their declared
// order: (wkt, oriented, planar, make_valid) when only the legacy
// positional `oriented` second-arg form is used, otherwise the
// analyzer fills all four from named arguments. We accept either
// flavour by treating any boolean arg with value TRUE as enabling
// `oriented`.
func BindStGeogFromText(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 4 {
		return nil, fmt.Errorf("ST_GEOGFROMTEXT: invalid number of arguments: got %d, want between 1 and 4", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	wkt, err := args[0].ToString()
	if err != nil {
		return nil, fmt.Errorf("ST_GEOGFROMTEXT: invalid argument: %w", err)
	}
	oriented := false
	for _, a := range args[1:] {
		if a == nil {
			continue
		}
		if b, err := a.ToBool(); err == nil && b {
			oriented = true
			break
		}
	}
	g, err := ST_GEOGFROMTEXT(wkt)
	if err != nil {
		return nil, err
	}
	if oriented && g != nil {
		applyOrientation(g.(*value.GeographyValue))
	}
	return g, nil
}

// applyOrientation marks the polygon as `inverted` when its outer
// ring is clockwise. CW vs CCW is decided by the Shoelace signed
// planar area on antimeridian-unwrapped longitudes; positive sum =
// CCW (interior is the enclosed region), negative = CW (interior is
// the complement).
func applyOrientation(g *value.GeographyValue) {
	if g == nil || g.Kind() != "POLYGON" {
		return
	}
	rings, _ := g.PolygonRings()
	if len(rings) == 0 {
		return
	}
	outer := rings[0]
	if len(outer) < 4 {
		return
	}
	unwrapped := unwrapLongitudes(outer)
	if signedPlanarArea(unwrapped) < 0 {
		g.MarkInverted()
	}
}

// unwrapLongitudes shifts longitudes that lag a >180° step from the
// previous vertex by +360 (or -360), so the resulting polyline does
// not jump across the antimeridian. The first vertex is left
// unchanged.
func unwrapLongitudes(ring [][2]float64) [][2]float64 {
	out := make([][2]float64, len(ring))
	if len(ring) == 0 {
		return out
	}
	out[0] = ring[0]
	prev := ring[0][0]
	for i := 1; i < len(ring); i++ {
		lng := ring[i][0]
		for lng-prev > 180 {
			lng -= 360
		}
		for prev-lng > 180 {
			lng += 360
		}
		out[i] = [2]float64{lng, ring[i][1]}
		prev = lng
	}
	return out
}

// signedPlanarArea is the Shoelace formula on the supplied ring,
// treating (lng, lat) as planar coordinates. Positive when the ring
// is wound CCW (math convention), negative for CW.
func signedPlanarArea(ring [][2]float64) float64 {
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
