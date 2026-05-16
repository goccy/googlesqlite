package value_test

import (
	"math"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestFloatValue covers FloatValue arithmetic / comparisons and the
// conversion paths including NaN, Inf, and zero-division.
func TestFloatValue(t *testing.T) {
	t.Parallel()

	t.Run("Add/Sub/Mul", func(t *testing.T) {
		if got, _ := value.FloatValue(1.5).Add(value.FloatValue(2.5)); got != value.FloatValue(4.0) {
			t.Fatalf("Add: %v", got)
		}
		if got, _ := value.FloatValue(5).Sub(value.FloatValue(2)); got != value.FloatValue(3) {
			t.Fatalf("Sub: %v", got)
		}
		if got, _ := value.FloatValue(2).Mul(value.FloatValue(3)); got != value.FloatValue(6) {
			t.Fatalf("Mul: %v", got)
		}
	})

	t.Run("Add/Sub/Mul/Div errors on non-numeric rhs", func(t *testing.T) {
		fv := value.FloatValue(1)
		if _, err := fv.Add(&value.ArrayValue{}); err == nil {
			t.Fatal("Add should err")
		}
		if _, err := fv.Sub(&value.ArrayValue{}); err == nil {
			t.Fatal("Sub should err")
		}
		if _, err := fv.Mul(&value.ArrayValue{}); err == nil {
			t.Fatal("Mul should err")
		}
		if _, err := fv.Div(&value.ArrayValue{}); err == nil {
			t.Fatal("Div should err")
		}
	})

	t.Run("Div", func(t *testing.T) {
		got, err := value.FloatValue(10).Div(value.FloatValue(4))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.FloatValue(2.5) {
			t.Fatalf("Div: %v", got)
		}
	})

	t.Run("Div by zero", func(t *testing.T) {
		if _, err := value.FloatValue(1).Div(value.FloatValue(0)); err == nil {
			t.Fatal("expected zero-div error")
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		a := value.FloatValue(1.0)
		b := value.FloatValue(2.0)
		if eq, _ := a.EQ(a); !eq {
			t.Fatal("EQ self")
		}
		if eq, _ := a.EQ(b); eq {
			t.Fatal("EQ diff")
		}
		if gt, _ := b.GT(a); !gt {
			t.Fatal("GT")
		}
		if gte, _ := a.GTE(a); !gte {
			t.Fatal("GTE")
		}
		if lt, _ := a.LT(b); !lt {
			t.Fatal("LT")
		}
		if lte, _ := a.LTE(a); !lte {
			t.Fatal("LTE")
		}
		if _, err := a.EQ(&value.ArrayValue{}); err == nil {
			t.Fatal("EQ err expected")
		}
		if _, err := a.GT(&value.ArrayValue{}); err == nil {
			t.Fatal("GT err expected")
		}
		if _, err := a.GTE(&value.ArrayValue{}); err == nil {
			t.Fatal("GTE err expected")
		}
		if _, err := a.LT(&value.ArrayValue{}); err == nil {
			t.Fatal("LT err expected")
		}
		if _, err := a.LTE(&value.ArrayValue{}); err == nil {
			t.Fatal("LTE err expected")
		}
	})

	t.Run("ToInt64/Float64/String/Bytes/JSON", func(t *testing.T) {
		fv := value.FloatValue(3.5)
		if got, _ := fv.ToInt64(); got != 3 {
			t.Fatalf("ToInt64: %d", got)
		}
		if got, _ := fv.ToFloat64(); got != 3.5 {
			t.Fatalf("ToFloat64: %f", got)
		}
		if got, _ := fv.ToString(); got != "3.5" {
			t.Fatalf("ToString: %s", got)
		}
		if got, _ := fv.ToBytes(); string(got) != "3.5" {
			t.Fatalf("ToBytes: %s", got)
		}
		if got, _ := fv.ToJSON(); got != "3.5" {
			t.Fatalf("ToJSON: %s", got)
		}
	})

	t.Run("ToBool", func(t *testing.T) {
		if v, _ := value.FloatValue(1).ToBool(); v != true {
			t.Fatalf("1.0 -> %v", v)
		}
		if v, _ := value.FloatValue(0).ToBool(); v != false {
			t.Fatalf("0.0 -> %v", v)
		}
		if _, err := value.FloatValue(2.5).ToBool(); err == nil {
			t.Fatal("2.5 should not convert")
		}
	})

	t.Run("ToArray/ToStruct error", func(t *testing.T) {
		fv := value.FloatValue(1)
		if _, err := fv.ToArray(); err == nil {
			t.Fatal("ToArray should err")
		}
		if _, err := fv.ToStruct(); err == nil {
			t.Fatal("ToStruct should err")
		}
	})

	t.Run("ToTime", func(t *testing.T) {
		got, err := value.FloatValue(123.456).ToTime()
		if err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Fatal("zero time")
		}
	})

	t.Run("ToRat", func(t *testing.T) {
		r, err := value.FloatValue(0.5).ToRat()
		if err != nil {
			t.Fatal(err)
		}
		if r.Cmp(r) != 0 {
			t.Fatal("comparison failed")
		}
	})

	t.Run("Format/Interface", func(t *testing.T) {
		fv := value.FloatValue(1.25)
		if fv.Format('t') != "1.25" {
			t.Fatalf("Format: %s", fv.Format('t'))
		}
		if got, ok := fv.Interface().(float64); !ok || got != 1.25 {
			t.Fatalf("Interface: %v (%T)", fv.Interface(), fv.Interface())
		}
	})

	t.Run("NaN/Inf", func(t *testing.T) {
		nan := value.FloatValue(math.NaN())
		if eq, _ := nan.EQ(nan); eq {
			t.Fatal("NaN should not equal itself")
		}
		inf := value.FloatValue(math.Inf(1))
		if got, _ := inf.ToFloat64(); !math.IsInf(got, 1) {
			t.Fatalf("Inf round-trip: %f", got)
		}
	})
}
