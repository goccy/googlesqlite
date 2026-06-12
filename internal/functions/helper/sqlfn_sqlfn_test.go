package helper_test

// Unit tests for the Scalar* binder helpers. These exercise the four
// invariants every wrapper has to honour: arity checking, NULL
// propagation, KeepNull observers, and successful arg fan-out. The
// ExistsNull helper is also covered here because its contract is
// trivially observable (any nil in the slice => true).

import (
	"errors"
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/intervalvalue"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// identity1 is a no-op semantic that lets us observe whether the
// wrapper actually called us with the expected arg.
func identity1(v value.Value) (value.Value, error) { return v, nil }

func TestScalar1_Arity(t *testing.T) {
	t.Parallel()

	fn := helper.Scalar1(identity1)
	if _, err := fn(); err == nil {
		t.Fatalf("expected arity error for 0 args")
	}
	if _, err := fn(value.IntValue(1), value.IntValue(2)); err == nil {
		t.Fatalf("expected arity error for 2 args")
	}
}

func TestScalar1_NullPropagation(t *testing.T) {
	t.Parallel()

	fn := helper.Scalar1(identity1)
	got, err := fn(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for NULL input, got %v", got)
	}
}

func TestScalar1_Success(t *testing.T) {
	t.Parallel()

	fn := helper.Scalar1(identity1)
	got, err := fn(value.IntValue(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	i, err := got.ToInt64()
	if err != nil {
		t.Fatalf("ToInt64: %v", err)
	}
	if i != 42 {
		t.Fatalf("expected 42, got %d", i)
	}
}

func TestScalar1_PropagatesInnerError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("boom")
	fn := helper.Scalar1(func(_ value.Value) (value.Value, error) {
		return nil, sentinel
	})
	if _, err := fn(value.IntValue(1)); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestScalar2(t *testing.T) {
	t.Parallel()

	add := func(a, b value.Value) (value.Value, error) { return a.Add(b) }
	fn := helper.Scalar2(add)

	t.Run("arity_low", func(t *testing.T) {
		t.Parallel()
		if _, err := fn(value.IntValue(1)); err == nil {
			t.Fatal("expected arity error")
		}
	})
	t.Run("arity_high", func(t *testing.T) {
		t.Parallel()
		if _, err := fn(value.IntValue(1), value.IntValue(2), value.IntValue(3)); err == nil {
			t.Fatal("expected arity error")
		}
	})
	t.Run("null_first", func(t *testing.T) {
		t.Parallel()
		got, err := fn(nil, value.IntValue(2))
		if err != nil || got != nil {
			t.Fatalf("expected (nil,nil), got (%v,%v)", got, err)
		}
	})
	t.Run("null_second", func(t *testing.T) {
		t.Parallel()
		got, err := fn(value.IntValue(1), nil)
		if err != nil || got != nil {
			t.Fatalf("expected (nil,nil), got (%v,%v)", got, err)
		}
	})
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		got, err := fn(value.IntValue(2), value.IntValue(3))
		if err != nil {
			t.Fatal(err)
		}
		i, _ := got.ToInt64()
		if i != 5 {
			t.Fatalf("expected 5, got %d", i)
		}
	})
}

func TestScalar3(t *testing.T) {
	t.Parallel()

	concat := func(a, b, c value.Value) (value.Value, error) {
		sa, _ := a.ToString()
		sb, _ := b.ToString()
		sc, _ := c.ToString()
		return value.StringValue(sa + sb + sc), nil
	}
	fn := helper.Scalar3(concat)

	if _, err := fn(value.StringValue("x")); err == nil {
		t.Fatal("expected arity error")
	}
	if got, err := fn(value.StringValue("x"), nil, value.StringValue("z")); err != nil || got != nil {
		t.Fatalf("expected (nil,nil) for middle NULL, got (%v,%v)", got, err)
	}
	if got, err := fn(value.StringValue("a"), value.StringValue("b"), value.StringValue("c")); err != nil {
		t.Fatal(err)
	} else {
		s, _ := got.ToString()
		if s != "abc" {
			t.Fatalf("expected abc, got %s", s)
		}
	}
}

func TestScalarN(t *testing.T) {
	t.Parallel()

	concat := func(args ...value.Value) (value.Value, error) {
		var out string
		for _, a := range args {
			s, _ := a.ToString()
			out += s
		}
		return value.StringValue(out), nil
	}
	fn := helper.ScalarN(concat)

	// 0 args is acceptable (ScalarN never enforces arity).
	if got, err := fn(); err != nil {
		t.Fatal(err)
	} else if s, _ := got.ToString(); s != "" {
		t.Fatalf("expected empty string, got %q", s)
	}
	// Any nil arg propagates NULL.
	if got, err := fn(value.StringValue("a"), nil, value.StringValue("c")); err != nil || got != nil {
		t.Fatalf("expected (nil,nil), got (%v,%v)", got, err)
	}
	if got, err := fn(value.StringValue("a"), value.StringValue("b")); err != nil {
		t.Fatal(err)
	} else if s, _ := got.ToString(); s != "ab" {
		t.Fatalf("expected ab, got %s", s)
	}
}

func TestScalar1KeepNull(t *testing.T) {
	t.Parallel()

	// IS_NULL-style: must see NULL.
	fn := helper.Scalar1KeepNull(func(v value.Value) (value.Value, error) {
		return value.BoolValue(v == nil), nil
	})
	if _, err := fn(); err == nil {
		t.Fatal("expected arity error")
	}
	if got, err := fn(nil); err != nil {
		t.Fatal(err)
	} else if b, _ := got.ToBool(); !b {
		t.Fatalf("expected TRUE for NULL input")
	}
	if got, err := fn(value.IntValue(1)); err != nil {
		t.Fatal(err)
	} else if b, _ := got.ToBool(); b {
		t.Fatalf("expected FALSE for non-NULL input")
	}
}

func TestScalar2KeepNull(t *testing.T) {
	t.Parallel()

	// IS_DISTINCT_FROM-style: must see both args, even when NULL.
	fn := helper.Scalar2KeepNull(func(a, b value.Value) (value.Value, error) {
		return value.BoolValue(a == nil && b == nil), nil
	})
	if _, err := fn(value.IntValue(1)); err == nil {
		t.Fatal("expected arity error")
	}
	got, err := fn(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBool(); !b {
		t.Fatalf("expected TRUE for (NULL,NULL)")
	}
}

func TestScalarNKeepNull(t *testing.T) {
	t.Parallel()

	// COALESCE-style: must observe NULL slots itself.
	fn := helper.ScalarNKeepNull(func(args ...value.Value) (value.Value, error) {
		for _, a := range args {
			if a != nil {
				return a, nil
			}
		}
		return nil, nil
	})
	got, err := fn(nil, nil, value.IntValue(3))
	if err != nil {
		t.Fatal(err)
	}
	if i, _ := got.ToInt64(); i != 3 {
		t.Fatalf("expected 3, got %d", i)
	}
	if got, err := fn(nil, nil); err != nil || got != nil {
		t.Fatalf("expected (nil,nil), got (%v,%v)", got, err)
	}
}

func TestExistsNull(t *testing.T) {
	t.Parallel()

	if helper.ExistsNull(nil) {
		t.Fatal("empty slice should not contain NULL")
	}
	if helper.ExistsNull([]value.Value{value.IntValue(1)}) {
		t.Fatal("non-NULL only must report false")
	}
	if !helper.ExistsNull([]value.Value{value.IntValue(1), nil}) {
		t.Fatal("slice with NULL must report true")
	}
}

// --------------------------------------------------------------------
// Option / ParseOptions tests
// --------------------------------------------------------------------

func TestParseOptions_NoMarkers(t *testing.T) {
	t.Parallel()

	args := []value.Value{value.IntValue(1), nil, value.StringValue("x")}
	out, opt := helper.ParseOptions(args...)
	if len(out) != 3 {
		t.Fatalf("expected 3 passthrough args, got %d", len(out))
	}
	if opt.Distinct || opt.IgnoreNulls || opt.Limit != nil || len(opt.OrderBy) != 0 {
		t.Fatalf("expected empty option, got %+v", opt)
	}
}

func TestParseOptions_AllMarkers(t *testing.T) {
	t.Parallel()

	distinct, err := helper.DISTINCT()
	if err != nil {
		t.Fatal(err)
	}
	ignoreNulls, err := helper.IGNORE_NULLS()
	if err != nil {
		t.Fatal(err)
	}
	limit, err := helper.LIMIT(7)
	if err != nil {
		t.Fatal(err)
	}
	orderBy, err := helper.ORDER_BY(value.IntValue(3), true)
	if err != nil {
		t.Fatal(err)
	}

	args := []value.Value{
		value.IntValue(1),
		distinct,
		ignoreNulls,
		limit,
		orderBy,
	}
	out, opt := helper.ParseOptions(args...)
	if len(out) != 1 {
		t.Fatalf("expected single non-marker arg, got %d", len(out))
	}
	if !opt.Distinct {
		t.Error("Distinct not set")
	}
	if !opt.IgnoreNulls {
		t.Error("IgnoreNulls not set")
	}
	if opt.Limit == nil || *opt.Limit != 7 {
		t.Errorf("Limit not 7, got %v", opt.Limit)
	}
	if len(opt.OrderBy) != 1 || !opt.OrderBy[0].IsAsc {
		t.Errorf("OrderBy not captured, got %+v", opt.OrderBy)
	}
	if v, _ := opt.OrderBy[0].Value.ToInt64(); v != 3 {
		t.Errorf("OrderBy value mismatch: %d", v)
	}
}

// --------------------------------------------------------------------
// Aggregator (Step / Done) tests
// --------------------------------------------------------------------

func TestAggregator_StepAndDone(t *testing.T) {
	t.Parallel()

	var sum int64
	agg := helper.NewAggregator(
		func(vs []value.Value, _ *helper.Option) error {
			for _, v := range vs {
				if v == nil {
					continue
				}
				i, err := v.ToInt64()
				if err != nil {
					return err
				}
				sum += i
			}
			return nil
		},
		func() (value.Value, error) {
			return value.IntValue(sum), nil
		},
	)
	if err := agg.Step(int64(2)); err != nil {
		t.Fatal(err)
	}
	if err := agg.Step(int64(3)); err != nil {
		t.Fatal(err)
	}
	out, err := agg.Done()
	if err != nil {
		t.Fatal(err)
	}
	// EncodeValue returns Go-side representation (int64 or string).
	if got, ok := out.(int64); !ok || got != 5 {
		t.Fatalf("expected 5, got %v (%T)", out, out)
	}
}

func TestAggregator_IgnoreNullsAndDistinct(t *testing.T) {
	t.Parallel()

	var captured []value.Value
	mkAgg := func() *helper.Aggregator {
		captured = nil
		return helper.NewAggregator(
			func(vs []value.Value, _ *helper.Option) error {
				captured = append(captured, vs...)
				return nil
			},
			func() (value.Value, error) { return value.IntValue(int64(len(captured))), nil },
		)
	}

	// IGNORE_NULLS: NULL slot is dropped before step is invoked.
	ignoreNullsArg, err := helper.IGNORE_NULLS()
	if err != nil {
		t.Fatal(err)
	}
	ignoreNullsArgStr, _ := ignoreNullsArg.ToString()
	a := mkAgg()
	if err := a.Step(nil, ignoreNullsArgStr); err != nil {
		t.Fatal(err)
	}
	// captured must be empty because NULL got filtered and only-marker
	// step is skipped (len(values)==0 after filter).
	if len(captured) != 0 {
		t.Fatalf("expected empty captured under IGNORE_NULLS for NULL arg, got %v", captured)
	}

	// DISTINCT: a repeated key collapses to one Step call.
	distinctArg, err := helper.DISTINCT()
	if err != nil {
		t.Fatal(err)
	}
	distinctArgStr, _ := distinctArg.ToString()
	a = mkAgg()
	if err := a.Step(int64(1), distinctArgStr); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(1), distinctArgStr); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(2), distinctArgStr); err != nil {
		t.Fatal(err)
	}
	if len(captured) != 2 {
		t.Fatalf("expected 2 distinct values, got %d (%v)", len(captured), captured)
	}
}

