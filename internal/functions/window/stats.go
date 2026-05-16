package window

import (
	"math"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Native, frame-driven implementations of the statistical window
// aggregators. Each one mirrors the predecessor's semantics but runs
// incrementally: Step appends to a sliding buffer, Inverse pops the
// oldest entry, Done computes the result over the active frame.
// Because these aggregators are non-trivial to update via running
// sums (e.g. STDDEV needs running sum and sum-of-squares simultaneously
// with a numerically stable variant), we keep the buffer-and-recompute
// shape — still O(frame_size) per Done() vs the predecessor's
// O(N²) full rescan.

// floatWindow is a shared base for single-input numeric statistical
// windows. The captured Distinct / IgnoreNulls flags match the
// predecessor's parsing of the option markers.
type floatWindow struct {
	values      []float64
	hasValue    []bool // parallel to values; tracks NULL inputs
	distinct    bool
	ignoreNulls bool
	once        bool
}

func (f *floatWindow) absorbOpts(values []value.Value) (filtered []value.Value, _ error) {
	values, opt := helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if !f.once {
		f.distinct = opt.Distinct
		f.ignoreNulls = opt.IgnoreNulls
		f.once = true
	}
	return values, nil
}

func (f *floatWindow) appendStep(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, err = f.absorbOpts(values)
	if err != nil {
		return err
	}
	if len(values) == 0 || values[0] == nil {
		f.values = append(f.values, 0)
		f.hasValue = append(f.hasValue, false)
		return nil
	}
	v, err := values[0].ToFloat64()
	if err != nil {
		return err
	}
	f.values = append(f.values, v)
	f.hasValue = append(f.hasValue, true)
	return nil
}

func (f *floatWindow) popFront() {
	if len(f.values) == 0 {
		return
	}
	f.values = f.values[1:]
	f.hasValue = f.hasValue[1:]
}

// activeValues returns the non-null float64 entries in the current
// frame, applying DISTINCT when requested.
func (f *floatWindow) activeValues() []float64 {
	if !f.distinct {
		out := make([]float64, 0, len(f.values))
		for i, v := range f.values {
			if !f.hasValue[i] {
				continue
			}
			out = append(out, v)
		}
		return out
	}
	seen := map[float64]struct{}{}
	out := make([]float64, 0, len(f.values))
	for i, v := range f.values {
		if !f.hasValue[i] {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// --- single-arg statistical windows ---------------------------------

type stddevPopWindow struct{ floatWindow }

func NewStddevPopWindowNative() func() any        { return func() any { return &stddevPopWindow{} } }
func (a *stddevPopWindow) Step(args ...any) error { return a.appendStep(args...) }
func (a *stddevPopWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *stddevPopWindow) Done() (any, error) {
	xs := a.activeValues()
	if len(xs) < 1 {
		return nil, nil
	}
	mean := mean(xs)
	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(xs))), nil
}

type stddevSampWindow struct{ floatWindow }

func NewStddevSampWindowNative() func() any        { return func() any { return &stddevSampWindow{} } }
func (a *stddevSampWindow) Step(args ...any) error { return a.appendStep(args...) }
func (a *stddevSampWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *stddevSampWindow) Done() (any, error) {
	xs := a.activeValues()
	if len(xs) < 2 {
		return nil, nil
	}
	mean := mean(xs)
	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(xs)-1)), nil
}

type varPopWindow struct{ floatWindow }

func NewVarPopWindowNative() func() any        { return func() any { return &varPopWindow{} } }
func (a *varPopWindow) Step(args ...any) error { return a.appendStep(args...) }
func (a *varPopWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *varPopWindow) Done() (any, error) {
	xs := a.activeValues()
	if len(xs) < 1 {
		return nil, nil
	}
	mean := mean(xs)
	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return ss / float64(len(xs)), nil
}

type varSampWindow struct{ floatWindow }

