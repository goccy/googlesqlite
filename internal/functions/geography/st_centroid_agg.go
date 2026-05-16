package geography

import (
	"fmt"
	"math"
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// ST_CENTROID_AGG computes the dimension-weighted centroid of a
// set of GEOGRAPHY values. Per BigQuery semantics:
//
//   - dimension 2 (POLYGON / MULTIPOLYGON) inputs dominate; if any
//     are present, dimension 0/1 inputs are excluded.
//   - dimension 1 (LINESTRING / MULTILINESTRING) inputs dominate
//     POINTs; POINTs are excluded when at least one LINESTRING is
//     present.
//   - dimension 0 (POINT / MULTIPOINT) inputs contribute equally
//     when no higher-dim input is present.
//
// Weighting:
//
//   - polygons: per-ring signed-area (planar Shoelace formula);
//     ring centroid is the standard area-weighted centroid; final
//     centroid is signed-area-weighted across rings (outer rings
//     positive, inner rings negative).
//   - lines: per-segment length; segment centroid is the midpoint.
//   - points: equal weight per point.
//
// This uses planar (longitude / latitude treated as Cartesian)
// math — the same caveat as the rest of our GEOGRAPHY surface
// (we don't depend on S2). Output is always a POINT.
type ST_CENTROID_AGG struct {
	once        sync.Once
	opt         *helper.Option
	pointSumLon float64
	pointSumLat float64
	pointCount  int64
	lineSumLon  float64
	lineSumLat  float64
	lineWeight  float64
	areaSumLon  float64
	areaSumLat  float64
	areaWeight  float64
}

func (f *ST_CENTROID_AGG) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	f.once.Do(func() { f.opt = opt })
	g, ok := v.(*value.GeographyValue)
	if !ok {
		return fmt.Errorf("ST_CENTROID_AGG: input must be a GEOGRAPHY value")
	}
	if lon, lat, ok := g.PointCoordinates(); ok {
		f.pointSumLon += lon
		f.pointSumLat += lat
		f.pointCount++
		return nil
	}
	if pts, ok := g.LineStringPoints(); ok {
		f.accumulateLine(pts)
		return nil
	}
	if lines, ok := g.MultiLineStringLines(); ok {
		for _, ls := range lines {
			f.accumulateLine(ls)
		}
		return nil
	}
	if rings, ok := g.PolygonRings(); ok {
		f.accumulatePolygon(rings)
		return nil
	}
	if polys, ok := g.MultiPolygonPolys(); ok {
		for _, rings := range polys {
			f.accumulatePolygon(rings)
		}
		return nil
	}
	if pts, ok := g.MultiPointPoints(); ok {
		for _, p := range pts {
			f.pointSumLon += p[0]
			f.pointSumLat += p[1]
			f.pointCount++
		}
		return nil
	}
	return fmt.Errorf("ST_CENTROID_AGG: unsupported GEOGRAPHY kind %q", g.Kind())
}

func (f *ST_CENTROID_AGG) accumulateLine(points [][2]float64) {
	for i := 1; i < len(points); i++ {
		ax, ay := points[i-1][0], points[i-1][1]
		bx, by := points[i][0], points[i][1]
		dx, dy := bx-ax, by-ay
		length := math.Hypot(dx, dy)
		if length == 0 {
			continue
		}
		midX, midY := (ax+bx)/2, (ay+by)/2
		f.lineSumLon += midX * length
		f.lineSumLat += midY * length
		f.lineWeight += length
	}
}

// accumulatePolygon uses the planar shoelace centroid for each
// ring and signs the contribution by ring orientation (outer
// rings positive, inner rings negative). Self-intersection is
// not handled — same as the rest of the pure-Go GEOGRAPHY pipe.
func (f *ST_CENTROID_AGG) accumulatePolygon(rings [][][2]float64) {
	for i, ring := range rings {
		cx, cy, area := ringCentroidAndArea(ring)
		if area == 0 {
			continue
		}
		sign := 1.0
		if i > 0 {
			// Inner rings (holes) subtract from outer rings.
			sign = -1.0
		}
		f.areaSumLon += cx * area * sign
		f.areaSumLat += cy * area * sign
		f.areaWeight += area * sign
	}
}

// ringCentroidAndArea computes the planar Shoelace area of a
// closed ring and its area-weighted centroid.
func ringCentroidAndArea(ring [][2]float64) (float64, float64, float64) {
	if len(ring) < 3 {
		return 0, 0, 0
	}
	var cx, cy, twoA float64
	for i := 0; i < len(ring)-1; i++ {
		x0, y0 := ring[i][0], ring[i][1]
		x1, y1 := ring[i+1][0], ring[i+1][1]
		cross := x0*y1 - x1*y0
		twoA += cross
		cx += (x0 + x1) * cross
		cy += (y0 + y1) * cross
	}
	if twoA == 0 {
		return 0, 0, 0
	}
	area := math.Abs(twoA) / 2
	cx /= 3 * twoA
	cy /= 3 * twoA
	return cx, cy, area
}

func (f *ST_CENTROID_AGG) Done() (value.Value, error) {
	switch {
	case f.areaWeight != 0:
		return value.NewGeographyPoint(f.areaSumLon/f.areaWeight, f.areaSumLat/f.areaWeight), nil
	case f.lineWeight != 0:
		return value.NewGeographyPoint(f.lineSumLon/f.lineWeight, f.lineSumLat/f.lineWeight), nil
	case f.pointCount != 0:
		return value.NewGeographyPoint(f.pointSumLon/float64(f.pointCount), f.pointSumLat/float64(f.pointCount)), nil
	}
	return nil, nil
}

func BindStCentroidAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &ST_CENTROID_AGG{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return fn.Step(args[0], opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}