// --------------------------------------------------------------------
// SortAggregatedValues tests
// --------------------------------------------------------------------

func TestSortAggregatedValues_Ascending(t *testing.T) {
	t.Parallel()

	values := []*helper.OrderedValue{
		{OrderBy: []*helper.OrderBy{{Value: value.IntValue(3), IsAsc: true}}, Value: value.IntValue(30)},
		{OrderBy: []*helper.OrderBy{{Value: value.IntValue(1), IsAsc: true}}, Value: value.IntValue(10)},
		{OrderBy: []*helper.OrderBy{{Value: value.IntValue(2), IsAsc: true}}, Value: value.IntValue(20)},
	}
	got := helper.SortAggregatedValues(values, &helper.Option{
		OrderBy: []*helper.OrderBy{{IsAsc: true}},
	})
	if v, _ := got[0].Value.ToInt64(); v != 10 {
		t.Fatalf("expected 10 first, got %d", v)
	}
	if v, _ := got[2].Value.ToInt64(); v != 30 {
		t.Fatalf("expected 30 last, got %d", v)
	}
}

func TestSortAggregatedValues_NoOrderBy(t *testing.T) {
	t.Parallel()

	values := []*helper.OrderedValue{
		{Value: value.IntValue(2)},
		{Value: value.IntValue(1)},
	}
	got := helper.SortAggregatedValues(values, &helper.Option{})
	// With no OrderBy in Option, the function must not reorder.
	if v, _ := got[0].Value.ToInt64(); v != 2 {
		t.Fatalf("expected stable order (2,1), got %d at position 0", v)
	}
}

