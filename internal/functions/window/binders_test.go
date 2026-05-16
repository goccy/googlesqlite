// Direct exercise of each BindWindow* constructor. The predecessor-
// emulation path of the formatter routes through these binders for
// any window call the SQLite-native engine can't drive directly. The
// existing parity / spectest layers don't reach them because the
// formatter prefers customNativeWindowFuncMap / SQLite-native
// versions for nearly every aggregator. These tests bypass the
// formatter and drive Step / Done on the raw *WindowAggregator
// returned by each binder.
//
// All assertions match the semantics documented for the underlying
// WINDOW_* spec in funcs.go and verified by the funcs_test.go unit
// tests against authoritative upstream BigQuery / GoogleSQL Examples.

package window

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// driveBinder runs the binder's Step over each row, then Done.
// `stepArgs` is the slice handed to Step (including any option
// markers); the final element of each call is the encoded
// WindowFuncStatus payload so the aggregator picks up the same
// frame/order/rowID as the analyzer would emit.
func driveBinder(t *testing.T, agg *WindowAggregator, rows [][]any) any {
	t.Helper()
	for _, r := range rows {
		if err := agg.Step(r...); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	got, err := agg.Done()
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	return got
}

// rowsOptMarkers returns the encoded marker strings for the
// {frame=ROWS, start=UNBOUNDED_PRECEDING, end=UNBOUNDED_FOLLOWING,
// rowID=row} window — i.e. SUM(x) OVER () shape.
func rowsOptMarkers(t *testing.T, row int64) []any {
	t.Helper()
	fu, err := WINDOW_FRAME_UNIT(int64(WindowFrameUnitRows))
	if err != nil {
		t.Fatal(err)
	}
	start, err := WINDOW_BOUNDARY_START(int64(WindowUnboundedPrecedingType), 0)
	if err != nil {
		t.Fatal(err)
	}
	end, err := WINDOW_BOUNDARY_END(int64(WindowUnboundedFollowingType), 0)
	if err != nil {
		t.Fatal(err)
	}
	rid, err := WINDOW_ROWID(row)
	if err != nil {
		t.Fatal(err)
	}
	return []any{toString(t, fu), toString(t, start), toString(t, end), toString(t, rid)}
}

func toString(t *testing.T, v value.Value) string {
	t.Helper()
	s, err := v.ToString()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// numericRows replays every value in `vals` against the binder, each
// tagged with the same window markers (one aggregator instance per
// row=1 — Step is called once per partition row).
func numericRows(t *testing.T, vals []int64) [][]any {
	t.Helper()
	markers := rowsOptMarkers(t, 1)
	rows := make([][]any, len(vals))
	for i, v := range vals {
		row := append([]any{v}, markers...)
		rows[i] = row
	}
	return rows
}

// TestBindWindowSum exercises BindWindowSum end-to-end. SUM(1, 2, 3)
// over (UNBOUNDED PRECEDING and FOLLOWING) -> 6, matching the
// upstream BigQuery SUM semantics ([SUM] aggregates over every row
// in the frame).
func TestBindWindowSum(t *testing.T) {
	t.Parallel()
	agg := BindWindowSum()()
	got := driveBinder(t, agg, numericRows(t, []int64{1, 2, 3}))
	if got == nil {
		t.Fatalf("SUM: got nil; want 6")
	}
	v, err := value.DecodeValue(got)
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := v.ToInt64(); i != 6 {
		t.Fatalf("SUM: got %d; want 6", i)
	}
}

// TestBindWindowAvgCount drives the AVG / COUNT / COUNT_STAR /
// COUNTIF binders. Each operates over the same UNBOUNDED-frame
// markers we used for SUM, with the value space chosen so the
// expected output is uniquely determined by upstream semantics.
func TestBindWindowAvgCount(t *testing.T) {
	t.Parallel()
	for name, bind := range map[string]func() func() *WindowAggregator{
		"AVG":   BindWindowAvg,
		"COUNT": BindWindowCount,
	} {
		agg := bind()()
		got := driveBinder(t, agg, numericRows(t, []int64{2, 4, 6}))
		if got == nil {
			t.Fatalf("%s: got nil", name)
		}
		// Just confirm we got back a value; specific equality is
		// covered by funcs_test.go.
		_ = got
	}

	// COUNT(*) doesn't take a value arg in the analyzer signature,
	// but the binder's step closure still strips the marker tail.
	agg := BindWindowCountStar()()
	rows := [][]any{}
	markers := rowsOptMarkers(t, 1)
	for i := 0; i < 3; i++ {
		row := append([]any{}, markers...)
		rows = append(rows, row)
	}
	got := driveBinder(t, agg, rows)
	v, err := value.DecodeValue(got)
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := v.ToInt64(); i != 3 {
		t.Fatalf("COUNT(*): got %d; want 3", i)
	}

	// COUNTIF
	cif := BindWindowCountIf()()
	cifRows := [][]any{}
	for _, b := range []bool{true, false, true} {
		row := append([]any{b}, markers...)
		cifRows = append(cifRows, row)
	}
	got2 := driveBinder(t, cif, cifRows)
	v2, _ := value.DecodeValue(got2)
	if i, _ := v2.ToInt64(); i != 2 {
		t.Fatalf("COUNTIF: got %d; want 2", i)
	}
}

// TestBindWindowMaxMin drives MAX / MIN.
func TestBindWindowMaxMin(t *testing.T) {
	t.Parallel()
	for name, bind := range map[string]struct {
		bind func() func() *WindowAggregator
		want int64
	}{
		"MAX": {BindWindowMax, 3},
		"MIN": {BindWindowMin, 1},
	} {
		agg := bind.bind()()
		got := driveBinder(t, agg, numericRows(t, []int64{1, 2, 3}))
		v, _ := value.DecodeValue(got)
		if i, _ := v.ToInt64(); i != bind.want {
			t.Fatalf("%s: got %d; want %d", name, i, bind.want)
		}
	}
}

// TestBindWindowLogical drives LOGICAL_OR / LOGICAL_AND.
func TestBindWindowLogical(t *testing.T) {
	t.Parallel()
	markers := rowsOptMarkers(t, 1)
	{
		agg := BindWindowLogicalOr()()
		rows := [][]any{
			append([]any{false}, markers...),
			append([]any{true}, markers...),
		}
		got := driveBinder(t, agg, rows)
		v, _ := value.DecodeValue(got)
		if b, _ := v.ToBool(); !b {
			t.Errorf("LOGICAL_OR(F,T): want TRUE")
		}
	}
	{
		agg := BindWindowLogicalAnd()()
		rows := [][]any{
			append([]any{true}, markers...),
			append([]any{false}, markers...),
		}
		got := driveBinder(t, agg, rows)
		v, _ := value.DecodeValue(got)
		if b, _ := v.ToBool(); b {
			t.Errorf("LOGICAL_AND(T,F): want FALSE")
		}
	}
}

// TestBindWindowAnyValueArrayAggStringAgg drives the value-bearing
// non-numeric binders. ANY_VALUE returns the first value; ARRAY_AGG
// returns an ArrayValue with len=N; STRING_AGG with a "," delimiter
// returns "a,b,c".
func TestBindWindowAnyValueArrayAggStringAgg(t *testing.T) {
	t.Parallel()
	markers := rowsOptMarkers(t, 1)

	anyAgg := BindWindowAnyValue()()
	rows := [][]any{
		append([]any{"x"}, markers...),
		append([]any{"y"}, markers...),
	}
	got := driveBinder(t, anyAgg, rows)
	v, _ := value.DecodeValue(got)
	if s, _ := v.ToString(); s != "x" {
		t.Errorf("ANY_VALUE: got %s; want x", s)
	}

	arr := BindWindowArrayAgg()()
	rows = [][]any{
		append([]any{int64(1)}, markers...),
		append([]any{int64(2)}, markers...),
	}
	got = driveBinder(t, arr, rows)
	v, _ = value.DecodeValue(got)
	a, _ := v.ToArray()
	if len(a.Values) != 2 {
		t.Errorf("ARRAY_AGG: got %d values; want 2", len(a.Values))
	}

	str := BindWindowStringAgg()()
	rows = [][]any{
		append([]any{"a", ","}, markers...),
		append([]any{"b", ","}, markers...),
		append([]any{"c", ","}, markers...),
	}
	got = driveBinder(t, str, rows)
	v, _ = value.DecodeValue(got)
	if s, _ := v.ToString(); s != "a,b,c" {
		t.Errorf("STRING_AGG: got %s; want a,b,c", s)
	}
}

// TestBindWindowFirstLastNthValue drives the navigation binders.
// FIRST_VALUE -> "a"; LAST_VALUE -> "c"; NTH_VALUE(_, 2) -> "b" over
// the UNBOUNDED frame.
func TestBindWindowFirstLastNthValue(t *testing.T) {
	t.Parallel()
	markers := rowsOptMarkers(t, 1)
	mkRows := func(vals ...string) [][]any {
		r := [][]any{}
		for _, s := range vals {
			r = append(r, append([]any{s}, markers...))
		}
		return r
	}

	{
		agg := BindWindowFirstValue()()
		got := driveBinder(t, agg, mkRows("a", "b", "c"))
		v, _ := value.DecodeValue(got)
		if s, _ := v.ToString(); s != "a" {
			t.Errorf("FIRST_VALUE: got %s; want a", s)
		}
	}
	{
		agg := BindWindowLastValue()()
		got := driveBinder(t, agg, mkRows("a", "b", "c"))
		v, _ := value.DecodeValue(got)
		if s, _ := v.ToString(); s != "c" {
			t.Errorf("LAST_VALUE: got %s; want c", s)
		}
	}
	{
		agg := BindWindowNthValue()()
		rows := [][]any{}
		for _, s := range []string{"a", "b", "c"} {
			// NTH_VALUE takes (value, n) plus markers.
			rows = append(rows, append([]any{s, int64(2)}, markers...))
		}
		got := driveBinder(t, agg, rows)
		v, _ := value.DecodeValue(got)
		if s, _ := v.ToString(); s != "b" {
			t.Errorf("NTH_VALUE(_, 2): got %s; want b", s)
		}
	}
}

// TestBindWindowLeadLag drives BindWindowLead / BindWindowLag with
// the 1-arg form (offset defaults to 1, default value NULL). The
// specific output is unit-tested directly elsewhere; here we only
// confirm the binder closures run end-to-end without error.
func TestBindWindowLeadLag(t *testing.T) {
	t.Parallel()
	markers := func(row int64) []any { return rowsOptMarkers(t, row) }

	for name, bind := range map[string]func() func() *WindowAggregator{
		"LEAD": BindWindowLead,
		"LAG":  BindWindowLag,
	} {
		agg := bind()()
		for _, s := range []string{"a", "b", "c"} {
			row := append([]any{s}, markers(2)...)
			if err := agg.Step(row...); err != nil {
				t.Fatalf("%s Step: %v", name, err)
			}
		}
		if _, err := agg.Done(); err != nil {
			t.Fatalf("%s Done: %v", name, err)
		}
	}
}

// TestBindWindowStatistical drives every statistical binder. Each
// just funnels into the underlying WINDOW_<NAME> spec which is unit-
// tested elsewhere; here we only confirm the binder closures run.
func TestBindWindowStatistical(t *testing.T) {
	t.Parallel()
	markers := rowsOptMarkers(t, 1)
	mkRows := func(xs, ys []int64) [][]any {
		r := [][]any{}
		for i, x := range xs {
			r = append(r, append([]any{x, ys[i]}, markers...))
		}
		return r
	}
	xs := []int64{1, 2, 3, 4}
	ys := []int64{2, 4, 6, 8}

	for name, b := range map[string]func() func() *WindowAggregator{
		"CORR":       BindWindowCorr,
		"COVAR_POP":  BindWindowCovarPop,
		"COVAR_SAMP": BindWindowCovarSamp,
	} {
		agg := b()()
		if _, err := func() (any, error) {
			for _, r := range mkRows(xs, ys) {
				if err := agg.Step(r...); err != nil {
					return nil, err
				}
			}
			return agg.Done()
		}(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}

	// Single-value variance / stddev binders.
	uni := numericRows(t, xs)
	for name, b := range map[string]func() func() *WindowAggregator{
		"STDDEV_POP":  BindWindowStddevPop,
		"STDDEV_SAMP": BindWindowStddevSamp,
		"STDDEV":      BindWindowStddev,
		"VAR_POP":     BindWindowVarPop,
		"VAR_SAMP":    BindWindowVarSamp,
		"VARIANCE":    BindWindowVariance,
	} {
		agg := b()()
		if _, err := func() (any, error) {
			for _, r := range uni {
				if err := agg.Step(r...); err != nil {
					return nil, err
				}
			}
			return agg.Done()
		}(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestBindWindowPercentile drives the PERCENTILE_CONT / PERCENTILE_DISC
// binders. Each is called with (value, percentile_arg=0.5) over the
// UNBOUNDED frame.
func TestBindWindowPercentile(t *testing.T) {
	t.Parallel()
	markers := rowsOptMarkers(t, 1)
	mkRows := func(vals []int64, p float64) [][]any {
		r := [][]any{}
		for _, v := range vals {
			r = append(r, append([]any{v, p}, markers...))
		}
		return r
	}
	for name, b := range map[string]func() func() *WindowAggregator{
		"PERCENTILE_CONT": BindWindowPercentileCont,
		"PERCENTILE_DISC": BindWindowPercentileDisc,
	} {
		agg := b()()
		rows := mkRows([]int64{1, 2, 3}, 0.5)
		for _, r := range rows {
			if err := agg.Step(r...); err != nil {
				t.Errorf("%s Step: %v", name, err)
			}
		}
		if _, err := agg.Done(); err != nil {
			t.Errorf("%s Done: %v", name, err)
		}
	}
}

// TestBindWindowNumbering drives the rank / dense_rank / percent_rank
// / cume_dist / ntile / row_number binders. Each takes the markers
// (no value args except ntile's bucket count).
func TestBindWindowNumbering(t *testing.T) {
	t.Parallel()
	for name, b := range map[string]func() func() *WindowAggregator{
		"RANK":         BindWindowRank,
		"DENSE_RANK":   BindWindowDenseRank,
		"PERCENT_RANK": BindWindowPercentRank,
		"CUME_DIST":    BindWindowCumeDist,
		"ROW_NUMBER":   BindWindowRowNumber,
	} {
		agg := b()()
		// Numbering functions are ROWS-frame current-row aware.
		// Drive them with rowID=2 against 3 rows so SortedValues has
		// 3 entries and the result is well defined.
		fu, _ := WINDOW_FRAME_UNIT(int64(WindowFrameUnitRows))
		start, _ := WINDOW_BOUNDARY_START(int64(WindowCurrentRowType), 0)
		end, _ := WINDOW_BOUNDARY_END(int64(WindowCurrentRowType), 0)
		rid, _ := WINDOW_ROWID(2)
		ord, _ := WINDOW_ORDER_BY(value.IntValue(2), true)
		markers := []any{toString(t, fu), toString(t, start), toString(t, end), toString(t, rid), toString(t, ord)}
		// Re-Step three rows with the same rowID (the same output row
		// being aggregated over the partition).
		for _, v := range []int64{1, 2, 3} {
			ordV, _ := WINDOW_ORDER_BY(value.IntValue(v), true)
			perRow := []any{toString(t, fu), toString(t, start), toString(t, end), toString(t, rid), toString(t, ordV)}
			_ = markers // marker order doesn't matter once stripped
			if err := agg.Step(perRow...); err != nil {
				t.Fatalf("%s Step: %v", name, err)
			}
		}
		if _, err := agg.Done(); err != nil {
			t.Errorf("%s Done: %v", name, err)
		}
	}

	// NTILE takes a bucket count arg.
	agg := BindWindowNtile()()
	fu, _ := WINDOW_FRAME_UNIT(int64(WindowFrameUnitRows))
	start, _ := WINDOW_BOUNDARY_START(int64(WindowCurrentRowType), 0)
	end, _ := WINDOW_BOUNDARY_END(int64(WindowCurrentRowType), 0)
	rid, _ := WINDOW_ROWID(1)
	for _, v := range []int64{1, 2, 3} {
		ordV, _ := WINDOW_ORDER_BY(value.IntValue(v), true)
		row := []any{int64(2), toString(t, fu), toString(t, start), toString(t, end), toString(t, rid), toString(t, ordV)}
		if err := agg.Step(row...); err != nil {
			t.Fatalf("NTILE Step: %v", err)
		}
	}
	if _, err := agg.Done(); err != nil {
		t.Errorf("NTILE Done: %v", err)
	}
}
