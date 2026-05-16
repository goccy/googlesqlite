package geography

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// BindStExtentAgg accumulates the axis-aligned bounding box over a
// set of geographies. Returns a STRUCT<xmin FLOAT64, ymin FLOAT64,
// xmax FLOAT64, ymax FLOAT64> at the end, or NULL when no row
// contributed a vertex (all inputs NULL / POINT EMPTY).
//
// Vertices contributed by antimeridian-crossing rings (any edge with
// longitude delta > 180°) have their negative longitudes unwrapped
// by +360°, so a globe-spanning aggregation around the antimeridian
// produces the narrow box BigQuery returns (e.g. xmax can exceed 180°).
func BindStExtentAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		// Collect every vertex from every contributing input first.
		// We need to decide whether the aggregate region spans the
		// antimeridian as a whole, not per-shape: if any shape
		// crosses the date line, then every negative longitude in
		// the aggregation is reinterpreted as `lng + 360`, so the
		// resulting bbox describes the narrow side. This is the
		// BigQuery behaviour — Example 1 of ST_EXTENT mixes a
		// western-hemisphere polygon (-125..-117) with one that
		// crosses the antimeridian (172..-141); the expected bbox is
		// `xmin=172, xmax=243` (= -117 + 360), so the western
		// polygon's longitudes also unwrap to the +360 side.
		var verts [][2]float64
		anyCross := false
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) == 0 || args[0] == nil {
					return nil
				}
				g := geographyArg(args[0])
				if g == nil {
					return nil
				}
				if crossesAntimeridian(g) {
					anyCross = true
				}
				verts = append(verts, pointsOf(g)...)
				return nil
			},
			func() (value.Value, error) {
				if len(verts) == 0 {
					return nil, nil
				}
				minLng, minLat := 180.0, 90.0
				maxLng, maxLat := -180.0, -90.0
				for _, p := range verts {
					lng := p[0]
					if anyCross && lng < 0 {
						lng += 360
					}
					if lng < minLng {
						minLng = lng
					}
					if lng > maxLng {
						maxLng = lng
					}
					if p[1] < minLat {
						minLat = p[1]
					}
					if p[1] > maxLat {
						maxLat = p[1]
					}
				}
				keys := []string{"xmin", "ymin", "xmax", "ymax"}
				values := []value.Value{
					value.FloatValue(minLng),
					value.FloatValue(minLat),
					value.FloatValue(maxLng),
					value.FloatValue(maxLat),
				}
				m := map[string]value.Value{}
				for i, k := range keys {
					m[k] = values[i]
				}
				return &value.StructValue{Keys: keys, Values: values, M: m}, nil
			},
		)
	}
}