// --------------------------------------------------------------------
// Date / month / year helpers
// --------------------------------------------------------------------

func TestAddMonth(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		input    [3]int // year, month, day
		months   int
		expected [3]int
	}{
		// Mirrors the BigQuery DATE_ADD docs example: 2017-01-31 +
		// INTERVAL 1 MONTH -> 2017-02-28 (clamped to last day of Feb).
		{"clamp_end_of_month", [3]int{2017, 1, 31}, 1, [3]int{2017, 2, 28}},
		// Plain non-clamping case.
		{"simple_forward", [3]int{2020, 1, 15}, 2, [3]int{2020, 3, 15}},
		// Backwards across year boundary.
		{"simple_backward_across_year", [3]int{2020, 2, 15}, -3, [3]int{2019, 11, 15}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in := mkUTC(tc.input[0], tc.input[1], tc.input[2])
			got := helper.AddMonth(in, tc.months)
			want := mkUTC(tc.expected[0], tc.expected[1], tc.expected[2])
			if !got.Equal(want) {
				t.Fatalf("AddMonth(%v, %d) = %v, want %v", in, tc.months, got, want)
			}
		})
	}
}

func TestAddYear(t *testing.T) {
	t.Parallel()

	// Per the BigQuery DATE_ADD INTERVAL ... YEAR examples and the
	// shared DATE_ADD reference: 2016-02-29 + 1 YEAR -> 2017-02-28.
	in := mkUTC(2016, 2, 29)
	got := helper.AddYear(in, 1)
	want := mkUTC(2017, 2, 28)
	if !got.Equal(want) {
		t.Fatalf("AddYear leap clamp: got %v want %v", got, want)
	}

	// Plain non-clamping case.
	in = mkUTC(2020, 6, 15)
	got = helper.AddYear(in, 3)
	want = mkUTC(2023, 6, 15)
	if !got.Equal(want) {
		t.Fatalf("AddYear simple: got %v want %v", got, want)
	}
}

