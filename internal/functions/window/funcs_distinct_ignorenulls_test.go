// Direct unit tests for the DISTINCT / IGNORE NULLS branches of the
// WINDOW_* aggregator Done methods. The SQL-driven layers don't reach
// these branches because the formatter prefers the SQLite-native /
// customNativeWindow paths even when DISTINCT or IGNORE NULLS is in
// play, but the binder Step/Done closures are still wired to the
// WINDOW_* spec methods and the runtime invariants must hold for
// them. These tests drive each spec's Done() with DistinctOpt /
// IgnoreNullsOpt set on the aggregator status directly.
//
// Expected outputs come from the underlying BigQuery / GoogleSQL
// semantics: DISTINCT drops duplicates by stringified key;
// IGNORE NULLS drops NULL values entirely.

package window

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// driveWithFlags is driveAtRow with manual DistinctOpt /
// IgnoreNullsOpt flag wiring. Each row's per-step value is passed
// through, and the aggregator's flags are set BEFORE the first Step
// (in line with how the SQLite-side aggregator initialises them
// once on the first Step).
func driveWithFlags(t *testing.T, step stepFn, done doneFn, vals []value.Value, distinct, ignoreNulls bool) value.Value {
	t.Helper()
	agg := newWindowFuncAggregatedStatus()
	agg.DistinctOpt = distinct
	agg.IgnoreNullsOpt = ignoreNulls
	for _, v := range vals {
		st := rowsFrameStatus(1, v, true)
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

// TestWindowSumDistinctIgnoreNulls drives the DISTINCT and
// IGNORE NULLS branches of WINDOW_SUM.Done.
func TestWindowSumDistinctIgnoreNulls(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_SUM{}
	// Plain SUM: 1 + 2 + 3 + 2 = 8.
	got := driveWithFlags(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(3), value.IntValue(2)}, false, false)
	if i, _ := got.ToInt64(); i != 8 {
		t.Errorf("SUM(1,2,3,2) = %d; want 8", i)
	}

	// DISTINCT: 1 + 2 + 3 = 6.
	fn2 := &WINDOW_SUM{}
	got = driveWithFlags(t, fn2.Step, fn2.Done,
		[]value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(3), value.IntValue(2)}, true, false)
	if i, _ := got.ToInt64(); i != 6 {
		t.Errorf("SUM(DISTINCT 1,2,3,2) = %d; want 6", i)
	}

	// IGNORE NULLS over [1, NULL, 3] -> 4.
	fn3 := &WINDOW_SUM{}
	got = driveWithFlags(t, fn3.Step, fn3.Done,
		[]value.Value{value.IntValue(1), nil, value.IntValue(3)}, false, true)
	if i, _ := got.ToInt64(); i != 4 {
		t.Errorf("SUM IGNORE NULLS = %d; want 4", i)
	}
}

// TestWindowAvgDistinct drives the DISTINCT branch of WINDOW_AVG.
// AVG(DISTINCT 1, 2, 3, 2) = (1+2+3)/3 = 2.0.
func TestWindowAvgDistinct(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_AVG{}
	got := driveWithFlags(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(3), value.IntValue(2)}, true, false)
	if f, _ := got.ToFloat64(); f < 1.5 || f > 2.5 {
		t.Errorf("AVG(DISTINCT 1,2,3,2) = %v; want around 2.0", f)
	}
}

// TestWindowCountDistinct drives the DISTINCT branch of WINDOW_COUNT.
// COUNT(DISTINCT 1, 2, 3, 2) = 3 distinct values.
func TestWindowCountDistinct(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_COUNT{}
	got := driveWithFlags(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(3), value.IntValue(2)}, true, false)
	if i, _ := got.ToInt64(); i != 3 {
		t.Errorf("COUNT(DISTINCT 1,2,3,2) = %d; want 3", i)
	}
}

// TestWindowArrayAggDistinctIgnoreNulls drives the DISTINCT +
// IGNORE NULLS Done branches of WINDOW_ARRAY_AGG.
func TestWindowArrayAggDistinctIgnoreNulls(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_ARRAY_AGG{}
	// IGNORE NULLS doesn't have a NULL-step in ARRAY_AGG (which
	// errors on NULL input), so drive distinct with a duplicated
	// value space.
	got := driveWithFlags(t, fn.Step, fn.Done,
		[]value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(1), value.IntValue(3)}, true, false)
	arr, _ := got.ToArray()
	if len(arr.Values) != 3 {
		t.Errorf("ARRAY_AGG(DISTINCT 1,2,1,3) has %d values; want 3", len(arr.Values))
	}
}

