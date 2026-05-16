// Direct unit tests for the RANGE-frame branches of
// WindowFuncAggregatedStatus.getIndexFromBoundaryByRange and the
// partition-aware index/range helpers in option.go. These paths only
// run when the analyzer materialises a `RANGE BETWEEN N PRECEDING AND
// N FOLLOWING` window via the predecessor-emulation path, which is
// only routed for DISTINCT / IGNORE NULLS aggregates that the SQLite
// native engine can't drive natively. The unit tests below bypass that
// gating by constructing the status directly — the same shape the
// formatter would emit for a `LOGICAL_OR(DISTINCT b) OVER (ORDER BY t
// RANGE BETWEEN 1 PRECEDING AND 1 FOLLOWING)` query.
//
// Expected indices come from the upstream BigQuery semantics
// documented at
// https://github.com/google/googlesql/blob/master/docs/window-function-calls.md#range-window-frame-rules
// "RANGE BETWEEN <preceding> AND <following>": the start is the row
// with the smallest ORDER BY value >= current-value - preceding; the
// end is the row with the largest ORDER BY value <= current-value +
// following.

package window

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// rangeFrameStatus builds a WindowFuncStatus for the
// `ORDER BY t RANGE BETWEEN <preceding> PRECEDING AND <following>
// FOLLOWING` frame.
func rangeFrameStatus(rowID int64, orderByVal value.Value, preceding, following int64) *WindowFuncStatus {
	return &WindowFuncStatus{
		FrameUnit: WindowFrameUnitRange,
		Start:     &WindowBoundary{Type: WindowOffsetPrecedingType, Offset: preceding},
		End:       &WindowBoundary{Type: WindowOffsetFollowingType, Offset: following},
		RowID:     rowID,
		OrderBy: []*WindowOrderBy{
			{Value: orderByVal, IsAsc: true},
		},
	}
}