func TestWeekPartToOffsetAndQuarterStartMonths(t *testing.T) {
	t.Parallel()

	if helper.WeekPartToOffset["WEEK"] != 0 {
		t.Errorf("WEEK offset must be 0")
	}
	if helper.WeekPartToOffset["WEEK_SATURDAY"] != 6 {
		t.Errorf("WEEK_SATURDAY offset must be 6")
	}
	if len(helper.QuarterStartMonths) != 4 {
		t.Errorf("expected 4 quarter months, got %d", len(helper.QuarterStartMonths))
	}
}

// --------------------------------------------------------------------
// BucketFloor tests (DATE_BUCKET / DATETIME_BUCKET / TIMESTAMP_BUCKET)
// --------------------------------------------------------------------

// makeInterval mirrors the intervalvalue.IntervalValue shape we construct
// from analyzer-side INTERVAL literals.
func makeInterval(y, m, d, h, mi, s, ns int32) *value.IntervalValue {
	return &value.IntervalValue{
		IntervalValue: &intervalvalue.IntervalValue{
			Years:          y,
			Months:         m,
			Days:           d,
			Hours:          h,
			Minutes:        mi,
			Seconds:        s,
			SubSecondNanos: ns,
		},
	}
}

func TestBucketFloor_DaysWidth(t *testing.T) {
	t.Parallel()

	// From DATE_BUCKET upstream docs: DATE_BUCKET('2020-04-19',
	// INTERVAL 7 DAY, '2020-01-01') -> 2020-04-16 (start of the
	// 16-day-aligned 7-day bucket).
	target := mkUTC(2020, 4, 19)
	origin := mkUTC(2020, 1, 1)
	iv := makeInterval(0, 0, 7, 0, 0, 0, 0)
	got, err := helper.BucketFloor(target, origin, iv)
	if err != nil {
		t.Fatal(err)
	}
	want := mkUTC(2020, 4, 15)
	if !got.Equal(want) {
		t.Fatalf("BucketFloor 7-day: got %v want %v", got, want)
	}
}

