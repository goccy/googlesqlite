package value_test

import (
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/goccy/googlesqlite/internal/value"
)

// TestDatetimeValue covers DatetimeValue add/sub semantics, comparison,
// and conversions.
func TestDatetimeValue(t *testing.T) {
	t.Parallel()

	base := value.DatetimeValue(time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC))

	t.Run("Add IntervalValue", func(t *testing.T) {
		iv := &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Hours: 1}}
		got, err := base.Add(iv)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.DatetimeValue); !ok {
			t.Fatalf("expected DatetimeValue, got %T", got)
		}
	})

	t.Run("Add unsupported rhs type", func(t *testing.T) {
		if _, err := base.Add(value.IntValue(1)); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Sub IntervalValue", func(t *testing.T) {
		iv := &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Hours: 1}}
		got, err := base.Sub(iv)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.DatetimeValue); !ok {
			t.Fatalf("expected DatetimeValue, got %T", got)
		}
	})

	t.Run("Sub of DatetimeValue returns interval", func(t *testing.T) {
		earlier := value.DatetimeValue(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC))
		got, err := base.Sub(earlier)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(*value.IntervalValue); !ok {
			t.Fatalf("expected IntervalValue, got %T", got)
		}
	})

	t.Run("Mul/Div unsupported", func(t *testing.T) {
		if _, err := base.Mul(base); err == nil {
			t.Fatal("Mul")
		}
		if _, err := base.Div(base); err == nil {
			t.Fatal("Div")
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		earlier := value.DatetimeValue(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC))
		if ok, _ := base.EQ(base); !ok {
			t.Fatal("EQ")
		}
		if ok, _ := base.GT(earlier); !ok {
			t.Fatal("GT")
		}
		if ok, _ := base.GTE(earlier); !ok {
			t.Fatal("GTE")
		}
		if ok, _ := earlier.LT(base); !ok {
			t.Fatal("LT")
		}
		if ok, _ := earlier.LTE(base); !ok {
			t.Fatal("LTE")
		}
	})

	t.Run("conversions", func(t *testing.T) {
		if _, err := base.ToInt64(); err != nil {
			t.Fatal(err)
		}
		if _, err := base.ToFloat64(); err != nil {
			t.Fatal(err)
		}
		s, _ := base.ToString()
		if s != "2020-01-01T12:30:45" {
			t.Fatalf("ToString: %s", s)
		}
		if got, _ := base.ToBytes(); string(got) != "2020-01-01T12:30:45" {
			t.Fatalf("ToBytes: %s", got)
		}
		if got, _ := base.ToJSON(); got != "2020-01-01T12:30:45" {
			t.Fatalf("ToJSON: %s", got)
		}
		if got, _ := base.ToTime(); got.IsZero() {
			t.Fatal("ToTime")
		}
	})

	t.Run("ToBool/ToArray/ToStruct/ToRat error", func(t *testing.T) {
		if _, err := base.ToBool(); err == nil {
			t.Fatal("ToBool")
		}
		if _, err := base.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := base.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
		if _, err := base.ToRat(); err == nil {
			t.Fatal("ToRat")
		}
	})

	t.Run("Format and Interface", func(t *testing.T) {
		if base.Format('t') != "2020-01-01T12:30:45" {
			t.Fatalf("Format t: %s", base.Format('t'))
		}
		if base.Format('T') != `DATETIME "2020-01-01T12:30:45"` {
			t.Fatalf("Format T: %s", base.Format('T'))
		}
		if base.Format('x') != "2020-01-01T12:30:45" {
			t.Fatalf("Format default: %s", base.Format('x'))
		}
		if got, ok := base.Interface().(string); !ok || got != "2020-01-01T12:30:45" {
			t.Fatalf("Interface: %v", base.Interface())
		}
	})
}
