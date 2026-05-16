package value_test

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestTimeValue covers TimeValue conversions and unsupported arithmetic.
func TestTimeValue(t *testing.T) {
	t.Parallel()

	tv := value.TimeValue(time.Date(0, 1, 1, 12, 30, 45, 0, time.UTC))

	t.Run("Add/Sub/Mul/Div unsupported", func(t *testing.T) {
		if _, err := tv.Add(tv); err == nil {
			t.Fatal("Add")
		}
		if _, err := tv.Sub(tv); err == nil {
			t.Fatal("Sub")
		}
		if _, err := tv.Mul(tv); err == nil {
			t.Fatal("Mul")
		}
		if _, err := tv.Div(tv); err == nil {
			t.Fatal("Div")
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		earlier := value.TimeValue(time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC))
		if ok, _ := tv.EQ(tv); !ok {
			t.Fatal("EQ")
		}
		if ok, _ := tv.GT(earlier); !ok {
			t.Fatal("GT")
		}
		if ok, _ := tv.GTE(earlier); !ok {
			t.Fatal("GTE")
		}
		if ok, _ := earlier.LT(tv); !ok {
			t.Fatal("LT")
		}
		if ok, _ := earlier.LTE(tv); !ok {
			t.Fatal("LTE")
		}
	})

	t.Run("conversions", func(t *testing.T) {
		s, _ := tv.ToString()
		if s != "12:30:45" {
			t.Fatalf("ToString: %s", s)
		}
		if got, _ := tv.ToBytes(); string(got) != "12:30:45" {
			t.Fatalf("ToBytes: %s", got)
		}
		if got, _ := tv.ToJSON(); got != "12:30:45" {
			t.Fatalf("ToJSON: %s", got)
		}
		if got, _ := tv.ToTime(); got.IsZero() {
			t.Fatal("ToTime")
		}
		if _, err := tv.ToInt64(); err != nil {
			t.Fatal(err)
		}
		if _, err := tv.ToFloat64(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ToBool/ToArray/ToStruct/ToRat error", func(t *testing.T) {
		if _, err := tv.ToBool(); err == nil {
			t.Fatal("ToBool")
		}
		if _, err := tv.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := tv.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
		if _, err := tv.ToRat(); err == nil {
			t.Fatal("ToRat")
		}
	})

	t.Run("Format/Interface", func(t *testing.T) {
		if tv.Format('t') != "12:30:45" {
			t.Fatalf("Format t: %s", tv.Format('t'))
		}
		if tv.Format('T') != `TIME "12:30:45"` {
			t.Fatalf("Format T: %s", tv.Format('T'))
		}
		if tv.Format('x') != "12:30:45" {
			t.Fatalf("Format default: %s", tv.Format('x'))
		}
		if got, ok := tv.Interface().(string); !ok || got != "12:30:45" {
			t.Fatalf("Interface: %v", tv.Interface())
		}
	})
}