func TestBucketFloor_MonthsWidth(t *testing.T) {
	t.Parallel()

	// DATETIME_BUCKET upstream docs:
	// DATETIME_BUCKET('2020-04-19 13:45:00', INTERVAL 1 MONTH,
	//                 '2020-01-01 00:00:00') -> 2020-04-01 00:00:00.
	target := mkUTC(2020, 4, 19)
	origin := mkUTC(2020, 1, 1)
	iv := makeInterval(0, 1, 0, 0, 0, 0, 0)
	got, err := helper.BucketFloor(target, origin, iv)
	if err != nil {
		t.Fatal(err)
	}
	want := mkUTC(2020, 4, 1)
	if !got.Equal(want) {
		t.Fatalf("BucketFloor 1-month: got %v want %v", got, want)
	}
}

func TestBucketFloor_NilInterval(t *testing.T) {
	t.Parallel()

	if _, err := helper.BucketFloor(mkUTC(2020, 1, 1), mkUTC(2020, 1, 1), nil); err == nil {
		t.Fatal("expected error for nil interval")
	}
	// IntervalValue itself is non-nil but the embedded pointer is.
	bad := &value.IntervalValue{}
	if _, err := helper.BucketFloor(mkUTC(2020, 1, 1), mkUTC(2020, 1, 1), bad); err == nil {
		t.Fatal("expected error for nil embedded interval")
	}
}

func TestBucketFloor_MixedParts(t *testing.T) {
	t.Parallel()

	// Mixing month + day is not supported per the implementation.
	iv := makeInterval(0, 1, 1, 0, 0, 0, 0)
	if _, err := helper.BucketFloor(mkUTC(2020, 1, 1), mkUTC(2020, 1, 1), iv); err == nil {
		t.Fatal("expected error for mixed month + day interval")
	}
}

func TestBucketFloor_NegativeMonths(t *testing.T) {
	t.Parallel()

	iv := makeInterval(0, -1, 0, 0, 0, 0, 0)
	if _, err := helper.BucketFloor(mkUTC(2020, 1, 1), mkUTC(2020, 1, 1), iv); err == nil {
		t.Fatal("expected error for negative months")
	}
}

func TestBucketFloor_ZeroWidth(t *testing.T) {
	t.Parallel()

	iv := makeInterval(0, 0, 0, 0, 0, 0, 0)
	if _, err := helper.BucketFloor(mkUTC(2020, 1, 1), mkUTC(2020, 1, 1), iv); err == nil {
		t.Fatal("expected error for zero-width bucket")
	}
}

// Helpers ------------------------------------------------------------

func mkUTC(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}
