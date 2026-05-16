package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestBoolValue covers BoolValue arithmetic-unsupported paths and the
// successful conversions to int64/float/string/bytes/JSON/rat.
func TestBoolValue(t *testing.T) {
	t.Parallel()

	t.Run("Add/Sub/Mul/Div all unsupported", func(t *testing.T) {
		bv := value.BoolValue(true)
		if _, err := bv.Add(bv); err == nil {
			t.Fatal("Add should be unsupported")
		}
		if _, err := bv.Sub(bv); err == nil {
			t.Fatal("Sub should be unsupported")
		}
		if _, err := bv.Mul(bv); err == nil {
			t.Fatal("Mul should be unsupported")
		}
		if _, err := bv.Div(bv); err == nil {
			t.Fatal("Div should be unsupported")
		}
	})

	t.Run("EQ matches", func(t *testing.T) {
		eq, err := value.BoolValue(true).EQ(value.BoolValue(true))
		if err != nil || !eq {
			t.Fatalf("true EQ true: %v / err=%v", eq, err)
		}
		eq, err = value.BoolValue(true).EQ(value.BoolValue(false))
		if err != nil || eq {
			t.Fatalf("true EQ false: %v / err=%v", eq, err)
		}
	})

	t.Run("EQ error when rhs not bool-convertible", func(t *testing.T) {
		// ArrayValue.ToBool returns an error.
		if _, err := value.BoolValue(true).EQ(&value.ArrayValue{}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("GT/GTE/LT/LTE unsupported", func(t *testing.T) {
		bv := value.BoolValue(true)
		if _, err := bv.GT(bv); err == nil {
			t.Fatal("GT should be unsupported")
		}
		if _, err := bv.GTE(bv); err == nil {
			t.Fatal("GTE should be unsupported")
		}
		if _, err := bv.LT(bv); err == nil {
			t.Fatal("LT should be unsupported")
		}
		if _, err := bv.LTE(bv); err == nil {
			t.Fatal("LTE should be unsupported")
		}
	})

	t.Run("conversions", func(t *testing.T) {
		bt := value.BoolValue(true)
		bf := value.BoolValue(false)
		if v, _ := bt.ToInt64(); v != 1 {
			t.Fatalf("true.ToInt64: %d", v)
		}
		if v, _ := bf.ToInt64(); v != 0 {
			t.Fatalf("false.ToInt64: %d", v)
		}
		if v, _ := bt.ToFloat64(); v != 1.0 {
			t.Fatalf("true.ToFloat64: %f", v)
		}
		if v, _ := bf.ToFloat64(); v != 0.0 {
			t.Fatalf("false.ToFloat64: %f", v)
		}
		if v, _ := bt.ToString(); v != "true" {
			t.Fatalf("true.ToString: %s", v)
		}
		if v, _ := bt.ToBytes(); string(v) != "true" {
			t.Fatalf("true.ToBytes: %s", v)
		}
		if v, _ := bt.ToBool(); v != true {
			t.Fatalf("true.ToBool: %v", v)
		}
		if v, _ := bt.ToJSON(); v != "true" {
			t.Fatalf("true.ToJSON: %s", v)
		}
	})

	t.Run("ToArray/ToStruct/ToTime return errors", func(t *testing.T) {
		bv := value.BoolValue(true)
		if _, err := bv.ToArray(); err == nil {
			t.Fatal("ToArray should err")
		}
		if _, err := bv.ToStruct(); err == nil {
			t.Fatal("ToStruct should err")
		}
		if _, err := bv.ToTime(); err == nil {
			t.Fatal("ToTime should err")
		}
	})

	t.Run("ToRat", func(t *testing.T) {
		r, err := value.BoolValue(true).ToRat()
		if err != nil || r.Num().Int64() != 1 {
			t.Fatalf("true.ToRat: %v / err=%v", r, err)
		}
		r, err = value.BoolValue(false).ToRat()
		if err != nil || r.Num().Int64() != 0 {
			t.Fatalf("false.ToRat: %v / err=%v", r, err)
		}
	})

	t.Run("Format and Interface", func(t *testing.T) {
		bv := value.BoolValue(true)
		if bv.Format('t') != "true" {
			t.Fatalf("Format t: %s", bv.Format('t'))
		}
		got, ok := bv.Interface().(bool)
		if !ok || !got {
			t.Fatalf("Interface: %v (%T)", bv.Interface(), bv.Interface())
		}
	})
}
