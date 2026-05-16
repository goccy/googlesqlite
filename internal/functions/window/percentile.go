package window

import (
	"fmt"
	"math"
	"sort"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// percentileWindow is the shared accumulator for PERCENTILE_CONT and
// PERCENTILE_DISC over OVER. Step buffers per-row (value, percentile)
// pairs; only the first non-NULL percentile is captured because
// BigQuery requires it to be a constant. Done sorts the active
// values and computes the requested percentile.
type percentileWindow struct {
	values     []value.Value
	null       []bool
	pct        float64
	pctSet     bool
	ignoreNull bool
	once       bool
}

func (p *percentileWindow) absorbStep(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, opt := helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if !p.once {
		p.ignoreNull = opt.IgnoreNulls
		p.once = true
	}
	if len(values) < 2 {
		return fmt.Errorf("percentile: requires (value, percentile) arguments")
	}
	if !p.pctSet && values[1] != nil {
		f, err := values[1].ToFloat64()
		if err != nil {
			return err
		}
		if f < 0 || f > 1 {
			return fmt.Errorf("percentile: value must be in [0, 1]; got %v", f)
		}
		p.pct = f
		p.pctSet = true
	}
	if values[0] == nil {
		p.values = append(p.values, nil)
		p.null = append(p.null, true)
	} else {
		p.values = append(p.values, values[0])
		p.null = append(p.null, false)
	}
	return nil
}

func (p *percentileWindow) popFront() {
	if len(p.values) == 0 {
		return
	}
	p.values = p.values[1:]
	p.null = p.null[1:]
}

// activeSorted returns the active values in BigQuery PERCENTILE_*
// sort order. When IGNORE NULLS (the default the formatter emits)
// is set, NULLs are dropped. Under RESPECT NULLS they are placed at
// the front of the sorted slice.
func (p *percentileWindow) activeSorted() ([]value.Value, error) {
	var nulls, nonNulls []value.Value
	for i, v := range p.values {
		if p.null[i] {
			if !p.ignoreNull {
				nulls = append(nulls, nil)
			}
			continue
		}
		nonNulls = append(nonNulls, v)
	}
	sort.SliceStable(nonNulls, func(i, j int) bool {
		c, err := nonNulls[i].LT(nonNulls[j])
		if err != nil {
			return false
		}
		return c
	})
	return append(nulls, nonNulls...), nil
}

// PERCENTILE_CONT: continuous percentile.
type percentileContWindow struct{ percentileWindow }

func NewPercentileContWindowNative() func() any {
	return func() any { return &percentileContWindow{} }
}
func (a *percentileContWindow) Step(args ...any) error { return a.absorbStep(args...) }
func (a *percentileContWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *percentileContWindow) Done() (any, error) {
	xs, err := a.activeSorted()
	if err != nil {
		return nil, err
	}
	if len(xs) == 0 {
		return nil, nil
	}
	if !a.pctSet {
		return nil, nil
	}
	// Position within [0, n-1].
	pos := a.pct * float64(len(xs)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	loVal, err := xs[lo].ToFloat64()
	if err != nil {
		return nil, err
	}
	if lo == hi {
		return loVal, nil
	}
	hiVal, err := xs[hi].ToFloat64()
	if err != nil {
		return nil, err
	}
	frac := pos - float64(lo)
	return loVal + frac*(hiVal-loVal), nil
}

// PERCENTILE_DISC: discrete percentile.
type percentileDiscWindow struct{ percentileWindow }

func NewPercentileDiscWindowNative() func() any {
	return func() any { return &percentileDiscWindow{} }
}
func (a *percentileDiscWindow) Step(args ...any) error { return a.absorbStep(args...) }
func (a *percentileDiscWindow) Inverse(_ ...any) error { a.popFront(); return nil }
func (a *percentileDiscWindow) Done() (any, error) {
	xs, err := a.activeSorted()
	if err != nil {
		return nil, err
	}
	if len(xs) == 0 {
		return nil, nil
	}
	if !a.pctSet {
		return nil, nil
	}
	// Smallest value v such that ceil(pct * n) values are <= v.
	// Equivalent to taking the value at index ceil(pct * n) - 1 when
	// 1-indexed, or floor(pct * (n-1)) zero-indexed; BigQuery uses
	// `index = ceil(pct * n) - 1` which matches.
	n := len(xs)
	idx := max(int(math.Ceil(a.pct*float64(n)))-1, 0)
	if idx >= n {
		idx = n - 1
	}
	// PERCENTILE_DISC returns the value at the picked position
	// unchanged; only PERCENTILE_CONT requires a numeric coercion
	// for interpolation.
	return value.EncodeValue(xs[idx])
}
