package geography

import (
	"fmt"
	"math"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/functions/window"
	"github.com/goccy/googlesqlite/internal/value"
)

// WINDOW_ST_CLUSTERDBSCAN implements DBSCAN as a window aggregator
// over a partition of geographies. The upstream form is
// `ST_CLUSTERDBSCAN(geo, eps, min_pts) OVER (window)`; in emulation
// mode the runtime sees `(SELECT st_clusterdbscan(...) FROM input)`
// per outer row, with the outer row's row_id threaded in as an
// option argument. We collect every (row_id_at_step, geography)
// pair, run DBSCAN once over the partition, and return the cluster
// id matching `agg.RowID` — which the option parser sets to the
// outer row's row_id (constant across all Step calls of one
// aggregation).
type WINDOW_ST_CLUSTERDBSCAN struct {
	eps      float64
	minPts   int
	captured bool
}

// Step buffers each input row's geography and captures the
// (eps, min_points) constants on first call. The row_id option the
// formatter appends is parsed by the aggregator wrapper, so it is
// not visible here as a value argument.
func (f *WINDOW_ST_CLUSTERDBSCAN) Step(geo, epsv, minPtsv value.Value, opt *window.WindowFuncStatus, agg *window.WindowFuncAggregatedStatus) error {
	if !f.captured {
		if epsv != nil {
			v, err := epsv.ToFloat64()
			if err != nil {
				return err
			}
			f.eps = v
		}
		if minPtsv != nil {
			n, err := minPtsv.ToInt64()
			if err != nil {
				return err
			}
			if n < 0 {
				return fmt.Errorf("ST_CLUSTERDBSCAN: minimum_geographies must be non-negative, got %d", n)
			}
			minPts, err := helper.SafeInt(n)
			if err != nil {
				return err
			}
			f.minPts = minPts
		}
		f.captured = true
	}
	return agg.Step(geo, opt)
}

// Done computes DBSCAN over every collected geography in the
// partition, then returns the cluster id at the outer row's
// position (1-indexed via agg.RowID). NULL geographies and empty
// geographies are noise — emitted as SQL NULL. Cluster ids count up
// from 0 in the order new clusters are discovered.
func (f *WINDOW_ST_CLUSTERDBSCAN) Done(agg *window.WindowFuncAggregatedStatus) (value.Value, error) {
	var result value.Value
	err := agg.Done(func(values []value.Value, start, end int) error {
		geos := make([]*value.GeographyValue, len(values))
		for i, v := range values {
			geos[i] = geographyArg(v)
		}
		clusters := dbscanClusters(geos, f.eps, f.minPts)
		idx, err := helper.SafeInt(agg.RowID - 1)
		if err != nil {
			return err
		}
		if idx < 0 || idx >= len(clusters) {
			return nil
		}
		c := clusters[idx]
		if c < 0 {
			// noise / empty / NULL
			return nil
		}
		result = value.IntValue(int64(c))
		return nil
	})
	return result, err
}

// dbscanClusters assigns a 0-based cluster id to each non-empty
// geography in `geos`. NULL and empty geographies receive -1. With
// `minPts == 0` every non-empty geography forms a cluster (BigQuery
// allows the degenerate single-point cluster case).
func dbscanClusters(geos []*value.GeographyValue, eps float64, minPts int) []int {
	n := len(geos)
	clusters := make([]int, n)
	for i := range clusters {
		clusters[i] = -2 // unvisited
	}
	if minPts < 1 {
		minPts = 1
	}
	nextCluster := 0
	for i := 0; i < n; i++ {
		if clusters[i] != -2 {
			continue
		}
		if geos[i] == nil || isEmptyGeography(geos[i]) {
			clusters[i] = -1
			continue
		}
		neighbors := neighborsWithin(geos, i, eps)
		if len(neighbors) < minPts {
			clusters[i] = -1
			continue
		}
		clusters[i] = nextCluster
		// expand cluster
		queue := append([]int(nil), neighbors...)
		for len(queue) > 0 {
			j := queue[0]
			queue = queue[1:]
			if clusters[j] == -1 {
				clusters[j] = nextCluster
			}
			if clusters[j] != -2 {
				continue
			}
			clusters[j] = nextCluster
			if geos[j] == nil || isEmptyGeography(geos[j]) {
				continue
			}
			jneighbors := neighborsWithin(geos, j, eps)
			if len(jneighbors) >= minPts {
				queue = append(queue, jneighbors...)
			}
		}
		nextCluster++
	}
	return clusters
}

