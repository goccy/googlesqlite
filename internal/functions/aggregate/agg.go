package aggregate

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// AGG is the property-graph MEASURE aggregator. The analyzer
// rewrites `AGG(measure_expr)` to `agg(value_expr, locking_key,
// kind)` where:
//
//   - value_expr is the underlying expression the MEASURE wraps
//     (e.g. `population` for `MEASURE(SUM(population))`).
//   - locking_key is the column tuple that identifies a unique
//     contribution (typically the node table's primary key).
//   - kind is the aggregation discriminator: "SUM", "AVG", "MIN",
//     "MAX", "COUNT", "COUNT_DISTINCT", or "ANY_VALUE". When the
//     wrapped aggregation is more exotic (HLL, etc.) AGG falls
//     back to ANY_VALUE — the first non-duplicate value wins.
//
// Locking-key dedup happens inside Step: a row whose key has
// already contributed is dropped. Done then applies `kind` to
// the surviving (key, value) pairs.
type AGG struct {
	seen   map[string]struct{}
	values []value.Value
	kind   string
}

func (a *AGG) Step(val, key value.Value, kind string, _ *helper.Option) error {
	if a.seen == nil {
		a.seen = map[string]struct{}{}
	}
	if kind != "" {
		a.kind = kind
	}
	keyStr := ""
	if key != nil {
		s, err := key.ToString()
		if err != nil {
			return err
		}
		keyStr = s
	}
	if _, ok := a.seen[keyStr]; ok {
		return nil
	}
	a.seen[keyStr] = struct{}{}
	a.values = append(a.values, val)
	return nil
}

func (a *AGG) Done() (value.Value, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	switch a.kind {
	case "", "ANY_VALUE":
		return a.values[0], nil
	case "COUNT":
		return value.IntValue(int64(len(a.values))), nil
	case "COUNT_DISTINCT":
		seen := map[string]struct{}{}
		for _, v := range a.values {
			if v == nil {
				continue
			}
			s, err := v.ToString()
			if err != nil {
				return nil, err
			}
			seen[s] = struct{}{}
		}
		return value.IntValue(int64(len(seen))), nil
	case "SUM":
		return aggSumFloat(a.values, false)
	case "AVG":
		return aggSumFloat(a.values, true)
	case "MIN":
		return aggMinMax(a.values, true)
	case "MAX":
		return aggMinMax(a.values, false)
	}
	return nil, fmt.Errorf("AGG: unsupported aggregation kind %q", a.kind)
}

// aggSumFloat sums all non-NULL values as float64. When mean is
// true it divides by the value count, yielding AVG.
func aggSumFloat(values []value.Value, mean bool) (value.Value, error) {
	var sum float64
	count := 0
	allInt := true
	for _, v := range values {
		if v == nil {
			continue
		}
		f, err := v.ToFloat64()
		if err != nil {
			return nil, err
		}
		sum += f
		count++
		if _, ok := v.(value.IntValue); !ok {
			allInt = false
		}
	}
	if count == 0 {
		return nil, nil
	}
	if mean {
		return value.FloatValue(sum / float64(count)), nil
	}
	if allInt {
		return value.IntValue(int64(sum)), nil
	}
	return value.FloatValue(sum), nil
}

func aggMinMax(values []value.Value, wantMin bool) (value.Value, error) {
	var picked value.Value
	for _, v := range values {
		if v == nil {
			continue
		}
		if picked == nil {
			picked = v
			continue
		}
		var cond bool
		var err error
		if wantMin {
			cond, err = v.LT(picked)
		} else {
			cond, err = v.GT(picked)
		}
		if err != nil {
			return nil, err
		}
		if cond {
			picked = v
		}
	}
	return picked, nil
}

// BindAGG returns the AGG aggregator constructor. The SQLite
// adapter is variadic to accept the three positional args the
// formatter emits — but it tolerates a single argument too, in
// which case AGG degrades gracefully to ANY_VALUE semantics
// (no dedup) so unrewritten AGG calls still produce a sensible
// answer.
func BindAGG() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &AGG{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if len(args) == 0 {
					return nil
				}
				val := args[0]
				var key value.Value
				kind := "ANY_VALUE"
				if len(args) >= 2 {
					key = args[1]
				}
				if len(args) >= 3 && args[2] != nil {
					s, err := args[2].ToString()
					if err != nil {
						return err
					}
					kind = s
				}
				return fn.Step(val, key, kind, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}
