package window

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Native, frame-driven DISTINCT variants of the standard aggregate
// window functions. SQLite's built-in SUM/COUNT/AVG don't accept
// `DISTINCT` inside `OVER (...)`, so the predecessor handled
// `SUM(DISTINCT x) OVER (...)` through its per-output-row
// correlated-subquery emulation. Wiring these custom variants
// through `Conn.CreateWindowFunction` lets the SQLite frame engine
// drive Step / Inverse / Done while we apply DISTINCT semantics in
// Done over the active frame.

// distinctNumericWindow buffers the typed numeric values currently in
// the frame. It's the shared base for SUM_DISTINCT / COUNT_DISTINCT /
// AVG_DISTINCT — each defines its own Done() over the deduped slice.
type distinctNumericWindow struct {
	values []float64
	hasInt []bool // parallel: true → original was integer (preserves sum int64 case)
	ints   []int64
	null   []bool // true → the row's value was NULL
}

func (d *distinctNumericWindow) appendValue(v value.Value) error {
	if v == nil {
		d.values = append(d.values, 0)
		d.ints = append(d.ints, 0)
		d.hasInt = append(d.hasInt, false)
		d.null = append(d.null, true)
		return nil
	}
	// Try to capture the value as int64 first; fall back to float64.
	if iv, err := v.ToInt64(); err == nil {
		d.values = append(d.values, float64(iv))
		d.ints = append(d.ints, iv)
		d.hasInt = append(d.hasInt, true)
		d.null = append(d.null, false)
		return nil
	}
	f, err := v.ToFloat64()
	if err != nil {
		return err
	}
	d.values = append(d.values, f)
	d.ints = append(d.ints, 0)
	d.hasInt = append(d.hasInt, false)
	d.null = append(d.null, false)
	return nil
}

func (d *distinctNumericWindow) popFront() {
	if len(d.values) == 0 {
		return
	}
	d.values = d.values[1:]
	d.ints = d.ints[1:]
	d.hasInt = d.hasInt[1:]
	d.null = d.null[1:]
}

// activeDistinct returns the deduplicated, non-null active values.
// The bool sister return is true when every captured value originated
// as an integer — callers (SUM) preserve int64 output in that case.
func (d *distinctNumericWindow) activeDistinct() (vals []float64, allInt bool, intVals []int64) {
	allInt = true
	seen := map[float64]struct{}{}
	for i, v := range d.values {
		if d.null[i] {
			continue
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		vals = append(vals, v)
		if !d.hasInt[i] {
			allInt = false
		} else {
			intVals = append(intVals, d.ints[i])
		}
	}
	if !allInt {
		intVals = nil
	}
	return vals, allInt, intVals
}

// stepArgs strips option markers and forwards the value to appendValue.
func (d *distinctNumericWindow) stepArgs(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 {
		return d.appendValue(nil)
	}
	return d.appendValue(values[0])
}

// --- SUM(DISTINCT x) OVER ... -------------------------------------

type sumDistinctWindow struct{ distinctNumericWindow }

func NewSumDistinctWindowNative() func() any {
	return func() any { return &sumDistinctWindow{} }
}

func (a *sumDistinctWindow) Step(args ...any) error { return a.stepArgs(args...) }
func (a *sumDistinctWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *sumDistinctWindow) Done() (any, error) {
	vals, allInt, intVals := a.activeDistinct()
	if len(vals) == 0 {
		return nil, nil
	}
	if allInt {
		var s int64
		for _, v := range intVals {
			s += v
		}
		return s, nil
	}
	var s float64
	for _, v := range vals {
		s += v
	}
	return s, nil
}

// --- COUNT(DISTINCT x) OVER ... ------------------------------------

type countDistinctWindow struct{ distinctNumericWindow }

func NewCountDistinctWindowNative() func() any {
	return func() any { return &countDistinctWindow{} }
}

func (a *countDistinctWindow) Step(args ...any) error { return a.stepArgs(args...) }
func (a *countDistinctWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *countDistinctWindow) Done() (any, error) {
	vals, _, _ := a.activeDistinct()
	return int64(len(vals)), nil
}

// --- AVG(DISTINCT x) OVER ... --------------------------------------

type avgDistinctWindow struct{ distinctNumericWindow }

func NewAvgDistinctWindowNative() func() any {
	return func() any { return &avgDistinctWindow{} }
}

func (a *avgDistinctWindow) Step(args ...any) error { return a.stepArgs(args...) }
func (a *avgDistinctWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *avgDistinctWindow) Done() (any, error) {
	vals, _, _ := a.activeDistinct()
	if len(vals) == 0 {
		return nil, nil
	}
	var s float64
	for _, v := range vals {
		s += v
	}
	return s / float64(len(vals)), nil
}