// TestRangeFrameOffsetPrecedingFollowing drives
// `getIndexFromBoundaryByRange` for OffsetPreceding and OffsetFollowing
// branches, plus `currentRangeValue` and the lookupMin/Max helpers.
//
// Six rows with ORDER BY values 1..5 and 10. Current row 3 has value
// 3. Expected frame for RANGE BETWEEN 1 PRECEDING AND 1 FOLLOWING:
// values in [2, 4] -> indices 1..3 (rows with values 2, 3, 4).
func TestRangeFrameOffsetPrecedingFollowing(t *testing.T) {
	t.Parallel()
	agg := newWindowFuncAggregatedStatus()
	vals := []int64{1, 2, 3, 4, 5, 10}
	for _, v := range vals {
		st := rangeFrameStatus(3, value.IntValue(v), 1, 1)
		if err := agg.Step(value.IntValue(v), st); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	var gotStart, gotEnd int
	var gotVals []value.Value
	if err := agg.Done(func(values []value.Value, start, end int) error {
		gotStart, gotEnd = start, end
		gotVals = values
		return nil
	}); err != nil {
		t.Fatalf("Done: %v", err)
	}
	if gotStart != 1 || gotEnd != 3 {
		t.Fatalf("RANGE frame for current=3, preceding=1, following=1: got start=%d end=%d; want 1..3", gotStart, gotEnd)
	}
	if len(gotVals) != 6 {
		t.Fatalf("got %d values; want 6", len(gotVals))
	}
}

// TestRangeFrameUnboundedPrecedingCurrentRow drives the
// WindowUnboundedPrecedingType and WindowCurrentRowType branches of
// getIndexFromBoundaryByRange. Frame `RANGE BETWEEN UNBOUNDED
// PRECEDING AND CURRENT ROW` over ORDER BY values 1, 2, 3 with current
// row at 2 -> frame indices 0..1 (the row with value 2 is the last
// row whose ORDER BY <= 2).
func TestRangeFrameUnboundedPrecedingCurrentRow(t *testing.T) {
	t.Parallel()
	agg := newWindowFuncAggregatedStatus()
	for _, v := range []int64{1, 2, 3} {
		st := &WindowFuncStatus{
			FrameUnit: WindowFrameUnitRange,
			Start:     &WindowBoundary{Type: WindowUnboundedPrecedingType},
			End:       &WindowBoundary{Type: WindowCurrentRowType},
			RowID:     2,
			OrderBy: []*WindowOrderBy{
				{Value: value.IntValue(v), IsAsc: true},
			},
		}
		if err := agg.Step(value.IntValue(v), st); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	var gotStart, gotEnd int
	if err := agg.Done(func(_ []value.Value, start, end int) error {
		gotStart, gotEnd = start, end
		return nil
	}); err != nil {
		t.Fatalf("Done: %v", err)
	}
	if gotStart != 0 || gotEnd != 1 {
		t.Fatalf("UNBOUNDED PRECEDING/CURRENT ROW: got start=%d end=%d; want 0..1", gotStart, gotEnd)
	}
}

// TestRangeFrameUnboundedFollowing drives the
// WindowUnboundedFollowingType branch. Frame `RANGE BETWEEN CURRENT
// ROW AND UNBOUNDED FOLLOWING` over 1, 2, 3 at current row=1 (value=1)
// -> frame indices 0..2.
func TestRangeFrameUnboundedFollowing(t *testing.T) {
	t.Parallel()
	agg := newWindowFuncAggregatedStatus()
	for _, v := range []int64{1, 2, 3} {
		st := &WindowFuncStatus{
			FrameUnit: WindowFrameUnitRange,
			Start:     &WindowBoundary{Type: WindowCurrentRowType},
			End:       &WindowBoundary{Type: WindowUnboundedFollowingType},
			RowID:     1,
			OrderBy: []*WindowOrderBy{
				{Value: value.IntValue(v), IsAsc: true},
			},
		}
		if err := agg.Step(value.IntValue(v), st); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	var gotStart, gotEnd int
	if err := agg.Done(func(_ []value.Value, start, end int) error {
		gotStart, gotEnd = start, end
		return nil
	}); err != nil {
		t.Fatalf("Done: %v", err)
	}
	if gotStart != 0 || gotEnd != 2 {
		t.Fatalf("CURRENT ROW/UNBOUNDED FOLLOWING: got start=%d end=%d; want 0..2", gotStart, gotEnd)
	}
}

// TestRangeFramePartitioned drives partitionedCurrentRangeValue and
// partitionedCurrentIndexByRows. Two partitions (g=A, g=B) with two
// rows each. Current row is the second row of partition A (value=20),
// so the frame `RANGE BETWEEN 5 PRECEDING AND CURRENT ROW` selects
// rows in partition A with ORDER BY in [15, 20] -> just the row with
// value=20 (index 1 inside the filtered slice).
func TestRangeFramePartitioned(t *testing.T) {
	t.Parallel()
	agg := newWindowFuncAggregatedStatus()
	rows := []struct {
		part string
		ord  int64
	}{
		{"A", 10},
		{"A", 20},
		{"B", 100},
		{"B", 200},
	}
	// Current output row is rowID=2 (partition A, value 20).
	for _, r := range rows {
		st := &WindowFuncStatus{
			FrameUnit:  WindowFrameUnitRange,
			Start:      &WindowBoundary{Type: WindowOffsetPrecedingType, Offset: 5},
			End:        &WindowBoundary{Type: WindowCurrentRowType},
			RowID:      2,
			Partitions: []value.Value{value.StringValue(r.part)},
			OrderBy: []*WindowOrderBy{
				{Value: value.IntValue(r.ord), IsAsc: true},
			},
		}
		if err := agg.Step(value.IntValue(r.ord), st); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	var gotStart, gotEnd int
	var gotVals []value.Value
	if err := agg.Done(func(values []value.Value, start, end int) error {
		gotStart, gotEnd = start, end
		gotVals = values
		return nil
	}); err != nil {
		t.Fatalf("Done: %v", err)
	}
	// Filtered values for partition A are [10, 20]. ORDER BY value at
	// the current row (rowID=2 -> partitionedCurrentRangeValue) is 20;
	// preceding 5 -> 15 (lookupMinIndex points at the first row with
	// value >= 15, which is index 1 in sorted [10, 20]). Current row
	// at value=20 (lookupMaxIndex points at the last row with value
	// <= 20, also index 1).
	if gotStart != 1 || gotEnd != 1 {
		t.Fatalf("partitioned RANGE frame at rowID=2: got start=%d end=%d; want 1..1", gotStart, gotEnd)
	}
	if len(gotVals) != 2 {
		t.Fatalf("filtered values: got %d; want 2 (partition A rows)", len(gotVals))
	}
}

// TestRowsFrameOffsetBranches drives the OffsetPreceding /
// OffsetFollowing branches of getIndexFromBoundaryByRows over a non-
// partitioned scan. Frame `ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING`
// on rows [10, 20, 30] at current row index 1 -> 0..2.
func TestRowsFrameOffsetBranches(t *testing.T) {
	t.Parallel()
	agg := newWindowFuncAggregatedStatus()
	vals := []int64{10, 20, 30}
	for _, v := range vals {
		st := &WindowFuncStatus{
			FrameUnit: WindowFrameUnitRows,
			Start:     &WindowBoundary{Type: WindowOffsetPrecedingType, Offset: 1},
			End:       &WindowBoundary{Type: WindowOffsetFollowingType, Offset: 1},
			RowID:     2, // current row index 1 (0-based 1 in sorted slice)
			OrderBy: []*WindowOrderBy{
				{Value: value.IntValue(v), IsAsc: true},
			},
		}
		if err := agg.Step(value.IntValue(v), st); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	var gotStart, gotEnd int
	if err := agg.Done(func(_ []value.Value, start, end int) error {
		gotStart, gotEnd = start, end
		return nil
	}); err != nil {
		t.Fatalf("Done: %v", err)
	}
	if gotStart != 0 || gotEnd != 2 {
		t.Fatalf("ROWS frame OffsetPreceding/Following at rowID=2: got start=%d end=%d; want 0..2", gotStart, gotEnd)
	}
}

// TestRowsFramePartitionedOffsetBranches drives
// partitionedCurrentIndexByRows for the ROWS-unit OffsetPreceding /
// OffsetFollowing branches. Two partitions A and B with three rows
// each; current row is the middle row of B (rowID=5 across the
// combined stream, which is the second row of partition B).
func TestRowsFramePartitionedOffsetBranches(t *testing.T) {
	t.Parallel()
	agg := newWindowFuncAggregatedStatus()
	rows := []struct {
		part string
		ord  int64
	}{
		{"A", 1},
		{"A", 2},
		{"A", 3},
		{"B", 10},
		{"B", 20},
		{"B", 30},
	}
	// Current row is rowID=5 (partition B, value=20, the middle row of
	// partition B). Frame `ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING`
	// over partition B (sorted) -> indices 0..2 of the filtered slice.
	for _, r := range rows {
		st := &WindowFuncStatus{
			FrameUnit:  WindowFrameUnitRows,
			Start:      &WindowBoundary{Type: WindowOffsetPrecedingType, Offset: 1},
			End:        &WindowBoundary{Type: WindowOffsetFollowingType, Offset: 1},
			RowID:      5,
			Partitions: []value.Value{value.StringValue(r.part)},
			OrderBy: []*WindowOrderBy{
				{Value: value.IntValue(r.ord), IsAsc: true},
			},
		}
		if err := agg.Step(value.IntValue(r.ord), st); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	var gotStart, gotEnd int
	var gotVals []value.Value
	if err := agg.Done(func(values []value.Value, start, end int) error {
		gotStart, gotEnd = start, end
		gotVals = values
		return nil
	}); err != nil {
		t.Fatalf("Done: %v", err)
	}
	if gotStart != 0 || gotEnd != 2 {
		t.Fatalf("partitioned ROWS frame at rowID=5: got start=%d end=%d; want 0..2", gotStart, gotEnd)
	}
	if len(gotVals) != 3 {
		t.Fatalf("filtered values: got %d; want 3 (partition B rows)", len(gotVals))
	}
}

// TestArrayAggWindowNativeIgnoreNullsDistinct drives both Done
// short-circuits in arrayAggWindowNative: ignoreNulls skips NULL
// entries, and distinct deduplicates by stringified key. With
// ignore_nulls + distinct enabled and the marker option strings fed in
// through Step, Done should emit [1, 2, 3] from input [1, NULL, 2, 2,
// 1, 3].
func TestArrayAggWindowNativeIgnoreNullsDistinct(t *testing.T) {
	t.Parallel()
	a := NewArrayAggWindowNative()().(*arrayAggWindowNative)
	// The aggregator helpers parse DISTINCT / IGNORE NULLS out of the
	// argument list using helper.ParseOptions; the option encoding is
	// a JSON tag with type "aggregate_distinct" / "aggregate_ignore_nulls".
	distinctMarker := `{"type":"aggregate_distinct"}`
	ignoreNullsMarker := `{"type":"aggregate_ignore_nulls"}`
	steps := []any{int64(1), nil, int64(2), int64(2), int64(1), int64(3)}
	for _, v := range steps {
		if err := a.Step(v, distinctMarker, ignoreNullsMarker); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	got, err := a.Done()
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if got == nil {
		t.Fatalf("ARRAY_AGG(DISTINCT IGNORE NULLS): got nil; want non-nil")
	}
	v, err := value.DecodeValue(got)
	if err != nil {
		t.Fatalf("DecodeValue: %v", err)
	}
	arr, err := v.ToArray()
	if err != nil {
		t.Fatalf("ToArray: %v", err)
	}
	// Input [1, NULL, 2, 2, 1, 3] with IGNORE NULLS + DISTINCT ->
	// [1, 2, 3] (insertion order, NULL dropped, duplicates dropped).
	if len(arr.Values) != 3 {
		t.Fatalf("ARRAY_AGG(DISTINCT IGNORE NULLS): got %d distinct values; want 3", len(arr.Values))
	}
}

// TestArrayAggWindowNativeAllNullIgnore drives the ignoreNulls branch
// where every value is NULL so the deduped slice is empty and Done
// returns NULL.
func TestArrayAggWindowNativeAllNullIgnore(t *testing.T) {
	t.Parallel()
	a := NewArrayAggWindowNative()().(*arrayAggWindowNative)
	ignoreNullsMarker := `{"type":"aggregate_ignore_nulls"}`
	for i := 0; i < 3; i++ {
		if err := a.Step(nil, ignoreNullsMarker); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	got, err := a.Done()
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if got != nil {
		t.Fatalf("ARRAY_AGG all-NULL IGNORE NULLS: got %v; want NULL", got)
	}
}
