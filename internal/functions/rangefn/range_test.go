package rangefn

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func mkDate(y int, m time.Month, d int) value.Value {
	return value.DateValue(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func mustInterval(t *testing.T, s string) value.Value {
	t.Helper()
	iv, err := value.ParseInterval(s)
	if err != nil {
		t.Fatalf("ParseInterval(%q): %v", s, err)
	}
	return iv
}

// --- BindRange / accessors ---

func TestBindRange(t *testing.T) {
	got, err := BindRange(mkDate(2024, 1, 1), mkDate(2024, 1, 5))
	if err != nil {
		t.Fatalf("BindRange: %v", err)
	}
	r, ok := got.(*value.RangeValue)
	if !ok {
		t.Fatalf("want *RangeValue, got %T", got)
	}
	if r.ElemHeader != value.DateValueType {
		t.Fatalf("want DateValueType element header, got %v", r.ElemHeader)
	}
}

func TestBindRangeUnboundedEnd(t *testing.T) {
	got, err := BindRange(mkDate(2024, 1, 1), nil)
	if err != nil {
		t.Fatalf("BindRange: %v", err)
	}
	r := got.(*value.RangeValue)
	if r.End != nil {
		t.Fatalf("want unbounded end")
	}
}

func TestBindRangeBothNull(t *testing.T) {
	if _, err := BindRange(nil, nil); err == nil {
		t.Fatalf("both bounds NULL should error")
	}
}

func TestBindRangeArity(t *testing.T) {
	if _, err := BindRange(mkDate(2024, 1, 1)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindRangeStart(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	got, err := BindRangeStart(r)
	if err != nil {
		t.Fatalf("BindRangeStart: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Day() != 1 {
		t.Fatalf("want day 1, got %d", tt.Day())
	}

	got, _ = BindRangeStart(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}

	if _, err := BindRangeStart(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindRangeStart(value.IntValue(1)); err == nil {
		t.Fatalf("non-range arg should error")
	}
}

func TestBindRangeEnd(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	got, _ := BindRangeEnd(r)
	tt, _ := got.ToTime()
	if tt.Day() != 5 {
		t.Fatalf("want day 5, got %d", tt.Day())
	}

	got, _ = BindRangeEnd(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindRangeEnd(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindRangeEnd(value.IntValue(1)); err == nil {
		t.Fatalf("non-range arg should error")
	}
}

func TestBindRangeIsStartUnbounded(t *testing.T) {
	r := &value.RangeValue{Start: nil, End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	got, _ := BindRangeIsStartUnbounded(r)
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want TRUE for unbounded start")
	}

	r2 := &value.RangeValue{Start: mkDate(2024, 1, 1), End: nil, ElemHeader: value.DateValueType}
	got, _ = BindRangeIsStartUnbounded(r2)
	b, _ = got.ToBool()
	if b {
		t.Fatalf("want FALSE for bounded start")
	}

	got, _ = BindRangeIsStartUnbounded(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindRangeIsStartUnbounded(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindRangeIsStartUnbounded(value.IntValue(1)); err == nil {
		t.Fatalf("non-range arg should error")
	}
}

func TestBindRangeIsEndUnbounded(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: nil, ElemHeader: value.DateValueType}
	got, _ := BindRangeIsEndUnbounded(r)
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want TRUE for unbounded end")
	}

	r2 := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	got, _ = BindRangeIsEndUnbounded(r2)
	b, _ = got.ToBool()
	if b {
		t.Fatalf("want FALSE for bounded end")
	}

	got, _ = BindRangeIsEndUnbounded(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindRangeIsEndUnbounded(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindRangeIsEndUnbounded(value.IntValue(1)); err == nil {
		t.Fatalf("non-range arg should error")
	}
}

// --- BindRangeContains ---

func TestBindRangeContainsRange(t *testing.T) {
	outer := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 12, 31), ElemHeader: value.DateValueType}
	inner := &value.RangeValue{Start: mkDate(2024, 3, 1), End: mkDate(2024, 4, 1), ElemHeader: value.DateValueType}

	got, err := BindRangeContains(outer, inner)
	if err != nil {
		t.Fatalf("BindRangeContains: %v", err)
	}
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("inner range should be contained")
	}

	// Inner that extends before outer.start.
	innerBefore := &value.RangeValue{Start: mkDate(2023, 12, 1), End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	got, _ = BindRangeContains(outer, innerBefore)
	b, _ = got.ToBool()
	if b {
		t.Fatalf("inner extending before outer must be FALSE")
	}

	// Inner that extends past outer.end.
	innerAfter := &value.RangeValue{Start: mkDate(2024, 12, 25), End: mkDate(2025, 1, 5), ElemHeader: value.DateValueType}
	got, _ = BindRangeContains(outer, innerAfter)
	b, _ = got.ToBool()
	if b {
		t.Fatalf("inner extending past outer must be FALSE")
	}
}

func TestBindRangeContainsPoint(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 12, 31), ElemHeader: value.DateValueType}
	got, _ := BindRangeContains(r, mkDate(2024, 6, 1))
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("point should be contained")
	}

	got, _ = BindRangeContains(r, mkDate(2023, 12, 1))
	b, _ = got.ToBool()
	if b {
		t.Fatalf("point before start must be FALSE")
	}

	got, _ = BindRangeContains(r, mkDate(2025, 1, 1))
	b, _ = got.ToBool()
	if b {
		t.Fatalf("point after end must be FALSE")
	}
}

func TestBindRangeContainsNullAndErrors(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 12, 31), ElemHeader: value.DateValueType}
	got, _ := BindRangeContains(nil, r)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindRangeContains(value.IntValue(1), r); err == nil {
		t.Fatalf("first arg non-RANGE should error")
	}
	if _, err := BindRangeContains(r); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindRangeOverlaps ---

func TestBindRangeOverlaps(t *testing.T) {
	a := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 6, 1), ElemHeader: value.DateValueType}
	b := &value.RangeValue{Start: mkDate(2024, 5, 1), End: mkDate(2024, 12, 1), ElemHeader: value.DateValueType}
	got, _ := BindRangeOverlaps(a, b)
	v, _ := got.ToBool()
	if !v {
		t.Fatalf("overlapping ranges want TRUE")
	}

	c := &value.RangeValue{Start: mkDate(2024, 7, 1), End: mkDate(2024, 12, 1), ElemHeader: value.DateValueType}
	got, _ = BindRangeOverlaps(a, c)
	v, _ = got.ToBool()
	if v {
		t.Fatalf("disjoint ranges want FALSE")
	}
}

func TestBindRangeOverlapsNullAndErrors(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 12, 31), ElemHeader: value.DateValueType}
	got, _ := BindRangeOverlaps(nil, r)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindRangeOverlaps(value.IntValue(1), r); err == nil {
		t.Fatalf("non-RANGE arg should error")
	}
	if _, err := BindRangeOverlaps(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindRangeIntersect ---

func TestBindRangeIntersect(t *testing.T) {
	a := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 6, 1), ElemHeader: value.DateValueType}
	b := &value.RangeValue{Start: mkDate(2024, 5, 1), End: mkDate(2024, 12, 1), ElemHeader: value.DateValueType}
	got, err := BindRangeIntersect(a, b)
	if err != nil {
		t.Fatalf("BindRangeIntersect: %v", err)
	}
	r, ok := got.(*value.RangeValue)
	if !ok {
		t.Fatalf("want RangeValue, got %T", got)
	}
	ts := time.Time(r.Start.(value.DateValue))
	te := time.Time(r.End.(value.DateValue))
	if ts.Day() != 1 || ts.Month() != 5 {
		t.Fatalf("want start 2024-05-01, got %v", ts)
	}
	if te.Day() != 1 || te.Month() != 6 {
		t.Fatalf("want end 2024-06-01, got %v", te)
	}
}