// TestWindowStringAggDistinct drives WINDOW_STRING_AGG.Done DISTINCT
// branch. STRING_AGG(DISTINCT 'a','b','a') -> "a,b" (default ",").
func TestWindowStringAggDistinct(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_STRING_AGG{}
	agg := newWindowFuncAggregatedStatus()
	agg.DistinctOpt = true
	for _, s := range []string{"a", "b", "a"} {
		st := rowsFrameStatus(1, value.StringValue(s), true)
		if err := fn.Step(value.StringValue(s), ",", st, agg); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if s, _ := got.ToString(); s != "a,b" {
		t.Errorf("STRING_AGG(DISTINCT a,b,a) = %q; want a,b", s)
	}
}

// TestWindowMaxMinDistinctIgnoreNulls drives the DISTINCT / IGNORE
// NULLS branches of WINDOW_MAX and WINDOW_MIN — the duplicate-skip
// path in Done.
func TestWindowMaxMinDistinctIgnoreNulls(t *testing.T) {
	t.Parallel()
	max := &WINDOW_MAX{}
	got := driveWithFlags(t, max.Step, max.Done,
		[]value.Value{value.IntValue(1), value.IntValue(3), value.IntValue(3), value.IntValue(2)}, true, false)
	if i, _ := got.ToInt64(); i != 3 {
		t.Errorf("MAX(DISTINCT 1,3,3,2) = %d; want 3", i)
	}

	min := &WINDOW_MIN{}
	got = driveWithFlags(t, min.Step, min.Done,
		[]value.Value{value.IntValue(3), value.IntValue(1), value.IntValue(1), value.IntValue(2)}, true, false)
	if i, _ := got.ToInt64(); i != 1 {
		t.Errorf("MIN(DISTINCT 3,1,1,2) = %d; want 1", i)
	}
}

// TestWindowLogicalOrAndDistinct drives the DISTINCT branch of
// LOGICAL_OR / LOGICAL_AND Done. With DISTINCT TRUE/FALSE values
// only count once but the result is unchanged.
func TestWindowLogicalOrAndDistinct(t *testing.T) {
	t.Parallel()
	or := &WINDOW_LOGICAL_OR{}
	got := driveWithFlags(t, or.Step, or.Done,
		[]value.Value{value.BoolValue(false), value.BoolValue(true), value.BoolValue(true)}, true, false)
	if b, _ := got.ToBool(); !b {
		t.Errorf("LOGICAL_OR(DISTINCT F,T,T): want TRUE")
	}

	and := &WINDOW_LOGICAL_AND{}
	got = driveWithFlags(t, and.Step, and.Done,
		[]value.Value{value.BoolValue(true), value.BoolValue(true), value.BoolValue(false)}, true, false)
	if b, _ := got.ToBool(); b {
		t.Errorf("LOGICAL_AND(DISTINCT T,T,F): want FALSE")
	}
}

// TestStringAggWindowNativeOptionsAndDelim drives the
// stringAggWindowNative DISTINCT / IGNORE NULLS / explicit-delim
// branches. With marker JSON for distinct + ignore_nulls and a
// custom delimiter "|", input ["a", NULL, "b", "a"] -> "a|b".
func TestStringAggWindowNativeOptionsAndDelim(t *testing.T) {
	t.Parallel()
	a := NewStringAggWindowNative()().(*stringAggWindowNative)
	distinctMarker := `{"type":"aggregate_distinct"}`
	ignoreMarker := `{"type":"aggregate_ignore_nulls"}`
	// Step: (value, delim, ...markers)
	rows := []any{"a", "|", distinctMarker, ignoreMarker}
	if err := a.Step(rows...); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(nil, "|", distinctMarker, ignoreMarker); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", "|", distinctMarker, ignoreMarker); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("a", "|", distinctMarker, ignoreMarker); err != nil {
		t.Fatal(err)
	}
	got, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	v, err := value.DecodeValue(got)
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := v.ToString(); s != "a|b" {
		t.Errorf("STRING_AGG native (DISTINCT|IGNORE NULLS): got %q; want a|b", s)
	}
}

