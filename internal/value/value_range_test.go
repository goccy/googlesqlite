package value_test

import (
	"strings"
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestRangeValue covers RangeValue equality (incl. nil-bound flips),
// the all-unsupported scalar comparisons, ToString rendering for the
// supported element kinds (DATE / DATETIME / TIMESTAMP), and the
// conversion error stubs.
func TestRangeValue(t *testing.T) {
	t.Parallel()

	d := func(y, m, day int) value.DateValue {
		return value.DateValue(time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC))
	}

	t.Run("EQ both bounded equal", func(t *testing.T) {
		a := &value.RangeValue{Start: d(2020, 1, 1), End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		b := &value.RangeValue{Start: d(2020, 1, 1), End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		ok, err := a.EQ(b)
		if err != nil || !ok {
			t.Fatalf("EQ both bounded: %v / err=%v", ok, err)
		}
	})

	t.Run("EQ start-nil mismatch", func(t *testing.T) {
		a := &value.RangeValue{Start: nil, End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		b := &value.RangeValue{Start: d(2020, 1, 1), End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		ok, err := a.EQ(b)
		if err != nil || ok {
			t.Fatalf("expected not equal")
		}
	})

	t.Run("EQ end-nil mismatch", func(t *testing.T) {
		a := &value.RangeValue{Start: d(2020, 1, 1), End: nil, ElemHeader: value.DateValueType}
		b := &value.RangeValue{Start: d(2020, 1, 1), End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		ok, err := a.EQ(b)
		if err != nil || ok {
			t.Fatalf("expected not equal")
		}
	})

	t.Run("EQ both unbounded", func(t *testing.T) {
		a := &value.RangeValue{ElemHeader: value.DateValueType}
		b := &value.RangeValue{ElemHeader: value.DateValueType}
		ok, err := a.EQ(b)
		if err != nil || !ok {
			t.Fatalf("unbounded EQ unbounded: %v / err=%v", ok, err)
		}
	})

	t.Run("EQ wrong type rhs", func(t *testing.T) {
		a := &value.RangeValue{ElemHeader: value.DateValueType}
		if _, err := a.EQ(value.IntValue(1)); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("comparisons unsupported", func(t *testing.T) {
		r := &value.RangeValue{}
		if _, err := r.GT(r); err == nil {
			t.Fatal("GT")
		}
		if _, err := r.GTE(r); err == nil {
			t.Fatal("GTE")
		}
		if _, err := r.LT(r); err == nil {
			t.Fatal("LT")
		}
		if _, err := r.LTE(r); err == nil {
			t.Fatal("LTE")
		}
		if _, err := r.Add(r); err == nil {
			t.Fatal("Add")
		}
		if _, err := r.Sub(r); err == nil {
			t.Fatal("Sub")
		}
		if _, err := r.Mul(r); err == nil {
			t.Fatal("Mul")
		}
		if _, err := r.Div(r); err == nil {
			t.Fatal("Div")
		}
	})

	t.Run("ToString DATE", func(t *testing.T) {
		r := &value.RangeValue{Start: d(2020, 1, 1), End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		s, err := r.ToString()
		if err != nil {
			t.Fatal(err)
		}
		if s != "[2020-01-01, 2020-01-05)" {
			t.Fatalf("ToString DATE: %s", s)
		}
	})

	t.Run("ToString DATETIME uses space separator", func(t *testing.T) {
		dt := value.DatetimeValue(time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC))
		r := &value.RangeValue{Start: dt, End: nil, ElemHeader: value.DatetimeValueType}
		s, _ := r.ToString()
		if !strings.Contains(s, "2020-01-01 12:30:45") {
			t.Fatalf("ToString DATETIME: %s", s)
		}
		if !strings.HasSuffix(s, "UNBOUNDED)") {
			t.Fatalf("expected UNBOUNDED end: %s", s)
		}
	})

	t.Run("ToString TIMESTAMP", func(t *testing.T) {
		ts := value.TimestampValue(time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC))
		r := &value.RangeValue{Start: nil, End: ts, ElemHeader: value.TimestampValueType}
		s, _ := r.ToString()
		if !strings.HasPrefix(s, "[UNBOUNDED, ") {
			t.Fatalf("expected UNBOUNDED start: %s", s)
		}
		if !strings.Contains(s, "+00") {
			t.Fatalf("expected +00 in TIMESTAMP bound: %s", s)
		}
	})

	t.Run("ToBytes/ToJSON", func(t *testing.T) {
		r := &value.RangeValue{Start: d(2020, 1, 1), End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		b, _ := r.ToBytes()
		if string(b) != "[2020-01-01, 2020-01-05)" {
			t.Fatalf("ToBytes: %s", b)
		}
		j, _ := r.ToJSON()
		if !strings.HasPrefix(j, `"`) || !strings.HasSuffix(j, `"`) {
			t.Fatalf("ToJSON unquoted: %s", j)
		}
	})

	t.Run("conversion errors", func(t *testing.T) {
		r := &value.RangeValue{}
		if _, err := r.ToInt64(); err == nil {
			t.Fatal("ToInt64")
		}
		if _, err := r.ToFloat64(); err == nil {
			t.Fatal("ToFloat64")
		}
		if _, err := r.ToBool(); err == nil {
			t.Fatal("ToBool")
		}
		if _, err := r.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := r.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
		if _, err := r.ToTime(); err == nil {
			t.Fatal("ToTime")
		}
		if _, err := r.ToRat(); err == nil {
			t.Fatal("ToRat")
		}
	})

	t.Run("Format/Interface", func(t *testing.T) {
		r := &value.RangeValue{Start: d(2020, 1, 1), End: d(2020, 1, 5), ElemHeader: value.DateValueType}
		if got := r.Format('t'); got != "[2020-01-01, 2020-01-05)" {
			t.Fatalf("Format t: %s", got)
		}
		if got := r.Format('T'); got != `RANGE "[2020-01-01, 2020-01-05)"` {
			t.Fatalf("Format T: %s", got)
		}
		if r.Interface() == nil {
			t.Fatal("Interface nil")
		}
	})
}