func NewVarSampWindowNative() func() any        { return func() any { return &varSampWindow{} } }
func (a *varSampWindow) Step(args ...any) error { return a.appendStep(args...) }
func (a *varSampWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *varSampWindow) Done() (any, error) {
	xs := a.activeValues()
	if len(xs) < 2 {
		return nil, nil
	}
	mean := mean(xs)
	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return ss / float64(len(xs)-1), nil
}

// --- two-arg statistical windows (CORR, COVAR_*) -------------------

// pairWindow is the analogue of floatWindow for functions that take
// two numeric arguments. NULLs in either coordinate cause the pair
// to be skipped (matches BigQuery semantics).
type pairWindow struct {
	xs, ys      []float64
	has         []bool
	once        bool
	ignoreNulls bool
	distinct    bool
}

func (p *pairWindow) appendStep(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, opt := helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if !p.once {
		p.ignoreNulls = opt.IgnoreNulls
		p.distinct = opt.Distinct
		p.once = true
	}
	if len(values) < 2 || values[0] == nil || values[1] == nil {
		p.xs = append(p.xs, 0)
		p.ys = append(p.ys, 0)
		p.has = append(p.has, false)
		return nil
	}
	xv, err := values[0].ToFloat64()
	if err != nil {
		return err
	}
	yv, err := values[1].ToFloat64()
	if err != nil {
		return err
	}
	p.xs = append(p.xs, xv)
	p.ys = append(p.ys, yv)
	p.has = append(p.has, true)
	return nil
}

func (p *pairWindow) popFront() {
	if len(p.xs) == 0 {
		return
	}
	p.xs = p.xs[1:]
	p.ys = p.ys[1:]
	p.has = p.has[1:]
}

func (p *pairWindow) activePairs() (xs, ys []float64) {
	xs = make([]float64, 0, len(p.xs))
	ys = make([]float64, 0, len(p.ys))
	for i := range p.xs {
		if !p.has[i] {
			continue
		}
		xs = append(xs, p.xs[i])
		ys = append(ys, p.ys[i])
	}
	return xs, ys
}

type corrWindow struct{ pairWindow }

func NewCorrWindowNative() func() any        { return func() any { return &corrWindow{} } }
func (a *corrWindow) Step(args ...any) error { return a.appendStep(args...) }
func (a *corrWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *corrWindow) Done() (any, error) {
	xs, ys := a.activePairs()
	if len(xs) < 2 {
		return nil, nil
	}
	mx, my := mean(xs), mean(ys)
	var sxy, sxx, syy float64
	for i := range xs {
		dx := xs[i] - mx
		dy := ys[i] - my
		sxy += dx * dy
		sxx += dx * dx
		syy += dy * dy
	}
	denom := math.Sqrt(sxx * syy)
	if denom == 0 {
		return nil, nil
	}
	return sxy / denom, nil
}

type covarPopWindow struct{ pairWindow }

func NewCovarPopWindowNative() func() any        { return func() any { return &covarPopWindow{} } }
func (a *covarPopWindow) Step(args ...any) error { return a.appendStep(args...) }
func (a *covarPopWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *covarPopWindow) Done() (any, error) {
	xs, ys := a.activePairs()
	if len(xs) < 1 {
		return nil, nil
	}
	mx, my := mean(xs), mean(ys)
	var sxy float64
	for i := range xs {
		sxy += (xs[i] - mx) * (ys[i] - my)
	}
	return sxy / float64(len(xs)), nil
}

type covarSampWindow struct{ pairWindow }

func NewCovarSampWindowNative() func() any        { return func() any { return &covarSampWindow{} } }
func (a *covarSampWindow) Step(args ...any) error { return a.appendStep(args...) }
func (a *covarSampWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *covarSampWindow) Done() (any, error) {
	xs, ys := a.activePairs()
	if len(xs) < 2 {
		return nil, nil
	}
	mx, my := mean(xs), mean(ys)
	var sxy float64
	for i := range xs {
		sxy += (xs[i] - mx) * (ys[i] - my)
	}
	return sxy / float64(len(xs)-1), nil
}

func mean(xs []float64) float64 {
	var s float64
	for _, v := range xs {
		s += v
	}
	return s / float64(len(xs))
}
