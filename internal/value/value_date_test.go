package value_test

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/intervalvalue"
	"github.com/goccy/googlesqlite/internal/value"
)

// TestDateValue covers DateValue add/sub semantics, interval handling,
// comparison, and conversions.
func TestDateValue(t *testing.T) {
	t.Parallel()

	base := value.DateValue(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))

	t.Run("Add IntValue advances days", func(t *testing.T) {
		got, err := base.Add(value.IntValue(3))
		if err != nil {
			t.Fatal(err)
		}
		out, _ := got.ToString()
		if out != "2020-01-04" {
			t.Fatalf("Add: %s", out)
		}
	})

	t.Run("Sub IntValue retreats days", func(t *testing.T) {
		got, err := base.Sub(value.IntValue(1))
		if err != nil {
			t.Fatal(err)
		}
		out, _ := got.ToString()
		if out != "2019-12-31" {
			t.Fatalf("Sub: %s", out)
		}
	})

	t.Run("Add IntervalValue produces Datetime", func(t *testing.T) {
		iv := &value.IntervalValue{IntervalValue: &intervalvalue.IntervalValue{Years: 1, Months: 2, Days: 3}}
		got, err := base.Add(iv)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.DatetimeValue); !ok {
			t.Fatalf("expected DatetimeValue, got %T", got)
		}
	})

	t.Run("Sub IntervalValue produces Datetime", func(t *testing.T) {
		iv := &value.IntervalValue{IntervalValue: &intervalvalue.IntervalValue{Days: 1}}
		got, err := base.Sub(iv)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.DatetimeValue); !ok {
			t.Fatalf("expected DatetimeValue, got %T", got)
		}
	})

	t.Run("Sub of DateValue returns interval days", func(t *testing.T) {
		earlier := value.DateValue(time.Date(2019, 12, 25, 0, 0, 0, 0, time.UTC))
		got, err := base.Sub(earlier)
		if err != nil {
			t.Fatal(err)
		}
		iv, ok := got.(*value.IntervalValue)
		if !ok {
			t.Fatalf("expected IntervalValue, got %T", got)
		}
		if iv.Days != 7 {
			t.Fatalf("expected 7 days, got %d", iv.Days)
		}
	})

	t.Run("Add unsupported rhs type", func(t *testing.T) {
		if _, err := base.Add(value.StringValue("x")); err == nil {
			t.Fatal("expected error")
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

	t.Run("EQ/GT/GTE/LT/LTE", func(t *testing.T) {
		earlier := value.DateValue(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
		if ok, _ := base.EQ(base); !ok {
			t.Fatal("EQ self")
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

	t.Run("ToInt64/Float64/String/Bytes/JSON/Time", func(t *testing.T) {
		if got, _ := base.ToString(); got != "2020-01-01" {
			t.Fatalf("ToString: %s", got)
		}
		if got, _ := base.ToBytes(); string(got) != "2020-01-01" {
			t.Fatalf("ToBytes: %s", got)
		}
		if got, _ := base.ToJSON(); got != "2020-01-01" {
			t.Fatalf("ToJSON: %s", got)
		}
		if got, _ := base.ToTime(); got.IsZero() {
			t.Fatal("ToTime")
		}
		if _, err := base.ToInt64(); err != nil {
			t.Fatalf("ToInt64: %v", err)
		}
		if _, err := base.ToFloat64(); err != nil {
			t.Fatalf("ToFloat64: %v", err)
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
		if base.Format('t') != "2020-01-01" {
			t.Fatalf("Format t: %s", base.Format('t'))
		}
		if base.Format('T') != `DATE "2020-01-01"` {
			t.Fatalf("Format T: %s", base.Format('T'))
		}
		if base.Format('x') != "2020-01-01" {
			t.Fatalf("Format default: %s", base.Format('x'))
		}
		if got, ok := base.Interface().(string); !ok || got != "2020-01-01" {
			t.Fatalf("Interface: %v", base.Interface())
		}
	})

	t.Run("AddDateWithInterval", func(t *testing.T) {
		cases := []struct {
			interval string
			want     string
		}{
			{"WEEK", "2020-01-08"},
			{"MONTH", "2020-02-01"},
			{"YEAR", "2021-01-01"},
			{"DAY", "2020-01-02"}, // default branch
		}
		for _, c := range cases {
			t.Run(c.interval, func(t *testing.T) {
				got, err := base.AddDateWithInterval(1, c.interval)
				if err != nil {
					t.Fatal(err)
				}
				out, _ := got.(value.DateValue).ToString()
				if out != c.want {
					t.Fatalf("AddDateWithInterval(%s): got %s, want %s", c.interval, out, c.want)
				}
			})
		}
	})
}