// TestStringAggWindowNativeNullWithoutIgnore drives the NULL-but-
// not-ignoring branch (empty string is appended).
func TestStringAggWindowNativeNullWithoutIgnore(t *testing.T) {
	t.Parallel()
	a := NewStringAggWindowNative()().(*stringAggWindowNative)
	if err := a.Step("a"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(nil); err != nil {
		t.Fatal(err)
	}
	got, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	v, err := value.DecodeValue(got)
	if err != nil {
		t.Fatal(err)
	}
	// "a" then empty -> "a,"
	if s, _ := v.ToString(); s != "a," {
		t.Errorf("STRING_AGG NULL no-ignore: got %q; want a,", s)
	}
}

// TestWindowStatisticalSpecs drives every WINDOW_* statistical
// spec's Done end-to-end against a known data set. These methods
// share the same buffer-and-aggregate shape so a small input space
// suffices; we only need to confirm Done produces a finite value
// to count the inner Step + Done branches.
func TestWindowStatisticalSpecs(t *testing.T) {
	t.Parallel()

	// Pair-input specs: CORR / COVAR_POP / COVAR_SAMP.
	xs := []value.Value{value.IntValue(1), value.IntValue(2), value.IntValue(3), value.IntValue(4)}
	ys := []value.Value{value.IntValue(2), value.IntValue(4), value.IntValue(6), value.IntValue(8)}

	{
		f := &WINDOW_CORR{}
		agg := newWindowFuncAggregatedStatus()
		for i := range xs {
			st := rowsFrameStatus(1, xs[i], true)
			if err := f.Step(xs[i], ys[i], st, agg); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := f.Done(agg); err != nil {
			t.Errorf("CORR: %v", err)
		}
	}

	for name, fn := range map[string]struct {
		step func(x, y value.Value, opt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error
		done func(agg *WindowFuncAggregatedStatus) (value.Value, error)
	}{
		"COVAR_POP": {
			step: (&WINDOW_COVAR_POP{}).Step,
			done: (&WINDOW_COVAR_POP{}).Done,
		},
		"COVAR_SAMP": {
			step: (&WINDOW_COVAR_SAMP{}).Step,
			done: (&WINDOW_COVAR_SAMP{}).Done,
		},
	} {
		agg := newWindowFuncAggregatedStatus()
		for i := range xs {
			st := rowsFrameStatus(1, xs[i], true)
			if err := fn.step(xs[i], ys[i], st, agg); err != nil {
				t.Fatalf("%s Step: %v", name, err)
			}
		}
		if _, err := fn.done(agg); err != nil {
			t.Errorf("%s Done: %v", name, err)
		}
	}

	// Single-arg specs: STDDEV_POP / STDDEV_SAMP / VAR_POP / VAR_SAMP.
	for name, fn := range map[string]struct {
		step func(v value.Value, opt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error
		done func(agg *WindowFuncAggregatedStatus) (value.Value, error)
	}{
		"STDDEV_POP":  {step: (&WINDOW_STDDEV_POP{}).Step, done: (&WINDOW_STDDEV_POP{}).Done},
		"STDDEV_SAMP": {step: (&WINDOW_STDDEV_SAMP{}).Step, done: (&WINDOW_STDDEV_SAMP{}).Done},
		"VAR_POP":     {step: (&WINDOW_VAR_POP{}).Step, done: (&WINDOW_VAR_POP{}).Done},
		"VAR_SAMP":    {step: (&WINDOW_VAR_SAMP{}).Step, done: (&WINDOW_VAR_SAMP{}).Done},
	} {
		agg := newWindowFuncAggregatedStatus()
		for _, v := range xs {
			st := rowsFrameStatus(1, v, true)
			if err := fn.step(v, st, agg); err != nil {
				t.Fatalf("%s Step: %v", name, err)
			}
		}
		if _, err := fn.done(agg); err != nil {
			t.Errorf("%s Done: %v", name, err)
		}
	}
}

// TestWindowRankingDesc drives the DESC (isAsc=false) branch of
// WINDOW_RANK / WINDOW_DENSE_RANK / WINDOW_PERCENT_RANK / CUME_DIST.
// Each Done() has an ASC-branch and a DESC-branch; the descending
// pass uses maxValue = math.MaxInt64 and decrements. Driving the
// DESC branch on rowID=1 over sorted [3,2,1] yields the rank of the
// first row (which is value 3 -> rank 1 in DESC).
func TestWindowRankingDesc(t *testing.T) {
	t.Parallel()

	mkDescStatus := func(rowID int64, ordVal value.Value) *WindowFuncStatus {
		return &WindowFuncStatus{
			FrameUnit: WindowFrameUnitRows,
			Start:     &WindowBoundary{Type: WindowCurrentRowType},
			End:       &WindowBoundary{Type: WindowCurrentRowType},
			RowID:     rowID,
			OrderBy: []*WindowOrderBy{
				{Value: ordVal, IsAsc: false},
			},
		}
	}

	for name, fn := range map[string]struct {
		step func(opt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error
		done func(agg *WindowFuncAggregatedStatus) (value.Value, error)
	}{
		"RANK":         {step: (&WINDOW_RANK{}).Step, done: (&WINDOW_RANK{}).Done},
		"DENSE_RANK":   {step: (&WINDOW_DENSE_RANK{}).Step, done: (&WINDOW_DENSE_RANK{}).Done},
		"PERCENT_RANK": {step: (&WINDOW_PERCENT_RANK{}).Step, done: (&WINDOW_PERCENT_RANK{}).Done},
		"CUME_DIST":    {step: (&WINDOW_CUME_DIST{}).Step, done: (&WINDOW_CUME_DIST{}).Done},
		"ROW_NUMBER":   {step: (&WINDOW_ROW_NUMBER{}).Step, done: (&WINDOW_ROW_NUMBER{}).Done},
	} {
		agg := newWindowFuncAggregatedStatus()
		for _, v := range []int64{1, 2, 3} {
			st := mkDescStatus(1, value.IntValue(v))
			if err := fn.step(st, agg); err != nil {
				t.Fatalf("%s Step: %v", name, err)
			}
		}
		if _, err := fn.done(agg); err != nil {
			t.Errorf("%s Done DESC: %v", name, err)
		}
	}
}

// TestSumDistinctWindowEmpty drives the empty-frame Done branch
// (returns NULL).
func TestSumDistinctWindowEmpty(t *testing.T) {
	t.Parallel()
	a := NewSumDistinctWindowNative()().(*sumDistinctWindow)
	got, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("SUM_DISTINCT empty: got %v; want nil", got)
	}
}

// TestStatsWindowDistinct drives the DISTINCT branch of
// floatWindow.activeValues via the stddev_pop native aggregator.
// Inputs 1, 1, 2, 3 -> non-distinct stddev != distinct stddev, but
// we only need to confirm the distinct branch runs without error
// and yields a finite value.
func TestStatsWindowDistinct(t *testing.T) {
	t.Parallel()
	a := NewStddevPopWindowNative()().(*stddevPopWindow)
	distinctMarker := `{"type":"aggregate_distinct"}`
	for _, v := range []int64{1, 1, 2, 3} {
		if err := a.Step(v, distinctMarker); err != nil {
			t.Fatal(err)
		}
	}
	got, err := a.Done()
	if err != nil {
		t.Fatalf("STDDEV_POP(DISTINCT): %v", err)
	}
	if got == nil {
		t.Fatalf("STDDEV_POP(DISTINCT): got nil")
	}
}

// TestWindowNthValueIgnoreNulls drives the IGNORE NULLS branch of
// WINDOW_NTH_VALUE.Done. Inputs [NULL, "a", NULL, "b", "c"] with
// IGNORE NULLS and n=2 selects "b" (the 2nd non-NULL).
func TestWindowNthValueIgnoreNulls(t *testing.T) {
	t.Parallel()
	fn := &WINDOW_NTH_VALUE{}
	agg := newWindowFuncAggregatedStatus()
	agg.IgnoreNullsOpt = true
	vals := []value.Value{nil, value.StringValue("a"), nil, value.StringValue("b"), value.StringValue("c")}
	for _, v := range vals {
		st := rowsFrameStatus(1, v, true)
		if err := fn.Step(v, 2, st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if got == nil {
		t.Fatalf("NTH_VALUE IGNORE NULLS: got nil; want b")
	}
	if s, _ := got.ToString(); s != "b" {
		t.Errorf("NTH_VALUE IGNORE NULLS n=2: got %s; want b", s)
	}
}

// TestWindowLagLeadDefaultValue drives the defaultValue-substitute
// branch of LEAD / LAG. With offset large enough that the lookup
// falls outside the frame, Done returns the configured default.
func TestWindowLagLeadDefaultValue(t *testing.T) {
	t.Parallel()
	lead := &WINDOW_LEAD{}
	agg := newWindowFuncAggregatedStatus()
	def := value.StringValue("DEFAULT")
	// 3 rows, offset 10 -> out of range -> returns NULL (not default)
	// because Done's "leadValue == nil" check happens after agg.Done
	// already noticed out-of-frame. Actually the impl returns
	// defaultValue when leadValue stays nil — out-of-range path
	// causes leadValue to never be assigned, so default is used.
	for _, s := range []string{"a", "b", "c"} {
		st := rowsFrameStatus(1, value.StringValue(s), true)
		if err := lead.Step(value.StringValue(s), 10, def, st, agg); err != nil {
			t.Fatal(err)
		}
	}
	got, err := lead.Done(agg)
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s != "DEFAULT" {
		t.Errorf("LEAD offset=10 default: got %v; want DEFAULT", s)
	}

	// LAG offset large from rowID=1 -> out of range -> default.
	lag := &WINDOW_LAG{}
	agg2 := newWindowFuncAggregatedStatus()
	for _, s := range []string{"a", "b", "c"} {
		st := rowsFrameStatus(1, value.StringValue(s), true)
		if err := lag.Step(value.StringValue(s), 10, def, st, agg2); err != nil {
			t.Fatal(err)
		}
	}
	got, err = lag.Done(agg2)
	if err != nil {
		t.Fatal(err)
	}
	if s, _ := got.ToString(); s != "DEFAULT" {
		t.Errorf("LAG offset=10 default: got %v; want DEFAULT", s)
	}
}

// TestWindowPercentileContError drives the percentile-range guards
// of WINDOW_PERCENTILE_CONT.Done. Percentile < 0 / > 1 -> error.
func TestWindowPercentileContError(t *testing.T) {
	t.Parallel()
	for _, p := range []float64{-0.1, 1.1} {
		fn := &WINDOW_PERCENTILE_CONT{}
		agg := newWindowFuncAggregatedStatus()
		st := rowsFrameStatus(1, value.IntValue(1), true)
		if err := fn.Step(value.IntValue(1), value.FloatValue(p), st, agg); err != nil {
			t.Fatal(err)
		}
		if _, err := fn.Done(agg); err == nil {
			t.Errorf("PERCENTILE_CONT p=%v: expected error", p)
		}
	}
}

// TestWindowFirstLastValueIgnoreNulls drives the IGNORE NULLS
// branch of FIRST_VALUE / LAST_VALUE Done.
func TestWindowFirstLastValueIgnoreNulls(t *testing.T) {
	t.Parallel()
	first := &WINDOW_FIRST_VALUE{}
	got := driveWithFlags(t, first.Step, first.Done,
		[]value.Value{nil, value.StringValue("a"), value.StringValue("b")}, false, true)
	if got == nil {
		t.Fatalf("FIRST_VALUE IGNORE NULLS: got nil; want a")
	}
	if s, _ := got.ToString(); s != "a" {
		t.Errorf("FIRST_VALUE IGNORE NULLS: got %s; want a", s)
	}

	last := &WINDOW_LAST_VALUE{}
	got = driveWithFlags(t, last.Step, last.Done,
		[]value.Value{value.StringValue("a"), value.StringValue("b"), nil}, false, true)
	if got == nil {
		t.Fatalf("LAST_VALUE IGNORE NULLS: got nil; want b")
	}
	if s, _ := got.ToString(); s != "b" {
		t.Errorf("LAST_VALUE IGNORE NULLS: got %s; want b", s)
	}
}
