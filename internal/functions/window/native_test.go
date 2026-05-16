// Step / Inverse / Done unit tests for the native SQLite-frame-driven
// window aggregators (the ones created via the NewXxxWindowNative
// constructors and wired through Conn.CreateWindowFunction). These
// keep their state purely inside the per-instance struct, so they're
// straightforward to drive in isolation.

package window

import (
	"math"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestArrayAggWindowNative(t *testing.T) {
	t.Parallel()
	a := NewArrayAggWindowNative()().(*arrayAggWindowNative)

	if err := a.Step(int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(2)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(3)); err != nil {
		t.Fatal(err)
	}
	// Inverse pops the oldest entry.
	if err := a.Inverse(int64(1)); err != nil {
		t.Fatal(err)
	}
	got, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	// EncodeValue returns a string; the array contents matter, not
	// the exact wire form, so re-decode through DecodeValue.
	if got == nil {
		t.Fatalf("ARRAY_AGG native: got nil")
	}
	v, err := value.DecodeValue(got)
	if err != nil {
		t.Fatal(err)
	}
	arr, err := v.ToArray()
	if err != nil {
		t.Fatalf("ToArray: %v", err)
	}
	if len(arr.Values) != 2 {
		t.Fatalf("ARRAY_AGG native: got %d values, want 2", len(arr.Values))
	}
	// Empty frame -> NULL.
	empty := NewArrayAggWindowNative()().(*arrayAggWindowNative)
	if got, _ := empty.Done(); got != nil {
		t.Fatalf("empty ARRAY_AGG: got %v, want NULL", got)
	}
}

func TestStringAggWindowNative(t *testing.T) {
	t.Parallel()
	a := NewStringAggWindowNative()().(*stringAggWindowNative)
	// Use default delim ",".
	for _, s := range []string{"a", "b", "c"} {
		if err := a.Step(s); err != nil {
			t.Fatal(err)
		}
	}
	out, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	v, _ := value.DecodeValue(out)
	if s, _ := v.ToString(); s != "a,b,c" {
		t.Fatalf("STRING_AGG default delim: got %s, want a,b,c", s)
	}
	// Inverse pops oldest.
	if err := a.Inverse("a"); err != nil {
		t.Fatal(err)
	}
	out, _ = a.Done()
	v, _ = value.DecodeValue(out)
	if s, _ := v.ToString(); s != "b,c" {
		t.Fatalf("STRING_AGG after inverse: got %s, want b,c", s)
	}
}

func TestCountifWindowNative(t *testing.T) {
	t.Parallel()
	a := NewCountifWindowNative()().(*countifWindowNative)
	if err := a.Step(true); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(false); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(true); err != nil {
		t.Fatal(err)
	}
	if got, _ := a.Done(); got.(int64) != 2 {
		t.Fatalf("COUNTIF: got %v, want 2", got)
	}
	// Inverse with TRUE should decrement.
	if err := a.Inverse(true); err != nil {
		t.Fatal(err)
	}
	if got, _ := a.Done(); got.(int64) != 1 {
		t.Fatalf("COUNTIF after inverse(true): got %v, want 1", got)
	}
	// Inverse with FALSE keeps count.
	if err := a.Inverse(false); err != nil {
		t.Fatal(err)
	}
	if got, _ := a.Done(); got.(int64) != 1 {
		t.Fatalf("COUNTIF after inverse(false): got %v, want 1", got)
	}
}

func TestCountStarWindowNative(t *testing.T) {
	t.Parallel()
	a := NewCountStarWindowNative()().(*countStarWindowNative)
	if err := a.Step(); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(); err != nil {
		t.Fatal(err)
	}
	if got, _ := a.Done(); got.(int64) != 3 {
		t.Fatalf("COUNT(*): got %v, want 3", got)
	}
	// Inverse decrements.
	if err := a.Inverse(); err != nil {
		t.Fatal(err)
	}
	if got, _ := a.Done(); got.(int64) != 2 {
		t.Fatalf("COUNT(*) after inverse: got %v, want 2", got)
	}
}

func TestAnyValueWindowNative(t *testing.T) {
	t.Parallel()
	a := NewAnyValueWindowNative()().(*anyValueWindowNative)
	if err := a.Step("x"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("y"); err != nil {
		t.Fatal(err)
	}
	out, _ := a.Done()
	v, _ := value.DecodeValue(out)
	if s, _ := v.ToString(); s != "x" {
		t.Fatalf("ANY_VALUE native: got %s, want x", s)
	}
	// Inverse pops oldest.
	if err := a.Inverse("x"); err != nil {
		t.Fatal(err)
	}
	out, _ = a.Done()
	v, _ = value.DecodeValue(out)
	if s, _ := v.ToString(); s != "y" {
		t.Fatalf("ANY_VALUE after inverse: got %s, want y", s)
	}
	// All-NULL -> NULL.
	empty := NewAnyValueWindowNative()().(*anyValueWindowNative)
	if err := empty.Step(nil); err != nil {
		t.Fatal(err)
	}
	if out, _ := empty.Done(); out != nil {
		t.Fatalf("ANY_VALUE all-NULL: got %v, want NULL", out)
	}
}

func TestLogicalOrAndWindowNative(t *testing.T) {
	t.Parallel()
	// LOGICAL_OR
	or := NewLogicalOrWindowNative()().(*logicalOrWindowNative)
	if err := or.Step(false); err != nil {
		t.Fatal(err)
	}
	if err := or.Step(true); err != nil {
		t.Fatal(err)
	}
	out, _ := or.Done()
	v, _ := value.DecodeValue(out)
	if b, _ := v.ToBool(); !b {
		t.Errorf("LOGICAL_OR(F,T): expected TRUE")
	}
	// All FALSE
	or2 := NewLogicalOrWindowNative()().(*logicalOrWindowNative)
	_ = or2.Step(false)
	_ = or2.Step(false)
	out, _ = or2.Done()
	v, _ = value.DecodeValue(out)
	if b, _ := v.ToBool(); b {
		t.Errorf("LOGICAL_OR(F,F): expected FALSE")
	}
	// All NULL -> NULL.
	or3 := NewLogicalOrWindowNative()().(*logicalOrWindowNative)
	_ = or3.Step(nil)
	_ = or3.Step(nil)
	if got, _ := or3.Done(); got != nil {
		t.Errorf("LOGICAL_OR(NULL,NULL): expected NULL")
	}

	// LOGICAL_AND
	and := NewLogicalAndWindowNative()().(*logicalAndWindowNative)
	_ = and.Step(true)
	_ = and.Step(false)
	out, _ = and.Done()
	v, _ = value.DecodeValue(out)
	if b, _ := v.ToBool(); b {
		t.Errorf("LOGICAL_AND(T,F): expected FALSE")
	}
	// All TRUE
	and2 := NewLogicalAndWindowNative()().(*logicalAndWindowNative)
	_ = and2.Step(true)
	_ = and2.Step(true)
	out, _ = and2.Done()
	v, _ = value.DecodeValue(out)
	if b, _ := v.ToBool(); !b {
		t.Errorf("LOGICAL_AND(T,T): expected TRUE")
	}
	// All NULL
	and3 := NewLogicalAndWindowNative()().(*logicalAndWindowNative)
	_ = and3.Step(nil)
	_ = and3.Step(nil)
	if got, _ := and3.Done(); got != nil {
		t.Errorf("LOGICAL_AND(NULL,NULL): expected NULL")
	}
	// Inverse on logicalOr / logicalAnd pops front.
	_ = or.Inverse()
	_ = and.Inverse()
}

func TestBitAggWindowNative(t *testing.T) {
	t.Parallel()
	and := NewBitAndAggWindowNative()().(*bitAndAggWindowNative)
	or := NewBitOrAggWindowNative()().(*bitOrAggWindowNative)
	xor := NewBitXorAggWindowNative()().(*bitXorAggWindowNative)
	for _, v := range []int64{0b1100, 0b1010} {
		if err := and.Step(v); err != nil {
			t.Fatal(err)
		}
		if err := or.Step(v); err != nil {
			t.Fatal(err)
		}
		if err := xor.Step(v); err != nil {
			t.Fatal(err)
		}
	}
	// BIT_AND(0b1100, 0b1010) = 0b1000 = 8.
	if got, _ := and.Done(); got.(int64) != 0b1000 {
		t.Errorf("BIT_AND: got %v, want %d", got, 0b1000)
	}
	// BIT_OR = 0b1110 = 14.
	if got, _ := or.Done(); got.(int64) != 0b1110 {
		t.Errorf("BIT_OR: got %v, want %d", got, 0b1110)
	}
	// BIT_XOR = 0b0110 = 6.
	if got, _ := xor.Done(); got.(int64) != 0b0110 {
		t.Errorf("BIT_XOR: got %v, want %d", got, 0b0110)
	}
	// Inverse pops oldest.
	_ = and.Inverse()
	_ = or.Inverse()
	_ = xor.Inverse()
	// Empty frame returns nil.
	empty := NewBitAndAggWindowNative()().(*bitAndAggWindowNative)
	if got, _ := empty.Done(); got != nil {
		t.Errorf("empty BIT_AND: got %v, want NULL", got)
	}
}

func TestArrayConcatAggWindowNative(t *testing.T) {
	t.Parallel()
	a := NewArrayConcatAggWindowNative()().(*arrayConcatAggWindowNative)
	// Build encoded arrays to feed through Step (ConvertArgs decodes
	// them via the base64-JSON envelope). The simplest path is to
	// pre-build value.Value arrays and run them through EncodeValue.
	arr1 := &value.ArrayValue{Values: []value.Value{value.IntValue(1), value.IntValue(2)}}
	arr2 := &value.ArrayValue{Values: []value.Value{value.IntValue(3)}}
	enc1, _ := value.EncodeValue(arr1)
	enc2, _ := value.EncodeValue(arr2)
	if err := a.Step(enc1); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(enc2); err != nil {
		t.Fatal(err)
	}
	out, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	v, _ := value.DecodeValue(out)
	arr, _ := v.ToArray()
	if len(arr.Values) != 3 {
		t.Fatalf("ARRAY_CONCAT_AGG: got %d, want 3", len(arr.Values))
	}
	// Inverse pops oldest array.
	_ = a.Inverse()
	out, _ = a.Done()
	v, _ = value.DecodeValue(out)
	arr, _ = v.ToArray()
	if len(arr.Values) != 1 {
		t.Fatalf("ARRAY_CONCAT_AGG after inverse: got %d, want 1", len(arr.Values))
	}
	// NULL input is skipped (BQ-compatible silent drop).
	empty := NewArrayConcatAggWindowNative()().(*arrayConcatAggWindowNative)
	if err := empty.Step(nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := empty.Done(); got != nil {
		t.Errorf("ARRAY_CONCAT_AGG empty: got %v, want NULL", got)
	}
}

func TestSumCountAvgDistinctWindowNative(t *testing.T) {
	t.Parallel()

	// SUM_DISTINCT
	sum := NewSumDistinctWindowNative()().(*sumDistinctWindow)
	for _, v := range []int64{1, 2, 2, 3} {
		if err := sum.Step(v); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := sum.Done(); got.(int64) != 6 {
		t.Fatalf("SUM_DISTINCT: got %v, want 6", got)
	}
	// Inverse pops front.
	_ = sum.Inverse()
	if got, _ := sum.Done(); got.(int64) != 5 {
		t.Fatalf("SUM_DISTINCT after inverse: got %v, want 5", got)
	}

	// COUNT_DISTINCT
	count := NewCountDistinctWindowNative()().(*countDistinctWindow)
	for _, v := range []int64{1, 2, 2, 3} {
		_ = count.Step(v)
	}
	if got, _ := count.Done(); got.(int64) != 3 {
		t.Fatalf("COUNT_DISTINCT: got %v, want 3", got)
	}

	// AVG_DISTINCT
	avg := NewAvgDistinctWindowNative()().(*avgDistinctWindow)
	for _, v := range []int64{1, 2, 2, 3} {
		_ = avg.Step(v)
	}
	if got, _ := avg.Done(); got.(float64) != 2.0 {
		t.Fatalf("AVG_DISTINCT: got %v, want 2.0", got)
	}

	// All-NULL frames -> NULL for SUM / AVG, 0 for COUNT.
	sumNull := NewSumDistinctWindowNative()().(*sumDistinctWindow)
	_ = sumNull.Step(nil)
	if got, _ := sumNull.Done(); got != nil {
		t.Errorf("SUM_DISTINCT all-NULL: got %v, want NULL", got)
	}
	avgNull := NewAvgDistinctWindowNative()().(*avgDistinctWindow)
	_ = avgNull.Step(nil)
	if got, _ := avgNull.Done(); got != nil {
		t.Errorf("AVG_DISTINCT all-NULL: got %v, want NULL", got)
	}
	countNull := NewCountDistinctWindowNative()().(*countDistinctWindow)
	_ = countNull.Step(nil)
	if got, _ := countNull.Done(); got.(int64) != 0 {
		t.Errorf("COUNT_DISTINCT all-NULL: got %v, want 0", got)
	}

	// All-integer SUM path: 1 + 2 + 3 = 6.
	sumInts := NewSumDistinctWindowNative()().(*sumDistinctWindow)
	_ = sumInts.Step(int64(1))
	_ = sumInts.Step(int64(2))
	_ = sumInts.Step(int64(3))
	if got, _ := sumInts.Done(); got.(int64) != 6 {
		t.Errorf("SUM_DISTINCT all-int: got %v (%T), want int64(6)", got, got)
	}
}

// TestWindowFuncOptionMarkers covers the WINDOW_* option-marker
// constructors and the matching parseWindowOptions decode path.
func TestWindowFuncOptionMarkers(t *testing.T) {
	t.Parallel()

	frame, err := WINDOW_FRAME_UNIT(int64(WindowFrameUnitRows))
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
	rowid, err := WINDOW_ROWID(42)
	if err != nil {
		t.Fatal(err)
	}
	part, err := WINDOW_PARTITION(value.StringValue("p1"))
	if err != nil {
		t.Fatal(err)
	}
	orderBy, err := WINDOW_ORDER_BY(value.IntValue(7), true)
	if err != nil {
		t.Fatal(err)
	}

	_, opt := parseWindowOptions(frame, start, end, rowid, part, orderBy)
	if opt.FrameUnit != WindowFrameUnitRows {
		t.Errorf("FrameUnit: got %v, want Rows", opt.FrameUnit)
	}
	if opt.Start == nil || opt.Start.Type != WindowUnboundedPrecedingType {
		t.Errorf("Start: got %+v", opt.Start)
	}
	if opt.End == nil || opt.End.Type != WindowUnboundedFollowingType {
		t.Errorf("End: got %+v", opt.End)
	}
	if opt.RowID != 42 {
		t.Errorf("RowID: got %d, want 42", opt.RowID)
	}
	if len(opt.Partitions) != 1 {
		t.Errorf("Partitions: got %d, want 1", len(opt.Partitions))
	}
	if len(opt.OrderBy) != 1 || !opt.OrderBy[0].IsAsc {
		t.Errorf("OrderBy: got %+v", opt.OrderBy)
	}
}

// TestWindowFuncStatus_Partition exercises the Partition()
// concatenation helper.
func TestWindowFuncStatus_Partition(t *testing.T) {
	t.Parallel()

	s := &WindowFuncStatus{Partitions: []value.Value{
		value.StringValue("a"),
		value.IntValue(1),
	}}
	got, err := s.Partition()
	if err != nil {
		t.Fatal(err)
	}
	if got != "a_1" {
		t.Errorf("Partition: got %q, want a_1", got)
	}
}

// TestPercentileContWindowNative drives the frame-driven
// PERCENTILE_CONT.  PERCENTILE_CONT(x, 0.5) over [0, 2] -> 1 (linear
// interpolation between the two values).
func TestPercentileContWindowNative(t *testing.T) {
	t.Parallel()
	a := NewPercentileContWindowNative()().(*percentileContWindow)
	for _, v := range []int64{0, 2} {
		if err := a.Step(v, 0.5); err != nil {
			t.Fatal(err)
		}
	}
	out, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	f := out.(float64)
	if f != 1.0 {
		t.Fatalf("PERCENTILE_CONT(0.5) of [0,2]: got %v, want 1.0", f)
	}
	// Empty frame -> NULL.
	empty := NewPercentileContWindowNative()().(*percentileContWindow)
	if got, _ := empty.Done(); got != nil {
		t.Errorf("empty PERCENTILE_CONT: got %v, want NULL", got)
	}
	// Inverse pops front.
	_ = a.Inverse()
	out, _ = a.Done()
	// After dropping the 0, frame is [2], so any percentile returns 2.
	f = out.(float64)
	if f != 2.0 {
		t.Errorf("PERCENTILE_CONT after inverse: got %v, want 2", f)
	}
	// Bad percentile rejected.
	bad := NewPercentileContWindowNative()().(*percentileContWindow)
	if err := bad.Step(int64(1), float64(2)); err == nil {
		t.Errorf("PERCENTILE_CONT pct=2 should error")
	}
	bad2 := NewPercentileContWindowNative()().(*percentileContWindow)
	if err := bad2.Step(int64(1)); err == nil {
		t.Errorf("PERCENTILE_CONT requires (value, pct) args")
	}
}

func TestPercentileDiscWindowNative(t *testing.T) {
	t.Parallel()
	a := NewPercentileDiscWindowNative()().(*percentileDiscWindow)
	for _, v := range []int64{0, 2, 4} {
		if err := a.Step(v, 0.5); err != nil {
			t.Fatal(err)
		}
	}
	out, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	// idx = ceil(0.5*3)-1 = 1, xs sorted = [0,2,4] -> xs[1]=2.
	v, _ := value.DecodeValue(out)
	if x, _ := v.ToInt64(); x != 2 {
		t.Fatalf("PERCENTILE_DISC(0.5) of [0,2,4]: got %d, want 2", x)
	}
	empty := NewPercentileDiscWindowNative()().(*percentileDiscWindow)
	if got, _ := empty.Done(); got != nil {
		t.Errorf("empty PERCENTILE_DISC: got %v, want NULL", got)
	}
}

func TestStddevVarWindowNatives(t *testing.T) {
	t.Parallel()

	rows := []int64{2, 4, 4, 4, 5, 5, 7, 9}

	pop := NewStddevPopWindowNative()().(*stddevPopWindow)
	for _, v := range rows {
		_ = pop.Step(v)
	}
	out, _ := pop.Done()
	if f := out.(float64); f != 2.0 {
		t.Errorf("STDDEV_POP: got %v, want 2.0", f)
	}

	samp := NewStddevSampWindowNative()().(*stddevSampWindow)
	for _, v := range rows {
		_ = samp.Step(v)
	}
	out, _ = samp.Done()
	if f := out.(float64); math.Abs(f-math.Sqrt(32.0/7.0)) > 1e-9 {
		t.Errorf("STDDEV_SAMP: got %v, want sqrt(32/7)", f)
	}

	vp := NewVarPopWindowNative()().(*varPopWindow)
	for _, v := range rows {
		_ = vp.Step(v)
	}
	out, _ = vp.Done()
	if f := out.(float64); f != 4.0 {
		t.Errorf("VAR_POP: got %v, want 4.0", f)
	}

	vs := NewVarSampWindowNative()().(*varSampWindow)
	for _, v := range rows {
		_ = vs.Step(v)
	}
	out, _ = vs.Done()
	if f := out.(float64); math.Abs(f-32.0/7.0) > 1e-9 {
		t.Errorf("VAR_SAMP: got %v, want 32/7", f)
	}

	// Empty / single-row sample stddev returns NULL.
	emptySamp := NewStddevSampWindowNative()().(*stddevSampWindow)
	if got, _ := emptySamp.Done(); got != nil {
		t.Errorf("empty STDDEV_SAMP: got %v, want NULL", got)
	}
	singleSamp := NewStddevSampWindowNative()().(*stddevSampWindow)
	_ = singleSamp.Step(int64(1))
	if got, _ := singleSamp.Done(); got != nil {
		t.Errorf("single STDDEV_SAMP: got %v, want NULL", got)
	}
	// Pop with one row returns 0 (variance of a single value).
	popSingle := NewStddevPopWindowNative()().(*stddevPopWindow)
	_ = popSingle.Step(int64(5))
	if got, _ := popSingle.Done(); got.(float64) != 0 {
		t.Errorf("single STDDEV_POP: got %v, want 0", got)
	}
	// Inverse pops oldest.
	_ = pop.Inverse()
	// After popping, the population stddev of the trailing 7 elements
	// should still be a real number (no panic).
	if _, err := pop.Done(); err != nil {
		t.Errorf("STDDEV_POP after inverse: %v", err)
	}
}

func TestCorrCovarWindowNatives(t *testing.T) {
	t.Parallel()

	// For (x, y) = (1,1), (2,2), (3,3):
	//   CORR = 1 (perfect correlation), COVAR_POP = 2/3, COVAR_SAMP = 1.
	corr := NewCorrWindowNative()().(*corrWindow)
	covPop := NewCovarPopWindowNative()().(*covarPopWindow)
	covSamp := NewCovarSampWindowNative()().(*covarSampWindow)
	for _, v := range []int64{1, 2, 3} {
		if err := corr.Step(v, v); err != nil {
			t.Fatal(err)
		}
		if err := covPop.Step(v, v); err != nil {
			t.Fatal(err)
		}
		if err := covSamp.Step(v, v); err != nil {
			t.Fatal(err)
		}
	}
	out, _ := corr.Done()
	if f := out.(float64); math.Abs(f-1.0) > 1e-9 {
		t.Errorf("CORR: got %v, want 1.0", f)
	}
	out, _ = covPop.Done()
	if f := out.(float64); math.Abs(f-2.0/3.0) > 1e-9 {
		t.Errorf("COVAR_POP: got %v, want 0.666...", f)
	}
	out, _ = covSamp.Done()
	if f := out.(float64); math.Abs(f-1.0) > 1e-9 {
		t.Errorf("COVAR_SAMP: got %v, want 1.0", f)
	}
	// Single-row CORR / COVAR_SAMP -> NULL.
	emptyCorr := NewCorrWindowNative()().(*corrWindow)
	_ = emptyCorr.Step(int64(1), int64(1))
	if got, _ := emptyCorr.Done(); got != nil {
		t.Errorf("single CORR: got %v, want NULL", got)
	}
	// Inverse pops front.
	_ = corr.Inverse()
	_ = covPop.Inverse()
	_ = covSamp.Inverse()
	// NULL operand -> skipped.
	corrNull := NewCorrWindowNative()().(*corrWindow)
	_ = corrNull.Step(nil, int64(1))
	if got, _ := corrNull.Done(); got != nil {
		t.Errorf("CORR NULL operand: got %v, want NULL", got)
	}
}

// TestGetOptionFuncSQL covers the SQL-string emitters that the
// analyzer-bridge uses to embed window option markers into the
// rewritten query. Each helper just stringifies its inputs; the test
// pins the format.
func TestGetOptionFuncSQL(t *testing.T) {
	t.Parallel()
	if got := GetWindowRowIDOptionFuncSQL(); got != "googlesqlite_window_rowid(`row_id`)" {
		t.Errorf("RowID SQL: got %q", got)
	}
	if got := GetWindowPartitionOptionFuncSQL("col"); got != "googlesqlite_window_partition(col)" {
		t.Errorf("Partition SQL: got %q", got)
	}
	if got := GetWindowOrderByOptionFuncSQL("col", true); got != "googlesqlite_window_order_by(col, true)" {
		t.Errorf("OrderBy SQL: got %q", got)
	}
	// Frame-unit / boundary SQL helpers stringify the typed ints.
	if got := GetWindowFrameUnitOptionFuncSQL(0); got == "" {
		t.Errorf("FrameUnit SQL empty")
	}
	if got := GetWindowBoundaryStartOptionFuncSQL(0, ""); got == "" {
		t.Errorf("BoundaryStart SQL empty")
	}
	if got := GetWindowBoundaryEndOptionFuncSQL(0, "42"); got == "" {
		t.Errorf("BoundaryEnd SQL empty")
	}
}

// TestDistinctInverseAndPartition exercises the popFront branch of
// distinct.go for COUNT_DISTINCT and AVG_DISTINCT (SUM_DISTINCT's
// inverse path is already covered).
func TestDistinctInverseAndPartition(t *testing.T) {
	t.Parallel()
	count := NewCountDistinctWindowNative()().(*countDistinctWindow)
	for _, v := range []int64{1, 2, 3} {
		_ = count.Step(v)
	}
	if got, _ := count.Done(); got.(int64) != 3 {
		t.Fatalf("COUNT_DISTINCT before inverse: %v", got)
	}
	_ = count.Inverse()
	if got, _ := count.Done(); got.(int64) != 2 {
		t.Fatalf("COUNT_DISTINCT after inverse: %v", got)
	}

	avg := NewAvgDistinctWindowNative()().(*avgDistinctWindow)
	for _, v := range []int64{2, 4, 6} {
		_ = avg.Step(v)
	}
	if got, _ := avg.Done(); got.(float64) != 4 {
		t.Fatalf("AVG_DISTINCT: %v", got)
	}
	_ = avg.Inverse()
	if got, _ := avg.Done(); got.(float64) != 5 {
		t.Fatalf("AVG_DISTINCT after inverse: %v", got)
	}

	// Percentile / stddev / var Inverse paths.
	pd := NewPercentileDiscWindowNative()().(*percentileDiscWindow)
	for _, v := range []int64{1, 2, 3} {
		_ = pd.Step(v, 0.5)
	}
	_ = pd.Inverse()
	if _, err := pd.Done(); err != nil {
		t.Errorf("PERCENTILE_DISC inverse: %v", err)
	}
	samp := NewStddevSampWindowNative()().(*stddevSampWindow)
	for _, v := range []int64{1, 2, 3} {
		_ = samp.Step(v)
	}
	_ = samp.Inverse()
	if _, err := samp.Done(); err != nil {
		t.Errorf("STDDEV_SAMP inverse: %v", err)
	}
	vp := NewVarPopWindowNative()().(*varPopWindow)
	for _, v := range []int64{1, 2, 3} {
		_ = vp.Step(v)
	}
	_ = vp.Inverse()
	if _, err := vp.Done(); err != nil {
		t.Errorf("VAR_POP inverse: %v", err)
	}
	vs := NewVarSampWindowNative()().(*varSampWindow)
	for _, v := range []int64{1, 2, 3} {
		_ = vs.Step(v)
	}
	_ = vs.Inverse()
	if _, err := vs.Done(); err != nil {
		t.Errorf("VAR_SAMP inverse: %v", err)
	}
}

// TestWindowFuncStatus_Partitioned exercises the Partition helper on
// WindowFuncAggregatedStatus when partition columns are set.
func TestWindowFuncStatus_Partitioned(t *testing.T) {
	t.Parallel()

	// Build a partition-aware step sequence: 2 rows in partition "a",
	// 1 row in partition "b". Then verify the FilteredValues helper
	// only returns the current row's partition entries.
	agg := newWindowFuncAggregatedStatus()
	makeOpt := func(partition string, rowID int64, orderByVal value.Value) *WindowFuncStatus {
		return &WindowFuncStatus{
			FrameUnit:  WindowFrameUnitRows,
			Start:      &WindowBoundary{Type: WindowUnboundedPrecedingType},
			End:        &WindowBoundary{Type: WindowUnboundedFollowingType},
			RowID:      rowID,
			Partitions: []value.Value{value.StringValue(partition)},
			OrderBy: []*WindowOrderBy{
				{Value: orderByVal, IsAsc: true},
			},
		}
	}

	if err := agg.Step(value.IntValue(1), makeOpt("a", 1, value.IntValue(1))); err != nil {
		t.Fatal(err)
	}
	if err := agg.Step(value.IntValue(2), makeOpt("a", 1, value.IntValue(2))); err != nil {
		t.Fatal(err)
	}
	if err := agg.Step(value.IntValue(99), makeOpt("b", 1, value.IntValue(99))); err != nil {
		t.Fatal(err)
	}
	// First two rows are in partition "a", last row in "b". Partition
	// for RowID=1 is "a".
	if got := agg.Partition(); got != "a" {
		t.Errorf("Partition: got %q, want a", got)
	}
	filtered := agg.FilteredValues()
	if len(filtered) != 2 {
		t.Errorf("FilteredValues for partition a: got %d, want 2", len(filtered))
	}
}

// TestAggregatorSurface covers WindowAggregator.Step / Done via the
// public NewWindowAggregator constructor. Step parses option markers,
// applies once-set IgnoreNulls / Distinct, then delegates to the
// caller-supplied step. Done runs the caller-supplied done and
// encodes its result.
func TestAggregatorSurface(t *testing.T) {
	t.Parallel()
	var captured []value.Value
	wa := NewWindowAggregator(
		func(args []value.Value, _ *WindowFuncStatus, _ *WindowFuncAggregatedStatus) error {
			captured = append(captured, args...)
			return nil
		},
		func(_ *WindowFuncAggregatedStatus) (value.Value, error) {
			return value.IntValue(int64(len(captured))), nil
		},
	)
	if err := wa.Step(int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := wa.Step(int64(2)); err != nil {
		t.Fatal(err)
	}
	got, err := wa.Done()
	if err != nil {
		t.Fatal(err)
	}
	if got.(int64) != 2 {
		t.Fatalf("aggregator captured count: got %v, want 2", got)
	}
}

// TestBindStringAggLeadLagNthValue exercises the per-Bind*
// constructor code paths that have extra argument-handling logic.
func TestBindStringAggLeadLagNthValue(t *testing.T) {
	t.Parallel()

	// Build a frame option marker so we can drive the bind wrapper
	// against a synthetic SQLite-style arg list (the bind wrapper
	// strips option markers via parseWindowOptions before forwarding
	// the value args to the per-spec Step).
	frameUnit, _ := WINDOW_FRAME_UNIT(int64(WindowFrameUnitRows))
	start, _ := WINDOW_BOUNDARY_START(int64(WindowUnboundedPrecedingType), 0)
	end, _ := WINDOW_BOUNDARY_END(int64(WindowUnboundedFollowingType), 0)
	rowid, _ := WINDOW_ROWID(1)
	orderBy, _ := WINDOW_ORDER_BY(value.IntValue(1), true)
	markersStr := func() []any {
		out := make([]any, 0, 5)
		for _, m := range []value.Value{frameUnit, start, end, rowid, orderBy} {
			s, _ := m.ToString()
			out = append(out, s)
		}
		return out
	}

	// STRING_AGG with explicit delim "|".
	agg := BindWindowStringAgg()()
	for _, v := range []string{"a", "b", "c"} {
		args := []any{v, "|"}
		args = append(args, markersStr()...)
		if err := agg.Step(args...); err != nil {
			t.Fatal(err)
		}
	}
	out, err := agg.Done()
	if err != nil {
		t.Fatal(err)
	}
	v, _ := value.DecodeValue(out)
	if s, _ := v.ToString(); s != "a|b|c" {
		t.Errorf("Bind STRING_AGG: got %s, want a|b|c", s)
	}

	// LEAD with explicit offset.
	lead := BindWindowLead()()
	for _, v := range []string{"a", "b", "c"} {
		args := []any{v, int64(1), "def"}
		args = append(args, markersStr()...)
		if err := lead.Step(args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := lead.Done(); err != nil {
		t.Errorf("Bind LEAD: %v", err)
	}
	// LAG with explicit offset.
	lag := BindWindowLag()()
	for _, v := range []string{"a", "b", "c"} {
		args := []any{v, int64(1), "def"}
		args = append(args, markersStr()...)
		if err := lag.Step(args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := lag.Done(); err != nil {
		t.Errorf("Bind LAG: %v", err)
	}
	// LEAD with NULL offset -> error.
	leadBad := BindWindowLead()()
	args := []any{"a", nil, "def"}
	args = append(args, markersStr()...)
	if err := leadBad.Step(args...); err == nil {
		t.Errorf("Bind LEAD: NULL offset must error")
	}
	// LAG with negative offset -> error.
	lagBad := BindWindowLag()()
	args = []any{"a", int64(-1), "def"}
	args = append(args, markersStr()...)
	if err := lagBad.Step(args...); err == nil {
		t.Errorf("Bind LAG: negative offset must error")
	}

	// NTH_VALUE with NULL num -> error.
	nthBad := BindWindowNthValue()()
	args = []any{"a", nil}
	args = append(args, markersStr()...)
	if err := nthBad.Step(args...); err == nil {
		t.Errorf("Bind NTH_VALUE: NULL num must error")
	}
}

// TestBindCtorsSmoke confirms each BindWindowX constructor returns a
// non-nil aggregator (it's a thin wrapper but its presence guarantees
// none of the constructors panic under their once.Do init).
func TestBindCtorsSmoke(t *testing.T) {
	t.Parallel()

	type ctor func() func() *WindowAggregator
	for _, c := range []ctor{
		BindWindowAnyValue, BindWindowArrayAgg, BindWindowAvg,
		BindWindowCount, BindWindowCountStar, BindWindowCountIf,
		BindWindowLogicalOr, BindWindowLogicalAnd,
		BindWindowMax, BindWindowMin, BindWindowStringAgg,
		BindWindowSum, BindWindowCorr, BindWindowCovarPop,
		BindWindowCovarSamp, BindWindowStddevPop,
		BindWindowStddevSamp, BindWindowStddev,
		BindWindowVarPop, BindWindowVarSamp, BindWindowVariance,
		BindWindowFirstValue, BindWindowLastValue,
		BindWindowNthValue, BindWindowLead, BindWindowLag,
		BindWindowPercentileCont, BindWindowPercentileDisc,
		BindWindowRank, BindWindowDenseRank, BindWindowPercentRank,
		BindWindowCumeDist, BindWindowNtile, BindWindowRowNumber,
	} {
		factory := c()
		if factory == nil {
			t.Errorf("ctor returned nil factory")
		}
		agg := factory()
		if agg == nil {
			t.Errorf("factory returned nil aggregator")
		}
	}
}
