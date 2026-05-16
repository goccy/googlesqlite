package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestArrayValue covers Has, Add/Sub/Mul/Div (unsupported), positional
// element compare paths, conversions, and Format / Interface.
func TestArrayValue(t *testing.T) {
	t.Parallel()

	mk := func(vs ...int64) *value.ArrayValue {
		out := &value.ArrayValue{}
		for _, v := range vs {
			out.Values = append(out.Values, value.IntValue(v))
		}
		return out
	}

	t.Run("Has", func(t *testing.T) {
		a := mk(1, 2, 3)
		if ok, err := a.Has(value.IntValue(2)); err != nil || !ok {
			t.Fatalf("Has 2: %v / err=%v", ok, err)
		}
		if ok, err := a.Has(value.IntValue(99)); err != nil || ok {
			t.Fatalf("Has 99: %v / err=%v", ok, err)
		}
	})

	t.Run("Has error path", func(t *testing.T) {
		// element EQ failing because of an incompatible rhs.
		a := mk(1)
		if _, err := a.Has(&value.ArrayValue{}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Add/Sub/Mul/Div unsupported", func(t *testing.T) {
		a := mk(1)
		if _, err := a.Add(a); err == nil {
			t.Fatal("Add")
		}
		if _, err := a.Sub(a); err == nil {
			t.Fatal("Sub")
		}
		if _, err := a.Mul(a); err == nil {
			t.Fatal("Mul")
		}
		if _, err := a.Div(a); err == nil {
			t.Fatal("Div")
		}
	})

	t.Run("EQ same shape", func(t *testing.T) {
		a := mk(1, 2)
		b := mk(1, 2)
		c := mk(1, 3)
		if ok, _ := a.EQ(b); !ok {
			t.Fatal("equal")
		}
		if ok, _ := a.EQ(c); ok {
			t.Fatal("not equal expected")
		}
		// length mismatch
		if ok, _ := a.EQ(mk(1, 2, 3)); ok {
			t.Fatal("length mismatch")
		}
	})

	t.Run("GT/GTE/LT/LTE", func(t *testing.T) {
		// The current ArrayValue comparisons compare element-wise; the
		// internal call is rhs.GT(lhs), so a.GT(b) returns true only
		// when every element of b is greater than the corresponding
		// element of a.
		small := mk(1, 1)
		big := mk(2, 2)
		if ok, _ := small.GT(big); !ok {
			t.Fatal("GT element-wise")
		}
		if ok, _ := small.GTE(big); !ok {
			t.Fatal("GTE element-wise")
		}
		if ok, _ := big.LT(small); !ok {
			t.Fatal("LT element-wise")
		}
		if ok, _ := big.LTE(small); !ok {
			t.Fatal("LTE element-wise")
		}
		// length mismatch short-circuits to false
		if ok, _ := small.GT(mk(1)); ok {
			t.Fatal("GT length mismatch")
		}
		if ok, _ := small.GTE(mk(1)); ok {
			t.Fatal("GTE length mismatch")
		}
		if ok, _ := small.LT(mk(1)); ok {
			t.Fatal("LT length mismatch")
		}
		if ok, _ := small.LTE(mk(1)); ok {
			t.Fatal("LTE length mismatch")
		}
	})

	t.Run("ToInt64/Float64/Bool error", func(t *testing.T) {
		a := mk(1)
		if _, err := a.ToInt64(); err == nil {
			t.Fatal("ToInt64")
		}
		if _, err := a.ToFloat64(); err == nil {
			t.Fatal("ToFloat64")
		}
		if _, err := a.ToBool(); err == nil {
			t.Fatal("ToBool")
		}
		if _, err := a.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
		if _, err := a.ToTime(); err == nil {
			t.Fatal("ToTime")
		}
		if _, err := a.ToRat(); err == nil {
			t.Fatal("ToRat")
		}
	})

	t.Run("ToString/ToBytes/ToJSON", func(t *testing.T) {
		a := mk(1, 2, 3)
		s, _ := a.ToString()
		if s != "[1,2,3]" {
			t.Fatalf("ToString: %s", s)
		}
		b, _ := a.ToBytes()
		if string(b) != "[1,2,3]" {
			t.Fatalf("ToBytes: %s", b)
		}
		j, _ := a.ToJSON()
		if j != "[1,2,3]" {
			t.Fatalf("ToJSON: %s", j)
		}
	})

	t.Run("ToArray identity", func(t *testing.T) {
		a := mk(1)
		got, err := a.ToArray()
		if err != nil || got != a {
			t.Fatalf("ToArray identity: got %v / err=%v", got, err)
		}
	})

	t.Run("ToString with nil element", func(t *testing.T) {
		a := &value.ArrayValue{Values: []value.Value{nil, value.IntValue(7)}}
		s, _ := a.ToString()
		if s != "[null,7]" {
			t.Fatalf("ToString w/ nil: %s", s)
		}
	})

	t.Run("Format with nil element renders NULL", func(t *testing.T) {
		a := &value.ArrayValue{Values: []value.Value{nil, value.IntValue(7)}}
		got := a.Format('t')
		if got != "[NULL, 7]" {
			t.Fatalf("Format: %s", got)
		}
	})

	t.Run("Interface", func(t *testing.T) {
		a := &value.ArrayValue{Values: []value.Value{nil, value.IntValue(7)}}
		iv, ok := a.Interface().([]any)
		if !ok {
			t.Fatalf("Interface: %T", a.Interface())
		}
		if len(iv) != 2 || iv[0] != nil {
			t.Fatalf("Interface unexpected: %v", iv)
		}
		if iv[1] != int64(7) {
			t.Fatalf("Interface[1]: %v (%T)", iv[1], iv[1])
		}
	})
}