func TestBindRangeIntersectDisjoint(t *testing.T) {
	a := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 3, 1), ElemHeader: value.DateValueType}
	b := &value.RangeValue{Start: mkDate(2024, 6, 1), End: mkDate(2024, 12, 1), ElemHeader: value.DateValueType}
	got, _ := BindRangeIntersect(a, b)
	if got != nil {
		t.Fatalf("disjoint ranges should yield NULL, got %v", got)
	}
}

func TestBindRangeIntersectNullAndErrors(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 12, 31), ElemHeader: value.DateValueType}
	got, _ := BindRangeIntersect(nil, r)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindRangeIntersect(value.IntValue(1), r); err == nil {
		t.Fatalf("non-RANGE arg should error")
	}
	if _, err := BindRangeIntersect(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindGenerateRangeArray ---

func TestBindGenerateRangeArray(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	step := mustInterval(t, "0-0 1 0:0:0") // 1 day
	got, err := BindGenerateRangeArray(r, step)
	if err != nil {
		t.Fatalf("BindGenerateRangeArray: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 4 {
		t.Fatalf("want 4 buckets (Jan 1->5 by 1d), got %d", len(arr.Values))
	}
	// Each bucket should be a RangeValue with DateValue bounds.
	for i, e := range arr.Values {
		rb := e.(*value.RangeValue)
		if _, ok := rb.Start.(value.DateValue); !ok {
			t.Fatalf("bucket %d start is %T, want DateValue", i, rb.Start)
		}
	}
}

func TestBindGenerateRangeArrayLastPartial(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 1, 4), ElemHeader: value.DateValueType}
	step := mustInterval(t, "0-0 2 0:0:0") // 2 days
	// With default last_partial_range=TRUE expect 2 buckets (Jan1-3, Jan3-4).
	got, _ := BindGenerateRangeArray(r, step)
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 buckets (last partial), got %d", len(arr.Values))
	}

	// With last_partial_range=FALSE expect 1 bucket only.
	got, _ = BindGenerateRangeArray(r, step, value.BoolValue(false))
	arr, _ = got.ToArray()
	if len(arr.Values) != 1 {
		t.Fatalf("want 1 bucket (no partial), got %d", len(arr.Values))
	}
}

