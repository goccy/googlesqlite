package aggregate

import (
	"math"
	"math/rand"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// flattenToFloats extracts every numeric leaf from a value, walking
// through arrays. When a value is an ArrayValue, each element is
// recursively flattened. Non-numeric leaves are skipped silently.
// Used by the partial-then-final DP aggregates (variance / stddev /
// percentile / quantile) where the analyzer's RewriteAnonymization
// stages collect per-privacy-unit values into an ARRAY before the
// final aggregate consumes them.
func flattenToFloats(v value.Value) []float64 {
	if v == nil {
		return nil
	}
	if arr, ok := v.(*value.ArrayValue); ok {
		var out []float64
		for _, el := range arr.Values {
			out = append(out, flattenToFloats(el)...)
		}
		return out
	}
	if f, err := v.ToFloat64(); err == nil {
		return []float64{f}
	}
	return nil
}

// laplaceSample returns a sample from Laplace(0, scale) using a
// uniform-random inverse-CDF transform. Used by the differential
// privacy aggregates to add calibrated noise that satisfies
// (epsilon, delta)-differential privacy.
func laplaceSample(scale float64) float64 {
	if scale <= 0 {
		return 0
	}
	u := rand.Float64() - 0.5
	if u < 0 {
		return scale * math.Log(1+2*u)
	}
	return -scale * math.Log(1-2*u)
}

// dpContributionBounds reads a STRUCT<lower, upper> Value pair
// (as produced by GoogleSQL's `contribution_bounds_per_group =>
// (a, b)` named argument) and returns the (lo, hi) bounds.
// Returns (-inf, +inf) when the value isn't a STRUCT — that
// path is unreachable for the rewriter output but defensive.
func dpContributionBounds(v value.Value) (float64, float64) {
	lo, hi := math.Inf(-1), math.Inf(1)
	sv, ok := v.(*value.StructValue)
	if !ok || sv == nil {
		return lo, hi
	}
	if len(sv.Values) >= 1 && sv.Values[0] != nil {
		if f, err := sv.Values[0].ToFloat64(); err == nil {
			lo = f
		}
	}
	if len(sv.Values) >= 2 && sv.Values[1] != nil {
		if f, err := sv.Values[1].ToFloat64(); err == nil {
			hi = f
		}
	}
	return lo, hi
}

// dpStep is the inner state shared by every DP aggregate.
type dpState struct {
	sum     float64
	count   int64
	lo      float64
	hi      float64
	eps     float64
	del     float64
	allInts bool
}

// dpStepInto reads the args [value, contribution_bounds, ...,
// epsilon, delta] and folds the value into the per-aggregator
// state after clamping to the bounds. The trailing (epsilon,
// delta) pair (appended by the DifferentialPrivacyAggregateScan
// formatter) is captured on the state so Done can use it.
func dpStepInto(s *dpState, args []value.Value) error {
	if len(args) < 1 {
		return nil
	}
	for _, a := range args[1:] {
		if sv, ok := a.(*value.StructValue); ok {
			s.lo, s.hi = dpContributionBounds(sv)
			break
		}
	}
	if len(args) >= 2 {
		if f, err := args[len(args)-2].ToFloat64(); err == nil {
			s.eps = f
		}
		if f, err := args[len(args)-1].ToFloat64(); err == nil {
			s.del = f
		}
	}
	if args[0] == nil {
		return nil
	}
	// Some analyzer rewrite paths box the scalar input as an ARRAY
	// (the per-privacy-unit partial aggregate's collected values).
	// Flatten through it so every leaf contributes to the running
	// sum / count, with clamping applied per leaf.
	for _, f := range flattenToFloats(args[0]) {
		if f < s.lo {
			f = s.lo
		}
		if f > s.hi {
			f = s.hi
		}
		s.sum += f
		s.count++
	}
	return nil
}

// dpSensitivitySum returns the L1 sensitivity for a SUM/AVG
// aggregate with contribution bounds (lo, hi).
func dpSensitivitySum(lo, hi float64) float64 {
	return math.Max(math.Abs(lo), math.Abs(hi))
}

// BindDifferentialPrivacySumInt / BindDifferentialPrivacySumDouble /
// BindDifferentialPrivacyCount / ... all share a common runtime
// shape: they accumulate (clamped, summed) values during Step and
// emit a noised final value in Done. We register them as separate
// SQLite aggregate functions because the analyzer dispatches by
// signature id.
//
// The arg list at runtime is (value, contribution_bounds_struct,
// ..., epsilon, delta) — the contributor-bounds STRUCT is
// optional (analyzer omits it for COUNT_STAR; the bounds default
// to (-inf, +inf) and we fall back to a fixed sensitivity = 1).

func BindDifferentialPrivacySum() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		st := &dpState{lo: math.Inf(-1), hi: math.Inf(1), eps: 1, allInts: true}
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				return dpStepInto(st, args)
			},
			func() (value.Value, error) {
				sens := dpSensitivitySum(st.lo, st.hi)
				if math.IsInf(sens, 1) {
					sens = 1
				}
				scale := sens / math.Max(st.eps, 1e-12)
				return value.FloatValue(st.sum + laplaceSample(scale)), nil
			},
		)
	}
}

