package rangefn

import (
	"fmt"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// BindRangeContains implements RANGE_CONTAINS(outer, inner) and the
// overload RANGE_CONTAINS(range, point). The outer-vs-inner overload
// returns TRUE when the outer range covers every element of the inner
// range; the point overload returns TRUE when `point ∈ range`.
func BindRangeContains(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("RANGE_CONTAINS: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	outer, ok := args[0].(*value.RangeValue)
	if !ok {
		return nil, fmt.Errorf("RANGE_CONTAINS: first argument must be RANGE")
	}
	if inner, ok := args[1].(*value.RangeValue); ok {
		if !rangeBoundLEQ(outer.Start, inner.Start, true) {
			return value.BoolValue(false), nil
		}
		if !rangeBoundLEQ(inner.End, outer.End, false) {
			return value.BoolValue(false), nil
		}
		return value.BoolValue(true), nil
	}
	point := args[1]
	if !rangeBoundLEQ(outer.Start, point, true) {
		return value.BoolValue(false), nil
	}
	if outer.End != nil {
		lt, err := pointLT(point, outer.End)
		if err != nil {
			return nil, err
		}
		if !lt {
			return value.BoolValue(false), nil
		}
	}
	return value.BoolValue(true), nil
}

// BindRangeOverlaps returns TRUE iff the two ranges have any point in
// common.
func BindRangeOverlaps(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("RANGE_OVERLAPS: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	a, aok := args[0].(*value.RangeValue)
	b, bok := args[1].(*value.RangeValue)
	if !aok || !bok {
		return nil, fmt.Errorf("RANGE_OVERLAPS: arguments must be RANGE")
	}
	// Two half-open intervals overlap iff a.start < b.end && b.start < a.end.
	if a.End != nil && b.Start != nil {
		lt, err := pointLT(b.Start, a.End)
		if err != nil {
			return nil, err
		}
		if !lt {
			return value.BoolValue(false), nil
		}
	}
	if b.End != nil && a.Start != nil {
		lt, err := pointLT(a.Start, b.End)
		if err != nil {
			return nil, err
		}
		if !lt {
			return value.BoolValue(false), nil
		}
	}
	return value.BoolValue(true), nil
}

// BindRangeIntersect returns the intersection of two RANGEs, or NULL
// when the ranges do not overlap.
func BindRangeIntersect(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("RANGE_INTERSECT: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	a, aok := args[0].(*value.RangeValue)
	b, bok := args[1].(*value.RangeValue)
	if !aok || !bok {
		return nil, fmt.Errorf("RANGE_INTERSECT: arguments must be RANGE")
	}
	overlap, err := BindRangeOverlaps(a, b)
	if err != nil {
		return nil, err
	}
	if ov, ok := overlap.(value.BoolValue); !ok || !bool(ov) {
		return nil, nil
	}
	start := maxBound(a.Start, b.Start, true)
	end := minBound(a.End, b.End, false)
	return &value.RangeValue{Start: start, End: end, ElemHeader: a.ElemHeader}, nil
}

// rangeBoundLEQ returns whether x ≤ y in the bound ordering. When
// `startSide` is true, NULL is the smallest value (-inf); otherwise
// NULL is the largest value (+inf).
func rangeBoundLEQ(x, y value.Value, startSide bool) bool {
	if x == nil {
		return startSide
	}
	if y == nil {
		return !startSide
	}
	leq, err := x.LTE(y)
	if err != nil {
		return false
	}
	return leq
}

// pointLT returns whether a < b for two scalar (non-range) values.
func pointLT(a, b value.Value) (bool, error) {
	if a == nil || b == nil {
		return false, nil
	}
	return a.LT(b)
}

func maxBound(a, b value.Value, startSide bool) value.Value {
	if a == nil {
		if startSide {
			return b
		}
		return a
	}
	if b == nil {
		if startSide {
			return a
		}
		return b
	}
	gt, err := a.GT(b)
	if err != nil {
		return a
	}
	if gt {
		return a
	}
	return b
}

func minBound(a, b value.Value, startSide bool) value.Value {
	if a == nil {
		if startSide {
			return a
		}
		return b
	}
	if b == nil {
		if startSide {
			return b
		}
		return a
	}
	lt, err := a.LT(b)
	if err != nil {
		return a
	}
	if lt {
		return a
	}
	return b
}

// BindGenerateRangeArray splits a RANGE into an ARRAY<RANGE> by
// stepping the start bound forward by `step` until reaching the end.
// step is an INTERVAL value; the implementation honours the common
// daily / hourly / monthly cases by delegating to the bound's own
// addition operator.
func BindGenerateRangeArray(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("GENERATE_RANGE_ARRAY: invalid number of arguments: got %d, want between 2 and 3", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	r, ok := args[0].(*value.RangeValue)
	if !ok {
		return nil, fmt.Errorf("GENERATE_RANGE_ARRAY: first argument must be RANGE")
	}
	if r.Start == nil || r.End == nil {
		return nil, fmt.Errorf("GENERATE_RANGE_ARRAY: bounds must not be unbounded")
	}
	step := args[1]
	// `last_partial_range` defaults to TRUE — include the trailing
	// short bucket. When set to FALSE the caller asks us to drop any
	// final bucket that doesn't span a full `step`.
	includePartial := true
	if len(args) == 3 && args[2] != nil {
		b, err := args[2].ToBool()
		if err != nil {
			return nil, fmt.Errorf("GENERATE_RANGE_ARRAY: third argument must be BOOL: %w", err)
		}
		includePartial = b
	}
	out := []value.Value{}
	cur := r.Start
	// Cap the loop at a generous bound to avoid runaway loops on
	// pathologically small step values.
	const maxBuckets = 1 << 16
	for i := 0; i < maxBuckets; i++ {
		next, err := cur.Add(step)
		if err != nil {
			return nil, err
		}
		// `DateValue.Add(IntervalValue)` returns a DatetimeValue in the
		// general case (so e.g. `DATE + INTERVAL 5 HOUR` carries the
		// time component through). Inside a RANGE bucket loop that
		// drifts the result type away from the source RANGE's element
		// type, producing buckets whose start/end render with mixed
		// `2020-01-01` / `2020-01-01 00:00:00` shapes — and breaking
		// the strict typing the caller expects from
		// `ARRAY<RANGE<DATE>>`. Snap the computed bound back to the
		// element type recorded on the source range.
		next = coerceRangeBound(next, r.ElemHeader)
		// Trim the last bucket to the range end if it overshoots.
		gt, err := next.GT(r.End)
		if err != nil {
			return nil, err
		}
		if gt {
			if includePartial {
				out = append(out, &value.RangeValue{Start: cur, End: r.End, ElemHeader: r.ElemHeader})
			}
			break
		}
		out = append(out, &value.RangeValue{Start: cur, End: next, ElemHeader: r.ElemHeader})
		// Reached the boundary exactly — the bucket we just appended
		// is the last full one.
		eq, err := next.EQ(r.End)
		if err == nil && eq {
			break
		}
		cur = next
	}
	return &value.ArrayValue{Values: out}, nil
}

// coerceRangeBound projects a freshly-computed bucket boundary onto the
// RANGE's element type. Without this, repeated `DateValue.Add(interval)`
// calls yield DatetimeValues that the formatter renders with a
// trailing `00:00:00` clock for what is logically a DATE bucket.
func coerceRangeBound(v value.Value, elem value.ValueType) value.Value {
	if v == nil {
		return nil
	}
	switch elem {
	case value.DateValueType:
		if dt, ok := v.(value.DatetimeValue); ok {
			t := time.Time(dt)
			return value.DateValue(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()))
		}
	}
	return v
}

// BindRangeSessionize is the scalar form of the RANGE_SESSIONIZE TVF.
// At the analyzer level this is rewritten into a windowed scan; the
// scalar binding exists so that direct UDF resolution succeeds. It
// returns its first argument unchanged.
func BindRangeSessionize(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("RANGE_SESSIONIZE: invalid number of arguments: got %d, want at least 1", len(args))
	}
	return args[0], nil
}
