package value_test

import (
	"math"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestIntValue exercises arithmetic, comparison, and conversion paths
// on value.IntValue including overflow, zero-division, and edge values.
func TestIntValue(t *testing.T) {
	t.Parallel()

	t.Run("Add", func(t *testing.T) {
		got, err := value.IntValue(2).Add(value.IntValue(3))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.IntValue(5) {
			t.Fatalf("Add: got %v, want 5", got)
		}
	})

	t.Run("Add error when rhs not int-convertible", func(t *testing.T) {
		// ArrayValue.ToInt64 returns an error.
		if _, err := value.IntValue(1).Add(&value.ArrayValue{}); err == nil {
			t.Fatal("expected error adding ArrayValue to IntValue")
		}
	})

	t.Run("Sub", func(t *testing.T) {
		got, err := value.IntValue(10).Sub(value.IntValue(4))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.IntValue(6) {
			t.Fatalf("Sub: got %v, want 6", got)
		}
	})

	t.Run("Sub error", func(t *testing.T) {
		if _, err := value.IntValue(1).Sub(&value.ArrayValue{}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Mul", func(t *testing.T) {
		got, err := value.IntValue(6).Mul(value.IntValue(7))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.IntValue(42) {
			t.Fatalf("Mul: got %v, want 42", got)
		}
	})

	t.Run("Mul error", func(t *testing.T) {
		if _, err := value.IntValue(1).Mul(&value.ArrayValue{}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Div", func(t *testing.T) {
		got, err := value.IntValue(20).Div(value.IntValue(4))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.IntValue(5) {
			t.Fatalf("Div: got %v, want 5", got)
		}
	})

	t.Run("Div by zero", func(t *testing.T) {
		if _, err := value.IntValue(1).Div(value.IntValue(0)); err == nil {
			t.Fatal("expected zero-division error")
		}
	})

	t.Run("Div error in conversion", func(t *testing.T) {
		if _, err := value.IntValue(1).Div(&value.ArrayValue{}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("EQ true", func(t *testing.T) {
		ok, err := value.IntValue(5).EQ(value.IntValue(5))
		if err != nil || !ok {
			t.Fatalf("EQ: %v, err=%v", ok, err)
		}
	})

	t.Run("EQ false", func(t *testing.T) {
		ok, err := value.IntValue(5).EQ(value.IntValue(6))
		if err != nil || ok {
			t.Fatalf("EQ: %v, err=%v", ok, err)
		}
	})

	t.Run("EQ error", func(t *testing.T) {
		if _, err := value.IntValue(1).EQ(&value.ArrayValue{}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GT GTE LT LTE", func(t *testing.T) {
		a := value.IntValue(3)
		b := value.IntValue(5)
		if ok, _ := a.GT(b); ok {
			t.Fatal("3 GT 5 should be false")
		}
		if ok, _ := b.GT(a); !ok {
			t.Fatal("5 GT 3 should be true")
		}
		if ok, _ := a.GTE(a); !ok {
			t.Fatal("3 GTE 3 should be true")
		}
		if ok, _ := a.LT(b); !ok {
			t.Fatal("3 LT 5 should be true")
		}
		if ok, _ := a.LTE(a); !ok {
			t.Fatal("3 LTE 3 should be true")
		}
		// error paths
		if _, err := a.GT(&value.ArrayValue{}); err == nil {
			t.Fatal("expected GT error")
		}
		if _, err := a.GTE(&value.ArrayValue{}); err == nil {
			t.Fatal("expected GTE error")
		}
		if _, err := a.LT(&value.ArrayValue{}); err == nil {
			t.Fatal("expected LT error")
		}
		if _, err := a.LTE(&value.ArrayValue{}); err == nil {
			t.Fatal("expected LTE error")
		}
	})

	t.Run("ToInt64/ToFloat64/ToString/ToBytes/ToJSON", func(t *testing.T) {
		iv := value.IntValue(42)
		if got, _ := iv.ToInt64(); got != 42 {
			t.Fatalf("ToInt64: %d", got)
		}
		if got, _ := iv.ToFloat64(); got != 42.0 {
			t.Fatalf("ToFloat64: %f", got)
		}
		if got, _ := iv.ToString(); got != "42" {
			t.Fatalf("ToString: %s", got)
		}
		if got, _ := iv.ToBytes(); string(got) != "42" {
			t.Fatalf("ToBytes: %s", got)
		}
		if got, _ := iv.ToJSON(); got != "42" {
			t.Fatalf("ToJSON: %s", got)
		}
	})

	t.Run("ToBool", func(t *testing.T) {
		if v, _ := value.IntValue(0).ToBool(); v != false {
			t.Fatalf("0 -> %v", v)
		}
		if v, _ := value.IntValue(1).ToBool(); v != true {
			t.Fatalf("1 -> %v", v)
		}
		if _, err := value.IntValue(2).ToBool(); err == nil {
			t.Fatal("2 should not convert to bool")
		}
	})

	t.Run("ToArray/ToStruct return errors", func(t *testing.T) {
		if _, err := value.IntValue(1).ToArray(); err == nil {
			t.Fatal("expected ToArray error")
		}
		if _, err := value.IntValue(1).ToStruct(); err == nil {
			t.Fatal("expected ToStruct error")
		}
	})

	t.Run("ToTime small int interprets as DATE", func(t *testing.T) {
		// Small ints fall through to DateFromInt64Value (days since epoch).
		got, err := value.IntValue(0).ToTime()
		if err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Fatalf("ToTime(0): got zero time")
		}
	})

	t.Run("ToTime large int interprets as TIMESTAMP", func(t *testing.T) {
		got, err := value.IntValue(2 * 1000 * 1000).ToTime()
		if err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Fatalf("ToTime(large): got zero time")
		}
	})

	t.Run("ToRat", func(t *testing.T) {
		r, err := value.IntValue(-7).ToRat()
		if err != nil {
			t.Fatal(err)
		}
		if r.Num().Int64() != -7 || r.Denom().Int64() != 1 {
			t.Fatalf("ToRat: %s", r.String())
		}
	})

	t.Run("Format and Interface", func(t *testing.T) {
		iv := value.IntValue(123)
		if iv.Format('t') != "123" {
			t.Fatalf("Format t: %s", iv.Format('t'))
		}
		if iv.Format('T') != "123" {
			t.Fatalf("Format T: %s", iv.Format('T'))
		}
		v, ok := iv.Interface().(int64)
		if !ok || v != 123 {
			t.Fatalf("Interface: %v (%T)", iv.Interface(), iv.Interface())
		}
	})

	t.Run("MaxInt64/MinInt64", func(t *testing.T) {
		if got, _ := value.IntValue(math.MaxInt64).ToInt64(); got != math.MaxInt64 {
			t.Fatalf("MaxInt64 round-trip: %d", got)
		}
		if got, _ := value.IntValue(math.MinInt64).ToInt64(); got != math.MinInt64 {
			t.Fatalf("MinInt64 round-trip: %d", got)
		}
	})
}
