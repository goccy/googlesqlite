package value_test

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/intervalvalue"
	"github.com/goccy/googlesqlite/internal/value"
)

// TestTimestampValue covers TimestampValue arithmetic, part-shift, and
// conversions.
func TestTimestampValue(t *testing.T) {
	t.Parallel()

	base := value.TimestampValue(time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC))

	t.Run("Add IntervalValue", func(t *testing.T) {
		iv := &value.IntervalValue{IntervalValue: &intervalvalue.IntervalValue{Hours: 2}}
		got, err := base.Add(iv)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.TimestampValue); !ok {
			t.Fatalf("expected TimestampValue, got %T", got)
		}
	})

	t.Run("Add unsupported", func(t *testing.T) {
		if _, err := base.Add(value.IntValue(1)); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Sub IntervalValue", func(t *testing.T) {
		iv := &value.IntervalValue{IntervalValue: &intervalvalue.IntervalValue{Hours: 1}}
		got, err := base.Sub(iv)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.TimestampValue); !ok {
			t.Fatalf("expected TimestampValue, got %T", got)
		}
	})

	t.Run("Sub TimestampValue returns interval", func(t *testing.T) {
		earlier := value.TimestampValue(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC))
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
		earlier := value.TimestampValue(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC))
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
		if _, err := base.ToString(); err != nil {
			t.Fatal(err)
		}
		if _, err := base.ToBytes(); err != nil {
			t.Fatal(err)
		}
		if _, err := base.ToJSON(); err != nil {
			t.Fatal(err)
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

	t.Run("Format/Interface", func(t *testing.T) {
		if got := base.Format('t'); got != "2020-01-01 12:30:45+00" {
			t.Fatalf("Format t: %s", got)
		}
		if got := base.Format('T'); got != `TIMESTAMP "2020-01-01 12:30:45+00"` {
			t.Fatalf("Format T: %s", got)
		}
		if got := base.Format('x'); got != "2020-01-01 12:30:45+00" {
			t.Fatalf("Format default: %s", got)
		}
		if got, ok := base.Interface().(string); !ok || got == "" {
			t.Fatalf("Interface: %v", base.Interface())
		}
	})

	t.Run("AddValueWithPart", func(t *testing.T) {
		cases := []struct {
			part string
		}{
			{"MICROSECOND"},
			{"MILLISECOND"},
			{"SECOND"},
			{"MINUTE"},
			{"HOUR"},
			{"DAY"},
		}
		for _, c := range cases {
			t.Run(c.part, func(t *testing.T) {
				got, err := base.AddValueWithPart(1, c.part)
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := got.(value.TimestampValue); !ok {
					t.Fatalf("%s -> %T", c.part, got)
				}
			})
		}
		if _, err := base.AddValueWithPart(1, "BOGUS"); err == nil {
			t.Fatal("expected error for unknown part")
		}
	})
}
