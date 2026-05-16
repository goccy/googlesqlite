package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestSafeValue confirms that SafeValue swallows errors from the
// wrapped Value (returning the zero result) while otherwise delegating
// to the inner value's behaviour.
func TestSafeValue(t *testing.T) {
	t.Parallel()

	t.Run("propagates success", func(t *testing.T) {
		sv := &value.SafeValue{Value: value.IntValue(5)}
		got, err := sv.Add(value.IntValue(3))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.IntValue(8) {
			t.Fatalf("Add: %v", got)
		}
	})

	t.Run("Add swallows error", func(t *testing.T) {
		sv := &value.SafeValue{Value: value.IntValue(1)}
		got, err := sv.Add(&value.ArrayValue{})
		if err != nil || got != nil {
			t.Fatalf("Add err=%v got=%v", err, got)
		}
	})

	t.Run("Sub/Mul/Div swallow errors", func(t *testing.T) {
		sv := &value.SafeValue{Value: value.IntValue(1)}
		if got, err := sv.Sub(&value.ArrayValue{}); err != nil || got != nil {
			t.Fatalf("Sub: %v %v", got, err)
		}
		if got, err := sv.Mul(&value.ArrayValue{}); err != nil || got != nil {
			t.Fatalf("Mul: %v %v", got, err)
		}
		if got, err := sv.Div(&value.ArrayValue{}); err != nil || got != nil {
			t.Fatalf("Div: %v %v", got, err)
		}
	})

	t.Run("comparisons swallow errors -> false,nil", func(t *testing.T) {
		sv := &value.SafeValue{Value: value.IntValue(1)}
		if ok, err := sv.EQ(&value.ArrayValue{}); err != nil || ok {
			t.Fatalf("EQ: %v %v", ok, err)
		}
		if ok, err := sv.GT(&value.ArrayValue{}); err != nil || ok {
			t.Fatalf("GT: %v %v", ok, err)
		}
		if ok, err := sv.GTE(&value.ArrayValue{}); err != nil || ok {
			t.Fatalf("GTE: %v %v", ok, err)
		}
		if ok, err := sv.LT(&value.ArrayValue{}); err != nil || ok {
			t.Fatalf("LT: %v %v", ok, err)
		}
		if ok, err := sv.LTE(&value.ArrayValue{}); err != nil || ok {
			t.Fatalf("LTE: %v %v", ok, err)
		}
	})

	t.Run("conversions swallow errors", func(t *testing.T) {
		// Wrap an IntValue(2) so ToBool() errors. SafeValue should
		// swallow the error and return zero value.
		sv := &value.SafeValue{Value: value.IntValue(2)}
		if got, err := sv.ToBool(); err != nil || got != false {
			t.Fatalf("ToBool: %v %v", got, err)
		}
		// ToInt64 succeeds (returns 2).
		if got, err := sv.ToInt64(); err != nil || got != 2 {
			t.Fatalf("ToInt64: %d %v", got, err)
		}
		// ToArray fails -> SafeValue substitutes an empty struct.
		if got, err := sv.ToArray(); err != nil || got == nil {
			t.Fatalf("ToArray: %v %v", got, err)
		}
		if got, err := sv.ToStruct(); err != nil || got == nil {
			t.Fatalf("ToStruct: %v %v", got, err)
		}
		// ToTime fails for IntValue(2) — actually it does not, the
		// IntValue path returns a valid time. Use BoolValue to force
		// failure.
		bv := &value.SafeValue{Value: value.BoolValue(true)}
		if got, err := bv.ToTime(); err != nil || !got.IsZero() {
			t.Fatalf("ToTime: %v %v", got, err)
		}
		// ToRat fails for BoolValue? Actually BoolValue.ToRat succeeds.
		// Use ArrayValue.
		av := &value.SafeValue{Value: &value.ArrayValue{}}
		if got, err := av.ToRat(); err != nil || got != nil {
			t.Fatalf("ToRat: %v %v", got, err)
		}
	})

	t.Run("ToBytes/ToString/ToFloat64/ToJSON success", func(t *testing.T) {
		sv := &value.SafeValue{Value: value.IntValue(7)}
		if got, err := sv.ToBytes(); err != nil || string(got) != "7" {
			t.Fatalf("ToBytes: %s %v", got, err)
		}
		if got, err := sv.ToString(); err != nil || got != "7" {
			t.Fatalf("ToString: %s %v", got, err)
		}
		if got, err := sv.ToFloat64(); err != nil || got != 7 {
			t.Fatalf("ToFloat64: %f %v", got, err)
		}
		if got, err := sv.ToJSON(); err != nil || got != "7" {
			t.Fatalf("ToJSON: %s %v", got, err)
		}
	})

	t.Run("Format and Interface delegate", func(t *testing.T) {
		sv := &value.SafeValue{Value: value.IntValue(7)}
		if sv.Format('t') != "7" {
			t.Fatalf("Format: %s", sv.Format('t'))
		}
		if v, ok := sv.Interface().(int64); !ok || v != 7 {
			t.Fatalf("Interface: %v (%T)", sv.Interface(), sv.Interface())
		}
	})
}