func BindDifferentialPrivacyCount() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		var n int64
		var lo, hi float64 = 0, 1
		eps := 1.0
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) == 0 || args[0] == nil {
					return nil
				}
				for _, a := range args[1:] {
					if sv, ok := a.(*value.StructValue); ok {
						lo, hi = dpContributionBounds(sv)
						break
					}
				}
				if len(args) >= 2 {
					if f, err := args[len(args)-2].ToFloat64(); err == nil {
						eps = f
					}
				}
				n++
				return nil
			},
			func() (value.Value, error) {
				sens := math.Max(math.Abs(lo), math.Abs(hi))
				if sens == 0 {
					sens = 1
				}
				scale := sens / math.Max(eps, 1e-12)
				return value.IntValue(int64(math.Round(float64(n) + laplaceSample(scale)))), nil
			},
		)
	}
}

func BindDifferentialPrivacyCountStar() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		var n int64
		var eps, _ = 1.0, 0.0
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) >= 2 {
					if f, err := args[len(args)-2].ToFloat64(); err == nil {
						eps = f
					}
				}
				n++
				return nil
			},
			func() (value.Value, error) {
				scale := 1.0 / math.Max(eps, 1e-12)
				return value.IntValue(int64(math.Round(float64(n) + laplaceSample(scale)))), nil
			},
		)
	}
}

func BindDifferentialPrivacyAvg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		st := &dpState{lo: math.Inf(-1), hi: math.Inf(1), eps: 1}
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				return dpStepInto(st, args)
			},
			func() (value.Value, error) {
				if st.count == 0 {
					return value.FloatValue(0), nil
				}
				sens := dpSensitivitySum(st.lo, st.hi)
				if math.IsInf(sens, 1) {
					sens = 1
				}
				scale := sens / math.Max(st.eps, 1e-12)
				mean := (st.sum + laplaceSample(scale)) / float64(st.count)
				return value.FloatValue(mean), nil
			},
		)
	}
}

func BindDifferentialPrivacyVarPop() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		var values []float64
		st := &dpState{lo: math.Inf(-1), hi: math.Inf(1), eps: 1}
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if err := dpStepInto(st, args); err != nil {
					return err
				}
				if args[0] != nil {
					values = append(values, flattenToFloats(args[0])...)
				}
				return nil
			},
			func() (value.Value, error) {
				if len(values) == 0 {
					return value.FloatValue(0), nil
				}
				mean := st.sum / float64(len(values))
				var ss float64
				for _, v := range values {
					ss += (v - mean) * (v - mean)
				}
				v := ss / float64(len(values))
				sens := dpSensitivitySum(st.lo, st.hi)
				if math.IsInf(sens, 1) {
					sens = 1
				}
				scale := (sens * sens) / math.Max(st.eps, 1e-12)
				return value.FloatValue(v + laplaceSample(scale)), nil
			},
		)
	}
}

func BindDifferentialPrivacyStddevPop() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		var values []float64
		st := &dpState{lo: math.Inf(-1), hi: math.Inf(1), eps: 1}
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if err := dpStepInto(st, args); err != nil {
					return err
				}
				if args[0] != nil {
					values = append(values, flattenToFloats(args[0])...)
				}
				return nil
			},
			func() (value.Value, error) {
				if len(values) == 0 {
					return value.FloatValue(0), nil
				}
				mean := st.sum / float64(len(values))
				var ss float64
				for _, v := range values {
					ss += (v - mean) * (v - mean)
				}
				v := ss / float64(len(values))
				sens := dpSensitivitySum(st.lo, st.hi)
				if math.IsInf(sens, 1) {
					sens = 1
				}
				scale := (sens * sens) / math.Max(st.eps, 1e-12)
				return value.FloatValue(math.Sqrt(math.Max(v+laplaceSample(scale), 0))), nil
			},
		)
	}
}