// neighborsWithin returns the indices of every other geography
// within `eps` meters of `geos[idx]` (including idx itself). The
// pairwise distance uses the minimum great-circle distance over
// every (vertex_i, vertex_j) pair so that geographies which share
// (or pass close to) a boundary vertex — e.g. a polygon and a
// line-string with overlapping endpoints — are correctly grouped
// together even when neither geography's call to DistanceTo (which
// may take the shape-centroid path) returns the literal touching
// distance.
func neighborsWithin(geos []*value.GeographyValue, idx int, eps float64) []int {
	out := []int{idx}
	if geos[idx] == nil {
		return out
	}
	src := pointsOf(geos[idx])
	for j, other := range geos {
		if j == idx || other == nil || isEmptyGeography(other) {
			continue
		}
		if minPairwiseDistanceMeters(src, pointsOf(other)) <= eps {
			out = append(out, j)
		}
	}
	return out
}

// minPairwiseDistanceMeters computes the smallest great-circle
// distance (in meters) between any vertex in `a` and any vertex in
// `b`. Empty inputs yield +inf so the caller treats them as
// non-overlapping.
func minPairwiseDistanceMeters(a, b [][2]float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 1e18
	}
	best := 1e18
	for _, pa := range a {
		for _, pb := range b {
			d := haversineMeters(pa[0], pa[1], pb[0], pb[1])
			if d < best {
				best = d
			}
		}
	}
	return best
}

// haversineMeters returns the great-circle distance between two
// (lng, lat) points in meters using the haversine formula and the
// WGS84 mean Earth radius. Sufficient for DBSCAN bucketing where
// the threshold is on the order of kilometers.
func haversineMeters(lng1, lat1, lng2, lat2 float64) float64 {
	const earthRadius = 6371008.8 // meters, mean radius (WGS84)
	const deg = math.Pi / 180
	dlat := (lat2 - lat1) * deg
	dlng := (lng2 - lng1) * deg
	lat1r := lat1 * deg
	lat2r := lat2 * deg
	sdlat := math.Sin(dlat / 2)
	sdlng := math.Sin(dlng / 2)
	a := sdlat*sdlat + math.Cos(lat1r)*math.Cos(lat2r)*sdlng*sdlng
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

// isEmptyGeography reports whether the geography contains no
// vertices (POINT EMPTY, GEOMETRYCOLLECTION EMPTY, etc.). DBSCAN
// treats empty geographies as noise per the upstream spec.
func isEmptyGeography(g *value.GeographyValue) bool {
	if g == nil {
		return true
	}
	return len(pointsOf(g)) == 0
}

// BindWindowStClusterDBSCAN wires the three-arg
// `ST_CLUSTERDBSCAN(geo, eps, min_pts)` into the window-aggregator
// pattern. Routed via the formatter's predecessor-emulation path so
// each outer row's row_id reaches the aggregator's RowID; the
// runtime computes DBSCAN once per partition and indexes into the
// cluster array by that row id.
func BindWindowStClusterDBSCAN() func() *window.WindowAggregator {
	return func() *window.WindowAggregator {
		fn := &WINDOW_ST_CLUSTERDBSCAN{}
		return window.NewWindowAggregator(
			func(args []value.Value, opt *window.WindowFuncStatus, agg *window.WindowFuncAggregatedStatus) error {
				if len(args) < 3 {
					return fmt.Errorf("ST_CLUSTERDBSCAN: need 3 arguments (geo, eps, min_points), got %d", len(args))
				}
				return fn.Step(args[0], args[1], args[2], opt, agg)
			},
			func(agg *window.WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}
