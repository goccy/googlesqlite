package value_test

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestErrorBranches drives the rhs-conversion-fails branches across
// every Value type that delegates to the rhs's To<X> method. Each
// subtest deliberately picks a rhs whose conversion to the relevant
// scalar fails, exercising the matching error path in the lhs's
// comparison method.
func TestErrorBranches(t *testing.T) {
	t.Parallel()

	// ArrayValue doesn't support most conversions, so it works as a
	// universal "this conversion will fail" rhs.
	bad := &value.ArrayValue{Values: []value.Value{value.IntValue(1)}}

	t.Run("ArrayValue comparisons error on element conversion", func(t *testing.T) {
		// Make element-wise EQ fail by having the lhs element be an
		// ArrayValue (which can't EQ a scalar without ToArray erroring).
		av := &value.ArrayValue{Values: []value.Value{
			&value.ArrayValue{Values: []value.Value{value.IntValue(1)}},
		}}
		mismatch := &value.ArrayValue{Values: []value.Value{value.IntValue(1)}}
		if _, err := av.EQ(mismatch); err == nil {
			t.Fatal("EQ should err (element ToArray fail)")
		}
		// rhs whose ToArray errors propagates.
		if _, err := av.EQ(value.IntValue(0)); err == nil {
			t.Fatal("EQ rhs ToArray err")
		}
	})

	t.Run("DateValue comparison errors", func(t *testing.T) {
		dv := value.DateValue(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
		if _, err := dv.EQ(bad); err == nil {
			t.Fatal("EQ")
		}
		if _, err := dv.GT(bad); err == nil {
			t.Fatal("GT")
		}
		if _, err := dv.GTE(bad); err == nil {
			t.Fatal("GTE")
		}
		if _, err := dv.LT(bad); err == nil {
			t.Fatal("LT")
		}
		if _, err := dv.LTE(bad); err == nil {
			t.Fatal("LTE")
		}
		if _, err := dv.Sub(bad); err == nil {
			t.Fatal("Sub")
		}
	})

	t.Run("DatetimeValue comparison errors", func(t *testing.T) {
		dv := value.DatetimeValue(time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC))
		if _, err := dv.EQ(bad); err == nil {
			t.Fatal("EQ")
		}
		if _, err := dv.GT(bad); err == nil {
			t.Fatal("GT")
		}
		if _, err := dv.GTE(bad); err == nil {
			t.Fatal("GTE")
		}
		if _, err := dv.LT(bad); err == nil {
			t.Fatal("LT")
		}
		if _, err := dv.LTE(bad); err == nil {
			t.Fatal("LTE")
		}
		if _, err := dv.Sub(bad); err == nil {
			t.Fatal("Sub")
		}
	})

	t.Run("TimeValue comparison errors", func(t *testing.T) {
		tv := value.TimeValue(time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC))
		if _, err := tv.EQ(bad); err == nil {
			t.Fatal("EQ")
		}
		if _, err := tv.GT(bad); err == nil {
			t.Fatal("GT")
		}
		if _, err := tv.GTE(bad); err == nil {
			t.Fatal("GTE")
		}
		if _, err := tv.LT(bad); err == nil {
			t.Fatal("LT")
		}
		if _, err := tv.LTE(bad); err == nil {
			t.Fatal("LTE")
		}
	})

	t.Run("TimestampValue comparison errors", func(t *testing.T) {
		tv := value.TimestampValue(time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC))
		if _, err := tv.EQ(bad); err == nil {
			t.Fatal("EQ")
		}
		if _, err := tv.GT(bad); err == nil {
			t.Fatal("GT")
		}
		if _, err := tv.GTE(bad); err == nil {
			t.Fatal("GTE")
		}
		if _, err := tv.LT(bad); err == nil {
			t.Fatal("LT")
		}
		if _, err := tv.LTE(bad); err == nil {
			t.Fatal("LTE")
		}
		if _, err := tv.Sub(bad); err == nil {
			t.Fatal("Sub")
		}
	})

	t.Run("BytesValue comparison errors", func(t *testing.T) {
		bv := value.BytesValue([]byte("x"))
		// Need a rhs whose ToBytes fails; ArrayValue.ToBytes succeeds
		// (returns a serialised "[...]"). Use SafeValue wrapping a
		// failing inner so that's also not quite right — actually we
		// cannot easily force ToBytes to fail because most types have
		// fallback implementations. Skip the rhs-error branch.
		//
		// The Add branch concatenates after ToBytes, so use the success
		// path to at least exercise ToBytes on the rhs.
		if _, err := bv.Add(value.IntValue(7)); err != nil {
			t.Fatalf("Add: %v", err)
		}
	})

	t.Run("StructValue comparison errors on rhs ToStruct", func(t *testing.T) {
		sv := &value.StructValue{
			Keys:   []string{"a"},
			Values: []value.Value{value.IntValue(1)},
			M:      map[string]value.Value{"a": value.IntValue(1)},
		}
		// rhs whose ToStruct errors.
		if _, err := sv.EQ(value.IntValue(0)); err == nil {
			t.Fatal("EQ rhs ToStruct err")
		}
		if _, err := sv.GT(value.IntValue(0)); err == nil {
			t.Fatal("GT")
		}
		if _, err := sv.GTE(value.IntValue(0)); err == nil {
			t.Fatal("GTE")
		}
		if _, err := sv.LT(value.IntValue(0)); err == nil {
			t.Fatal("LT")
		}
		if _, err := sv.LTE(value.IntValue(0)); err == nil {
			t.Fatal("LTE")
		}
	})

	t.Run("StructValue comparison errors via structCompare element error", func(t *testing.T) {
		// Force structCompare to err: element EQ returns an error.
		// Inner element of lhs is an ArrayValue, which when EQ'd against
		// an IntValue errors.
		lhs := &value.StructValue{
			Keys:   []string{"a"},
			Values: []value.Value{&value.ArrayValue{Values: []value.Value{value.IntValue(1)}}},
			M:      map[string]value.Value{"a": &value.ArrayValue{}},
		}
		rhs := &value.StructValue{
			Keys:   []string{"a"},
			Values: []value.Value{value.StringValue("x")},
			M:      map[string]value.Value{"a": value.StringValue("x")},
		}
		if _, err := lhs.EQ(rhs); err == nil {
			t.Fatal("structCompare element err")
		}
	})
}