func TestBindGenerateRangeArrayUnboundedError(t *testing.T) {
	r := &value.RangeValue{Start: nil, End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	step := mustInterval(t, "0-0 1 0:0:0")
	if _, err := BindGenerateRangeArray(r, step); err == nil {
		t.Fatalf("unbounded start should error")
	}
}

func TestBindGenerateRangeArrayNullAndErrors(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	step := mustInterval(t, "0-0 1 0:0:0")
	got, _ := BindGenerateRangeArray(nil, step)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindGenerateRangeArray(value.IntValue(1), step); err == nil {
		t.Fatalf("first arg non-RANGE should error")
	}
	if _, err := BindGenerateRangeArray(r); err == nil {
		t.Fatalf("arity error expected (1)")
	}
	if _, err := BindGenerateRangeArray(r, step, value.IntValue(1), value.IntValue(2)); err == nil {
		t.Fatalf("arity error expected (4)")
	}
}

// --- BindRangeSessionize ---

func TestBindRangeSessionizePassthrough(t *testing.T) {
	r := &value.RangeValue{Start: mkDate(2024, 1, 1), End: mkDate(2024, 1, 5), ElemHeader: value.DateValueType}
	got, err := BindRangeSessionize(r)
	if err != nil {
		t.Fatalf("BindRangeSessionize: %v", err)
	}
	if got != r {
		t.Fatalf("expected first arg returned unchanged")
	}
	if _, err := BindRangeSessionize(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- coerceRangeBound / bound helpers ---

// TestRangeBoundLEQ exercises both NULL-on-start and NULL-on-end
// branches plus the typed comparison path. These helpers are
// otherwise covered only through the Bind* surface above.
func TestRangeBoundLEQ(t *testing.T) {
	if !rangeBoundLEQ(nil, mkDate(2024, 1, 1), true) {
		t.Fatalf("NULL on start side must be smallest")
	}
	if rangeBoundLEQ(mkDate(2024, 1, 1), nil, true) {
		t.Fatalf("non-NULL ≤ NULL on start side must be FALSE")
	}
	if rangeBoundLEQ(nil, mkDate(2024, 1, 1), false) {
		t.Fatalf("NULL on end side must NOT be ≤ a finite point")
	}
	if !rangeBoundLEQ(mkDate(2024, 1, 1), mkDate(2024, 6, 1), true) {
		t.Fatalf("Jan ≤ Jun must be TRUE")
	}
}

func TestPointLT(t *testing.T) {
	lt, err := pointLT(nil, mkDate(2024, 1, 1))
	if err != nil || lt {
		t.Fatalf("NULL LT date must be FALSE")
	}
	lt, _ = pointLT(mkDate(2024, 1, 1), mkDate(2024, 6, 1))
	if !lt {
		t.Fatalf("Jan LT Jun must be TRUE")
	}
}

func TestMaxMinBound(t *testing.T) {
	a := mkDate(2024, 1, 1)
	b := mkDate(2024, 6, 1)
	if maxBound(a, b, true) != b {
		t.Fatalf("max(a,b) want b")
	}
	if minBound(a, b, true) != a {
		t.Fatalf("min(a,b) want a")
	}
	// NULL handling: NULL on start side means -inf for max → returns
	// non-NULL.
	if maxBound(nil, b, true) != b {
		t.Fatalf("max(nil,b,start) want b")
	}
	if minBound(nil, b, true) != nil {
		t.Fatalf("min(nil,b,start) want nil (NULL is -inf)")
	}
}

func TestCoerceRangeBound(t *testing.T) {
	// DatetimeValue projected onto DateValueType becomes a DateValue at midnight.
	dt := value.DatetimeValue(time.Date(2024, 1, 2, 5, 0, 0, 0, time.UTC))
	got := coerceRangeBound(dt, value.DateValueType)
	if _, ok := got.(value.DateValue); !ok {
		t.Fatalf("want DateValue, got %T", got)
	}
	// Non-Date element type leaves the value unchanged.
	got = coerceRangeBound(dt, value.TimestampValueType)
	if _, ok := got.(value.DatetimeValue); !ok {
		t.Fatalf("non-Date elem should pass through DatetimeValue, got %T", got)
	}
	// NULL stays NULL.
	if coerceRangeBound(nil, value.DateValueType) != nil {
		t.Fatalf("NULL bound must stay NULL")
	}
}
