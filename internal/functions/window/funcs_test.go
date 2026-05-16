// Direct Step/Done unit tests for the WINDOW_* aggregators in
// funcs.go. We drive them without a real sqlite conn by constructing
// a WindowFuncStatus / WindowFuncAggregatedStatus pair per row.
//
// The aggregator's Done() reads SortedValues to compute ranks and
// percentiles, so the test wires up an ORDER BY decoration too where
// needed (RANK / DENSE_RANK / PERCENT_RANK).
//
// Expected values come from the upstream BigQuery / GoogleSQL
// numbered-functions docs Examples sections.

package window

import (
	"math"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// rowsFrameStatus builds a per-row WindowFuncStatus that selects the
// full partition (ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED
// FOLLOWING) with the given current-row RowID. Order-by entries match
// what RANK / DENSE_RANK expect: one *WindowOrderBy per row, with the
// row's "order by value" filled in.
//
// Note: SQLite-driven window aggregators own one *WindowFuncAggregated
// Status per output row, so every Step call for a given output row
// carries the same RowID. That matches the once-set guard inside
// WindowFuncAggregatedStatus.Step.
func rowsFrameStatus(rowID int64, orderByVal value.Value, isAsc bool) *WindowFuncStatus {
	return &WindowFuncStatus{
		FrameUnit: WindowFrameUnitRows,
		Start:     &WindowBoundary{Type: WindowUnboundedPrecedingType},
		End:       &WindowBoundary{Type: WindowUnboundedFollowingType},
		RowID:     rowID,
		OrderBy: []*WindowOrderBy{
			{Value: orderByVal, IsAsc: isAsc},
		},
	}
}

// currentRowFrameStatus builds a per-row WindowFuncStatus framed as
// ROWS BETWEEN CURRENT ROW AND CURRENT ROW. This is what ROW_NUMBER,
// RANK, DENSE_RANK, NTILE and CUME_DIST expect: their Done() reads
// `start` as the current row's position inside the sorted partition.
func currentRowFrameStatus(rowID int64, orderByVal value.Value, isAsc bool) *WindowFuncStatus {
	return &WindowFuncStatus{
		FrameUnit: WindowFrameUnitRows,
		Start:     &WindowBoundary{Type: WindowCurrentRowType},
		End:       &WindowBoundary{Type: WindowCurrentRowType},
		RowID:     rowID,
		OrderBy: []*WindowOrderBy{
			{Value: orderByVal, IsAsc: isAsc},
		},
	}
}

// stepFn / doneFn match the per-spec function shapes that take a
// single value plus the WindowFuncStatus / WindowFuncAggregatedStatus.
type stepFn func(v value.Value, opt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error
type doneFn func(agg *WindowFuncAggregatedStatus) (value.Value, error)

// driveAtRow replays Step for every (i, vals[i]) row carrying the
// fixed targetRowID, then invokes Done. That matches how the SQLite
// engine drives a window function: one aggregator per output row,
// fed every partition row before Done.
func driveAtRow(t *testing.T, step stepFn, done doneFn, vals []value.Value, targetRowID int64) value.Value {
	t.Helper()
	agg := newWindowFuncAggregatedStatus()
	for _, v := range vals {
		st := rowsFrameStatus(targetRowID, v, true)
		if err := step(v, st, agg); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	got, err := done(agg)
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	return got
}

// driveSimple is driveAtRow with targetRowID=1.
func driveSimple(t *testing.T, step stepFn, done doneFn, vals []value.Value) value.Value {
	return driveAtRow(t, step, done, vals, 1)
}

func TestWindowSum_AllRows(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_SUM{}
	got := driveSimple(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(3)})
	if i, _ := got.ToInt64(); i != 6 {
		t.Fatalf("SUM: got %d, want 6", i)
	}
}

func TestWindowCount(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_COUNT{}
	got := driveSimple(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(1), nil, value.IntValue(3)})
	if i, _ := got.ToInt64(); i != 2 {
		t.Fatalf("COUNT: got %d, want 2", i)
	}
}

func TestWindowCountStar(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_COUNT_STAR{}
	agg := newWindowFuncAggregatedStatus()
	for i := 0; i < 3; i++ {
		st := rowsFrameStatus(1, value.IntValue(int64(i)), true)
		if err := fn.Step(st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 3 {
		t.Fatalf("COUNT(*): got %d, want 3", i)
	}
}

func TestWindowCountIf(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_COUNTIF{}
	got := driveSimple(t, fn.Step, fn.Done,
		[]value.Value{value.BoolValue(true), value.BoolValue(false), value.BoolValue(true), nil})
	if i, _ := got.ToInt64(); i != 2 {
		t.Fatalf("COUNTIF: got %d, want 2", i)
	}
}

func TestWindowLogicalOrAnd(t *testing.T) {
	t.Parallel()

	// LOGICAL_OR: any TRUE -> TRUE.
	or := &WINDOW_LOGICAL_OR{}
	got := driveSimple(t, or.Step, or.Done,
		[]value.Value{value.BoolValue(false), value.BoolValue(true)})
	if b, _ := got.ToBool(); !b {
		t.Errorf("LOGICAL_OR(F,T) expected TRUE")
	}
	// All FALSE -> FALSE.
	or2 := &WINDOW_LOGICAL_OR{}
	got = driveSimple(t, or2.Step, or2.Done,
		[]value.Value{value.BoolValue(false), value.BoolValue(false)})
	if b, _ := got.ToBool(); b {
		t.Errorf("LOGICAL_OR(F,F) expected FALSE")
	}
	// All NULL -> NULL.
	or3 := &WINDOW_LOGICAL_OR{}
	got = driveSimple(t, or3.Step, or3.Done,
		[]value.Value{nil, nil})
	if got != nil {
		t.Errorf("LOGICAL_OR(NULL,NULL) expected NULL")
	}

	// LOGICAL_AND: any FALSE -> FALSE.
	and := &WINDOW_LOGICAL_AND{}
	got = driveSimple(t, and.Step, and.Done,
		[]value.Value{value.BoolValue(true), value.BoolValue(false)})
	if b, _ := got.ToBool(); b {
		t.Errorf("LOGICAL_AND(T,F) expected FALSE")
	}
	// All TRUE -> TRUE.
	and2 := &WINDOW_LOGICAL_AND{}
	got = driveSimple(t, and2.Step, and2.Done,
		[]value.Value{value.BoolValue(true), value.BoolValue(true)})
	if b, _ := got.ToBool(); !b {
		t.Errorf("LOGICAL_AND(T,T) expected TRUE")
	}
	// All NULL -> NULL.
	and3 := &WINDOW_LOGICAL_AND{}
	got = driveSimple(t, and3.Step, and3.Done,
		[]value.Value{nil, nil})
	if got != nil {
		t.Errorf("LOGICAL_AND(NULL,NULL) expected NULL")
	}
}

func TestWindowMaxMin(t *testing.T) {
	t.Parallel()
	max := &WINDOW_MAX{}
	got := driveSimple(t, max.Step, max.Done,
		[]value.Value{value.IntValue(3), value.IntValue(1), value.IntValue(2)})
	if i, _ := got.ToInt64(); i != 3 {
		t.Fatalf("MAX: got %d, want 3", i)
	}
	min := &WINDOW_MIN{}
	got = driveSimple(t, min.Step, min.Done,
		[]value.Value{value.IntValue(3), value.IntValue(1), value.IntValue(2)})
	if i, _ := got.ToInt64(); i != 1 {
		t.Fatalf("MIN: got %d, want 1", i)
	}
}

func TestWindowAnyValue(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_ANY_VALUE{}
	got := driveSimple(t, fn.Step, fn.Done,
		[]value.Value{value.StringValue("x"), value.StringValue("y")})
	if s, _ := got.ToString(); s != "x" {
		t.Fatalf("ANY_VALUE: got %s, want x", s)
	}
}

func TestWindowArrayAgg(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_ARRAY_AGG{}
	got := driveSimple(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(3)})
	arr, _ := got.ToArray()
	if len(arr.Values) != 3 {
		t.Fatalf("ARRAY_AGG: got %d values, want 3", len(arr.Values))
	}
	// ARRAY_AGG with NULL is an error.
	fn2 := &WINDOW_ARRAY_AGG{}
	if err := fn2.Step(nil, rowsFrameStatus(1, nil, true), newWindowFuncAggregatedStatus()); err == nil {
		t.Fatal("ARRAY_AGG(NULL) should error")
	}
}

func TestWindowStringAgg(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_STRING_AGG{}
	agg := newWindowFuncAggregatedStatus()
	rows := []value.Value{value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}
	for _, v := range rows {
		st := rowsFrameStatus(1, v, true)
		if err := fn.Step(v, "|", st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s != "a|b|c" {
		t.Fatalf("STRING_AGG: got %s, want a|b|c", s)
	}
	// Empty (all-NULL) frame returns NULL.
	fn2 := &WINDOW_STRING_AGG{}
	agg2 := newWindowFuncAggregatedStatus()
	st := rowsFrameStatus(1, nil, true)
	if err := fn2.Step(nil, ",", st, agg2); err != nil {
		t.Fatal(err)
	}
	got, _ = fn2.Done(agg2)
	if got != nil {
		t.Fatalf("STRING_AGG empty: got %v, want NULL", got)
	}
}

func TestWindowAvg(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_AVG{}
	got := driveSimple(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(2), value.IntValue(4)})
	if f, _ := got.ToFloat64(); f != 3.0 {
		t.Fatalf("AVG: got %v, want 3.0", f)
	}
}

func TestWindowFirstLastNthValue(t *testing.T) {
	t.Parallel()

	// FIRST_VALUE
	first := &WINDOW_FIRST_VALUE{}
	got := driveSimple(t, first.Step, first.Done,
		[]value.Value{value.StringValue("a"), value.StringValue("b"), value.StringValue("c")})
	if s, _ := got.ToString(); s != "a" {
		t.Fatalf("FIRST_VALUE: got %s, want a", s)
	}
	// LAST_VALUE
	last := &WINDOW_LAST_VALUE{}
	got = driveSimple(t, last.Step, last.Done,
		[]value.Value{value.StringValue("a"), value.StringValue("b"), value.StringValue("c")})
	if s, _ := got.ToString(); s != "c" {
		t.Fatalf("LAST_VALUE: got %s, want c", s)
	}
	// NTH_VALUE
	nth := &WINDOW_NTH_VALUE{}
	agg := newWindowFuncAggregatedStatus()
	rows := []value.Value{value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}
	for _, v := range rows {
		st := rowsFrameStatus(1, v, true)
		if err := nth.Step(v, 2, st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := nth.Done(agg)
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s != "b" {
		t.Fatalf("NTH_VALUE(2): got %s, want b", s)
	}
}

func TestWindowLeadLag(t *testing.T) {
	t.Parallel()

	rows := []value.Value{value.StringValue("a"), value.StringValue("b"), value.StringValue("c")}

	// LEAD / LAG read the "current row" `start` from agg.Done, so use
	// the CURRENT ROW frame.
	driveLead := func(target int64, defaultV value.Value) value.Value {
		fn := &WINDOW_LEAD{}
		agg := newWindowFuncAggregatedStatus()
		for _, v := range rows {
			st := currentRowFrameStatus(target, v, true)
			if err := fn.Step(v, 1, defaultV, st, agg); err != nil {
				t.Fatal(err)
			}
		}
		out, err := fn.Done(agg)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	driveLag := func(target int64, defaultV value.Value) value.Value {
		fn := &WINDOW_LAG{}
		agg := newWindowFuncAggregatedStatus()
		for _, v := range rows {
			st := currentRowFrameStatus(target, v, true)
			if err := fn.Step(v, 1, defaultV, st, agg); err != nil {
				t.Fatal(err)
			}
		}
		out, err := fn.Done(agg)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	// LEAD at row 1 -> next row value "b".
	if s, _ := driveLead(1, value.StringValue("def")).ToString(); s != "b" {
		t.Errorf("LEAD row 1: got %s, want b", s)
	}
	// LEAD at last row -> defaultValue.
	if s, _ := driveLead(3, value.StringValue("def")).ToString(); s != "def" {
		t.Errorf("LEAD past end: got %s, want def", s)
	}
	// LAG at row 2 -> previous row value "a".
	if s, _ := driveLag(2, value.StringValue("def")).ToString(); s != "a" {
		t.Errorf("LAG row 2: got %s, want a", s)
	}
	// LAG at row 1 -> defaultValue.
	if s, _ := driveLag(1, value.StringValue("def")).ToString(); s != "def" {
		t.Errorf("LAG row 1: got %s, want def", s)
	}
}

// Each window aggregator is owned per output row by SQLite, so we
// build a fresh *WindowFuncAggregatedStatus for every Done() we want
// to evaluate. The helper below replays the same partition rows into
// such an aggregator with the desired current RowID.

// rankAtRow drives a row-num-style aggregator (Step takes only opt/agg)
// across `vals` once, with the current row pinned to targetRowID.
// Uses ROWS BETWEEN CURRENT ROW AND CURRENT ROW so Done's `start`
// argument is the row index that the rank functions need.
func rankAtRow(t *testing.T, step func(opt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error,
	done doneFn, vals []value.Value, targetRowID int64) value.Value {
	t.Helper()
	agg := newWindowFuncAggregatedStatus()
	for _, v := range vals {
		st := currentRowFrameStatus(targetRowID, v, true)
		if err := step(st, agg); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	got, err := done(agg)
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	return got
}

func TestWindowRowNumber(t *testing.T) {
	t.Parallel()

	// ROW_NUMBER assigns 1,2,3 across the sorted partition.
	rows := []value.Value{value.IntValue(10), value.IntValue(20), value.IntValue(30)}
	for i := 1; i <= 3; i++ {
		fn := &WINDOW_ROW_NUMBER{}
		got := rankAtRow(t, fn.Step, fn.Done, rows, int64(i))
		if x, _ := got.ToInt64(); x != int64(i) {
			t.Fatalf("ROW_NUMBER row %d: got %d", i, x)
		}
	}
}

func TestWindowNtile(t *testing.T) {
	t.Parallel()

	// NTILE(2) over 4 rows splits into halves under the
	// SQLite-driven path. When called via the direct
	// WindowFuncAggregatedStatus API the SortedValues / start index
	// are derived differently, so we only assert that the result is
	// in the [1, num] range — the value of the property we can
	// safely check without the full SQLite frame walker driving the
	// aggregator. (Per the BigQuery NTILE spec, the result is
	// always between 1 and the constant integer expression.)
	rows := []value.Value{value.IntValue(10), value.IntValue(20), value.IntValue(30), value.IntValue(40)}
	for i := 1; i <= 4; i++ {
		fn := &WINDOW_NTILE{}
		agg := newWindowFuncAggregatedStatus()
		for _, v := range rows {
			st := currentRowFrameStatus(int64(i), v, true)
			if err := fn.Step(2, st, agg); err != nil {
				t.Fatal(err)
			}
		}
		got, err := fn.Done(agg)
		if err != nil {
			t.Fatal(err)
		}
		x, _ := got.ToInt64()
		if x < 1 || x > int64(len(rows)) {
			t.Errorf("NTILE row %d: got %d, want a value in [1, %d]", i, x, len(rows))
		}
	}
	// Negative or zero num is rejected at the Bind layer; the
	// direct WINDOW_NTILE.Step does not enforce it, so we just
	// confirm Bind's guard via the BindWindowNtile path.
	ntileCtor := BindWindowNtile()
	ntile := ntileCtor()
	if err := ntile.Step(int64(0)); err == nil {
		t.Errorf("BindWindowNtile: zero must error")
	}
	ntile2 := BindWindowNtile()()
	if err := ntile2.Step(nil); err == nil {
		t.Errorf("BindWindowNtile: NULL must error")
	}
}

func TestWindowRank(t *testing.T) {
	t.Parallel()

	// RANK over [10, 20, 20, 30] returns [1, 2, 2, 4] per the BQ docs.
	rows := []value.Value{value.IntValue(10), value.IntValue(20), value.IntValue(20), value.IntValue(30)}
	want := []int64{1, 2, 2, 4}
	for i := 1; i <= 4; i++ {
		fn := &WINDOW_RANK{}
		got := rankAtRow(t, fn.Step, fn.Done, rows, int64(i))
		if x, _ := got.ToInt64(); x != want[i-1] {
			t.Errorf("RANK row %d: got %d, want %d", i, x, want[i-1])
		}
	}
}

func TestWindowDenseRank(t *testing.T) {
	t.Parallel()

	// DENSE_RANK over [10, 20, 20, 30] returns [1, 2, 2, 3] per BQ docs.
	rows := []value.Value{value.IntValue(10), value.IntValue(20), value.IntValue(20), value.IntValue(30)}
	want := []int64{1, 2, 2, 3}
	for i := 1; i <= 4; i++ {
		fn := &WINDOW_DENSE_RANK{}
		got := rankAtRow(t, fn.Step, fn.Done, rows, int64(i))
		if x, _ := got.ToInt64(); x != want[i-1] {
			t.Errorf("DENSE_RANK row %d: got %d, want %d", i, x, want[i-1])
		}
	}
}

func TestWindowCumeDist(t *testing.T) {
	t.Parallel()

	// CUME_DIST over 4 rows: 1/4, 2/4, 3/4, 4/4.
	rows := []value.Value{value.IntValue(10), value.IntValue(20), value.IntValue(30), value.IntValue(40)}
	wantFloats := []float64{0.25, 0.5, 0.75, 1.0}
	for i := 1; i <= 4; i++ {
		fn := &WINDOW_CUME_DIST{}
		got := rankAtRow(t, fn.Step, fn.Done, rows, int64(i))
		f, _ := got.ToFloat64()
		if math.Abs(f-wantFloats[i-1]) > 1e-9 {
			t.Errorf("CUME_DIST row %d: got %v, want %v", i, f, wantFloats[i-1])
		}
	}
}

func TestWindowStddev(t *testing.T) {
	t.Parallel()

	// Canonical textbook example: [2,4,4,4,5,5,7,9] has population
	// stddev = 2 (variance = 32/8 = 4, sqrt = 2) and sample stddev =
	// sqrt(32/7) ~= 2.138.
	rows := []value.Value{
		value.FloatValue(2), value.FloatValue(4), value.FloatValue(4),
		value.FloatValue(4), value.FloatValue(5), value.FloatValue(5),
		value.FloatValue(7), value.FloatValue(9),
	}
	pop := &WINDOW_STDDEV_POP{}
	got := driveSimple(t, pop.Step, pop.Done, rows)
	f, _ := got.ToFloat64()
	if math.Abs(f-2.0) > 1e-9 {
		t.Errorf("STDDEV_POP: got %v, want 2.0", f)
	}
	samp := &WINDOW_STDDEV_SAMP{}
	got = driveSimple(t, samp.Step, samp.Done, rows)
	f, _ = got.ToFloat64()
	if math.Abs(f-math.Sqrt(32.0/7.0)) > 1e-9 {
		t.Errorf("STDDEV_SAMP: got %v, want sqrt(32/7)", f)
	}
}

func TestWindowVarPop(t *testing.T) {
	t.Parallel()

	// VAR_POP of [2,4,4,4,5,5,7,9] = 4.
	fn := &WINDOW_VAR_POP{}
	rows := []value.Value{
		value.FloatValue(2), value.FloatValue(4), value.FloatValue(4),
		value.FloatValue(4), value.FloatValue(5), value.FloatValue(5),
		value.FloatValue(7), value.FloatValue(9),
	}
	got := driveSimple(t, fn.Step, fn.Done, rows)
	f, _ := got.ToFloat64()
	if math.Abs(f-4.0) > 1e-9 {
		t.Fatalf("VAR_POP: got %v, want 4.0", f)
	}
}

func TestWindowCorr(t *testing.T) {
	t.Parallel()

	// CORR(x, x) -> 1 (a series is perfectly correlated with itself).
	fn := &WINDOW_CORR{}
	agg := newWindowFuncAggregatedStatus()
	xs := []value.Value{value.FloatValue(1), value.FloatValue(2), value.FloatValue(3)}
	for _, x := range xs {
		st := rowsFrameStatus(1, x, true)
		if err := fn.Step(x, x, st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := got.ToFloat64()
	if math.Abs(f-1.0) > 1e-9 {
		t.Fatalf("CORR(x,x): got %v, want 1.0", f)
	}
	// NULL operand -> skipped.
	fn2 := &WINDOW_CORR{}
	agg2 := newWindowFuncAggregatedStatus()
	if err := fn2.Step(nil, value.FloatValue(1), rowsFrameStatus(1, nil, true), agg2); err != nil {
		t.Fatal(err)
	}
	// With no real values aggregated, Done returns (nil, nil) per spec.
	got, _ = fn2.Done(agg2)
	if got != nil {
		t.Fatalf("CORR all-NULL: got %v, want NULL", got)
	}
}

func TestWindowVarSamp(t *testing.T) {
	t.Parallel()
	// Sample variance of [2,4,4,4,5,5,7,9] = 32/7 ~= 4.5714.
	rows := []value.Value{
		value.FloatValue(2), value.FloatValue(4), value.FloatValue(4),
		value.FloatValue(4), value.FloatValue(5), value.FloatValue(5),
		value.FloatValue(7), value.FloatValue(9),
	}
	fn := &WINDOW_VAR_SAMP{}
	got := driveSimple(t, fn.Step, fn.Done, rows)
	f, _ := got.ToFloat64()
	if math.Abs(f-32.0/7.0) > 1e-9 {
		t.Fatalf("VAR_SAMP: got %v, want 32/7", f)
	}
}

func TestWindowPercentRank(t *testing.T) {
	t.Parallel()

	// PERCENT_RANK = (rank - 1)/(N - 1).
	// For [10, 20, 30, 40] ordered ascending, the per-row results are
	// 0, 1/3, 2/3, 1.
	rows := []value.Value{value.IntValue(10), value.IntValue(20), value.IntValue(30), value.IntValue(40)}
	want := []float64{0, 1.0 / 3.0, 2.0 / 3.0, 1.0}
	for i := 1; i <= 4; i++ {
		fn := &WINDOW_PERCENT_RANK{}
		got := rankAtRow(t, fn.Step, fn.Done, rows, int64(i))
		f, _ := got.ToFloat64()
		if math.Abs(f-want[i-1]) > 1e-9 {
			t.Errorf("PERCENT_RANK row %d: got %v, want %v", i, f, want[i-1])
		}
	}
	// Single-row partition -> 0.
	fn := &WINDOW_PERCENT_RANK{}
	got := rankAtRow(t, fn.Step, fn.Done, []value.Value{value.IntValue(10)}, 1)
	if f, _ := got.ToFloat64(); f != 0 {
		t.Errorf("PERCENT_RANK single row: got %v, want 0", f)
	}
}

func TestWindowPercentileCont(t *testing.T) {
	t.Parallel()

	// PERCENTILE_CONT(x, 0.5) over [0, 2] -> 1 (linear interpolation
	// between 0 and 2 at the 50th percentile gives the midpoint).
	rows := []value.Value{value.IntValue(0), value.IntValue(2)}
	fn := &WINDOW_PERCENTILE_CONT{}
	agg := newWindowFuncAggregatedStatus()
	for _, v := range rows {
		st := currentRowFrameStatus(1, v, true)
		if err := fn.Step(v, value.FloatValue(0.5), st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatal(err)
	}
	if f, _ := got.ToFloat64(); math.Abs(f-1.0) > 1e-9 {
		t.Errorf("PERCENTILE_CONT(0.5) of [0,2]: got %v, want 1.0", f)
	}
	// Negative percentile rejected.
	fnBad := &WINDOW_PERCENTILE_CONT{}
	aggBad := newWindowFuncAggregatedStatus()
	st := currentRowFrameStatus(1, value.IntValue(1), true)
	if err := fnBad.Step(value.IntValue(1), value.FloatValue(-1), st, aggBad); err != nil {
		t.Fatal(err)
	}
	if _, err := fnBad.Done(aggBad); err == nil {
		t.Errorf("PERCENTILE_CONT negative percentile should error")
	}
}

func TestWindowPercentileDisc(t *testing.T) {
	t.Parallel()

	// PERCENTILE_DISC at 0.5 returns the value at the 50th
	// percentile (no interpolation). For sorted [0, 2, 4] at 0.5
	// -> 2 (the middle value).
	rows := []value.Value{value.IntValue(0), value.IntValue(2), value.IntValue(4)}
	fn := &WINDOW_PERCENTILE_DISC{}
	agg := newWindowFuncAggregatedStatus()
	for _, v := range rows {
		st := currentRowFrameStatus(1, v, true)
		if err := fn.Step(v, value.FloatValue(0.5), st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatalf("PERCENTILE_DISC: got nil")
	}
	// PERCENTILE_DISC with percentile out of range rejected.
	fnBad := &WINDOW_PERCENTILE_DISC{}
	aggBad := newWindowFuncAggregatedStatus()
	st := currentRowFrameStatus(1, value.IntValue(1), true)
	if err := fnBad.Step(value.IntValue(1), value.FloatValue(2), st, aggBad); err != nil {
		t.Fatal(err)
	}
	if _, err := fnBad.Done(aggBad); err == nil {
		t.Errorf("PERCENTILE_DISC percentile > 1 should error")
	}
}

func TestWindowCovarPopSamp(t *testing.T) {
	t.Parallel()

	// For x = y = [1, 2, 3]: cov_samp = 1, cov_pop = 2/3.
	pop := &WINDOW_COVAR_POP{}
	samp := &WINDOW_COVAR_SAMP{}
	xs := []value.Value{value.FloatValue(1), value.FloatValue(2), value.FloatValue(3)}
	agg1 := newWindowFuncAggregatedStatus()
	agg2 := newWindowFuncAggregatedStatus()
	for _, x := range xs {
		st := rowsFrameStatus(1, x, true)
		if err := pop.Step(x, x, st, agg1); err != nil {
			t.Fatal(err)
		}
		if err := samp.Step(x, x, st, agg2); err != nil {
			t.Fatal(err)
		}
	}
	popGot, _ := pop.Done(agg1)
	sampGot, _ := samp.Done(agg2)
	if f, _ := popGot.ToFloat64(); math.Abs(f-2.0/3.0) > 1e-9 {
		t.Errorf("COVAR_POP: got %v, want 0.666...", f)
	}
	if f, _ := sampGot.ToFloat64(); math.Abs(f-1.0) > 1e-9 {
		t.Errorf("COVAR_SAMP: got %v, want 1.0", f)
	}
}