func BindDifferentialPrivacyApproxCountDistinct() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		seen := map[string]struct{}{}
		var eps = 1.0
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) >= 2 {
					if f, err := args[len(args)-2].ToFloat64(); err == nil {
						eps = f
					}
				}
				if len(args) == 0 || args[0] == nil {
					return nil
				}
				s, err := args[0].ToString()
				if err != nil {
					return err
				}
				seen[s] = struct{}{}
				return nil
			},
			func() (value.Value, error) {
				scale := 1.0 / math.Max(eps, 1e-12)
				return value.IntValue(int64(math.Round(float64(len(seen)) + laplaceSample(scale)))), nil
			},
		)
	}
}

// BindDifferentialPrivacyPercentileCont approximates the
// differentially-private continuous percentile. The percentile
// argument is captured from args[1]; values flow through args[0]
// (potentially wrapped in an ARRAY by the partial-aggregate stage).
// At Done we sort the collected values and linearly interpolate the
// requested percentile, then add Laplace noise scaled by the
// contribution-bound sensitivity / epsilon.
func BindDifferentialPrivacyPercentileCont() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		var values []float64
		st := &dpState{lo: math.Inf(-1), hi: math.Inf(1), eps: 1}
		percentile := 0.5
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) >= 2 && args[1] != nil {
					if f, err := args[1].ToFloat64(); err == nil {
						percentile = f
					}
				}
				if err := dpStepInto(st, args); err != nil {
					return err
				}
				if args[0] != nil {
					values = append(values, flattenToFloats(args[0])...)
				}
				return nil
			},
			func() (value.Value, error) {
				if len(values) == 0 {
					return value.FloatValue(0), nil
				}
				sorted := append([]float64(nil), values...)
				sortFloats(sorted)
				p := percentile
				if p < 0 {
					p = 0
				}
				if p > 1 {
					p = 1
				}
				idx := p * float64(len(sorted)-1)
				lo := int(math.Floor(idx))
				hi := int(math.Ceil(idx))
				v := sorted[lo]
				if hi != lo {
					v += (sorted[hi] - sorted[lo]) * (idx - float64(lo))
				}
				sens := dpSensitivitySum(st.lo, st.hi)
				if math.IsInf(sens, 1) {
					sens = 1
				}
				scale := sens / math.Max(st.eps, 1e-12)
				return value.FloatValue(v + laplaceSample(scale)), nil
			},
		)
	}
}

// BindDifferentialPrivacyQuantiles emits an ARRAY of quantile cut
// points. The second argument is the number of buckets `n`; the
// returned array has length `n+1` (`0/n, 1/n, ..., n/n`). Each
// quantile point gets independent Laplace noise.
func BindDifferentialPrivacyQuantiles() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		var values []float64
		st := &dpState{lo: math.Inf(-1), hi: math.Inf(1), eps: 1}
		buckets := int64(4)
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) >= 2 && args[1] != nil {
					if n, err := args[1].ToInt64(); err == nil && n > 0 {
						buckets = n
					}
				}
				if err := dpStepInto(st, args); err != nil {
					return err
				}
				if args[0] != nil {
					values = append(values, flattenToFloats(args[0])...)
				}
				return nil
			},
			func() (value.Value, error) {
				if len(values) == 0 {
					return &value.ArrayValue{}, nil
				}
				sorted := append([]float64(nil), values...)
				sortFloats(sorted)
				sens := dpSensitivitySum(st.lo, st.hi)
				if math.IsInf(sens, 1) {
					sens = 1
				}
				scale := sens / math.Max(st.eps, 1e-12)
				out := make([]value.Value, 0, buckets+1)
				for i := int64(0); i <= buckets; i++ {
					p := float64(i) / float64(buckets)
					idx := p * float64(len(sorted)-1)
					lo := int(math.Floor(idx))
					hi := int(math.Ceil(idx))
					v := sorted[lo]
					if hi != lo {
						v += (sorted[hi] - sorted[lo]) * (idx - float64(lo))
					}
					out = append(out, value.FloatValue(v+laplaceSample(scale)))
				}
				return &value.ArrayValue{Values: out}, nil
			},
		)
	}
}

func sortFloats(s []float64) {
	// Small inline insertion sort is enough for the per-group
	// sizes our DP examples exercise (≤ a few hundred).
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
