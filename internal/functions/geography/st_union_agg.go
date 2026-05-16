package geography

import (
	"sort"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// BindStUnionAgg returns the geographic union of every input. The
// aggregator collects all inputs and at Done either:
//
//   - returns a single merged LINESTRING / MULTILINESTRING when
//     every input is a (multi)linestring, by chaining shared
//     endpoints into one or more polylines. Duplicate segments are
//     folded away.
//   - falls back to the iterative pairwise BindStUnion for any other
//     kind (points / polygons / mixed inputs).
func BindStUnionAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		var inputs []value.Value
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) == 0 || args[0] == nil {
					return nil
				}
				inputs = append(inputs, args[0])
				return nil
			},
			func() (value.Value, error) {
				if len(inputs) == 0 {
					return nil, nil
				}
				if merged := mergeLines(inputs); merged != nil {
					return merged, nil
				}
				acc := inputs[0]
				for _, in := range inputs[1:] {
					v, err := BindStUnion(acc, in)
					if err != nil {
						return nil, err
					}
					acc = v
				}
				return acc, nil
			},
		)
	}
}

// mergeLines folds a list of (MULTI)LINESTRING inputs into a single
// LINESTRING when the segments form a single connected polyline,
// otherwise a MULTILINESTRING of distinct chains. Returns nil when
// any input is not a (MULTI)LINESTRING — caller falls back to the
// pairwise union path.
func mergeLines(inputs []value.Value) value.Value {
	edges := []edge2{}
	for _, in := range inputs {
		g, ok := in.(*value.GeographyValue)
		if !ok {
			return nil
		}
		switch g.Kind() {
		case "LINESTRING":
			pts, _ := g.LineStringPoints()
			for i := 1; i < len(pts); i++ {
				edges = append(edges, edge2{pts[i-1], pts[i]})
			}
		case "MULTILINESTRING":
			lines, _ := g.MultiLineStringLines()
			for _, ls := range lines {
				for i := 1; i < len(ls); i++ {
					edges = append(edges, edge2{ls[i-1], ls[i]})
				}
			}
		default:
			return nil
		}
	}
	if len(edges) == 0 {
		return nil
	}
	// Deduplicate edges, ignoring direction.
	seen := map[edge2]bool{}
	var uniq []edge2
	for _, e := range edges {
		k := canonEdge2(e.a, e.b)
		if seen[k] {
			continue
		}
		seen[k] = true
		uniq = append(uniq, e)
	}
	chains := chainSegments(uniq)
	if len(chains) == 1 {
		return value.NewGeographyLineString(chains[0])
	}
	return value.NewGeographyMultiLineString(chains)
}

// chainSegments greedily walks the segment graph from degree-1
// endpoints to produce one or more open polylines covering every
// edge. Cycles (no degree-1 vertex) are walked from an arbitrary
// start vertex.
func chainSegments(edges []edge2) [][][2]float64 {
	type pt = [2]float64
	adj := map[pt][]pt{}
	addAdj := func(u, v pt) {
		adj[u] = append(adj[u], v)
	}
	for _, e := range edges {
		addAdj(e.a, e.b)
		addAdj(e.b, e.a)
	}
	visited := map[edge2]bool{}
	consume := func(u, v pt) bool {
		k := canonEdge2(u, v)
		if visited[k] {
			return false
		}
		visited[k] = true
		return true
	}
	// Sort vertices for deterministic traversal order.
	starts := make([]pt, 0, len(adj))
	for v := range adj {
		starts = append(starts, v)
	}
	// Start at the lex-GREATEST endpoint of each chain so the
	// emitted walk matches upstream BigQuery's display convention
	// (e.g. -100.19 before -122.12 before -122.19 for the
	// LINESTRING-union Example).
	sort.Slice(starts, func(i, j int) bool {
		if starts[i][0] != starts[j][0] {
			return starts[i][0] > starts[j][0]
		}
		return starts[i][1] > starts[j][1]
	})
	endpointsFirst := append([]pt{}, starts...)
	sort.SliceStable(endpointsFirst, func(i, j int) bool {
		return len(adj[endpointsFirst[i]]) < len(adj[endpointsFirst[j]])
	})
	var chains [][]pt
	walk := func(start pt) []pt {
		path := []pt{start}
		cur := start
		for {
			var next pt
			found := false
			for _, n := range adj[cur] {
				if consume(cur, n) {
					next = n
					found = true
					break
				}
			}
			if !found {
				return path
			}
			path = append(path, next)
			cur = next
		}
	}
	for _, s := range endpointsFirst {
		for {
			has := false
			for _, n := range adj[s] {
				if !visited[canonEdge2(s, n)] {
					has = true
					break
				}
			}
			if !has {
				break
			}
			chains = append(chains, walk(s))
		}
	}
	return chains
}

type edge2 struct {
	a, b [2]float64
}

func canonEdge2(a, b [2]float64) edge2 {
	if a[0] > b[0] || (a[0] == b[0] && a[1] > b[1]) {
		a, b = b, a
	}
	return edge2{a, b}
}
