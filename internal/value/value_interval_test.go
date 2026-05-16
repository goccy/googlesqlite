package value_test

import (
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/goccy/googlesqlite/internal/value"
)

// TestIntervalValue covers IntervalValue's all-unsupported arithmetic,
// the conversion stubs that return errors, and ToString / ToJSON
// formatting including the negative-month sign-fixup branch.
func TestIntervalValue(t *testing.T) {
	t.Parallel()

	iv := &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Months: 2, Days: 3}}

	t.Run("arithmetic and comparison unsupported", func(t *testing.T) {
		if _, err := iv.Add(iv); err == nil {
			t.Fatal("Add")
		}
		if _, err := iv.Sub(iv); err == nil {
			t.Fatal("Sub")
		}
		if _, err := iv.Mul(iv); err == nil {
			t.Fatal("Mul")
		}
		if _, err := iv.Div(iv); err == nil {
			t.Fatal("Div")
		}
		if _, err := iv.EQ(iv); err == nil {
			t.Fatal("EQ")
		}
		if _, err := iv.GT(iv); err == nil {
			t.Fatal("GT")
		}
		if _, err := iv.GTE(iv); err == nil {
			t.Fatal("GTE")
		}
		if _, err := iv.LT(iv); err == nil {
			t.Fatal("LT")
		}
		if _, err := iv.LTE(iv); err == nil {
			t.Fatal("LTE")
		}
	})

	t.Run("ToInt64/Float64/Bool/Array/Struct/Time/Rat error", func(t *testing.T) {
		if _, err := iv.ToInt64(); err == nil {
			t.Fatal("ToInt64")
		}
		if _, err := iv.ToFloat64(); err == nil {
			t.Fatal("ToFloat64")
		}
		if _, err := iv.ToBool(); err == nil {
			t.Fatal("ToBool")
		}
		if _, err := iv.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := iv.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
		if _, err := iv.ToTime(); err == nil {
			t.Fatal("ToTime")
		}
		if _, err := iv.ToRat(); err == nil {
			t.Fatal("ToRat")
		}
	})

	t.Run("ToString/ToBytes/ToJSON", func(t *testing.T) {
		s, _ := iv.ToString()
		if s == "" {
			t.Fatal("ToString empty")
		}
		b, _ := iv.ToBytes()
		if string(b) != s {
			t.Fatalf("ToBytes mismatch: %s vs %s", b, s)
		}
		j, _ := iv.ToJSON()
		// JSON is the string in quotes
		if len(j) < 2 || j[0] != '"' || j[len(j)-1] != '"' {
			t.Fatalf("ToJSON not quoted: %s", j)
		}
	})

	t.Run("ToString negative months fix-up", func(t *testing.T) {
		// Years==0 && Months<0 triggers the leading-minus branch.
		neg := &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Months: -1}}
		s, _ := neg.ToString()
		if s == "" || s[0] != '-' {
			t.Fatalf("ToString negative: %q", s)
		}
	})

	t.Run("Format/Interface", func(t *testing.T) {
		if iv.Format('t') == "" {
			t.Fatal("Format empty")
		}
		if iv.Interface() == nil {
			t.Fatal("Interface nil")
		}
	})
}
