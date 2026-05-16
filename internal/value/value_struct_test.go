package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestStructValue covers positional struct comparison, conversions,
// Format / Interface, and the anonymous-struct int->float coercion.
func TestStructValue(t *testing.T) {
	t.Parallel()

	mk := func(keys []string, vals []value.Value) *value.StructValue {
		m := map[string]value.Value{}
		for i, k := range keys {
			m[k] = vals[i]
		}
		return &value.StructValue{Keys: keys, Values: vals, M: m}
	}

	t.Run("Add/Sub/Mul/Div unsupported", func(t *testing.T) {
		s := mk([]string{"a"}, []value.Value{value.IntValue(1)})
		if _, err := s.Add(s); err == nil {
			t.Fatal("Add")
		}
		if _, err := s.Sub(s); err == nil {
			t.Fatal("Sub")
		}
		if _, err := s.Mul(s); err == nil {
			t.Fatal("Mul")
		}
		if _, err := s.Div(s); err == nil {
			t.Fatal("Div")
		}
	})

	t.Run("EQ positional", func(t *testing.T) {
		a := mk([]string{"x"}, []value.Value{value.IntValue(1)})
		b := mk([]string{"y"}, []value.Value{value.IntValue(1)})
		if ok, _ := a.EQ(b); !ok {
			t.Fatal("expected equal across renamed key")
		}
	})

	t.Run("comparisons longer = greater", func(t *testing.T) {
		a := mk([]string{"x"}, []value.Value{value.IntValue(1)})
		b := mk([]string{"x", "y"}, []value.Value{value.IntValue(1), value.IntValue(2)})
		if ok, _ := a.LT(b); !ok {
			t.Fatal("shorter LT longer")
		}
		if ok, _ := b.GT(a); !ok {
			t.Fatal("longer GT shorter")
		}
		if ok, _ := a.LTE(a); !ok {
			t.Fatal("LTE self")
		}
		if ok, _ := a.GTE(a); !ok {
			t.Fatal("GTE self")
		}
	})

	t.Run("nil-vs-non-nil field short-circuits", func(t *testing.T) {
		a := mk([]string{"x"}, []value.Value{nil})
		b := mk([]string{"x"}, []value.Value{value.IntValue(1)})
		if ok, _ := a.LT(b); !ok {
			t.Fatal("nil should compare less")
		}
		if ok, _ := b.GT(a); !ok {
			t.Fatal("non-nil should compare greater than nil")
		}
		// nil EQ nil case
		a2 := mk([]string{"x"}, []value.Value{nil})
		if ok, _ := a.EQ(a2); !ok {
			t.Fatal("nil EQ nil")
		}
	})

	t.Run("scalar comparison fall-through", func(t *testing.T) {
		// when EQ returns false but GT returns true / false
		a := mk([]string{"x"}, []value.Value{value.IntValue(2)})
		b := mk([]string{"x"}, []value.Value{value.IntValue(1)})
		if ok, _ := a.GT(b); !ok {
			t.Fatal("a>b")
		}
		if ok, _ := b.LT(a); !ok {
			t.Fatal("b<a")
		}
	})

	t.Run("ToInt64/Float64/Bool/Array/Time/Rat error", func(t *testing.T) {
		s := mk([]string{"x"}, []value.Value{value.IntValue(1)})
		if _, err := s.ToInt64(); err == nil {
			t.Fatal("ToInt64")
		}
		if _, err := s.ToFloat64(); err == nil {
			t.Fatal("ToFloat64")
		}
		if _, err := s.ToBool(); err == nil {
			t.Fatal("ToBool")
		}
		if _, err := s.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := s.ToTime(); err == nil {
			t.Fatal("ToTime")
		}
		if _, err := s.ToRat(); err == nil {
			t.Fatal("ToRat")
		}
	})

	t.Run("ToStruct identity", func(t *testing.T) {
		s := mk([]string{"x"}, []value.Value{value.IntValue(1)})
		got, err := s.ToStruct()
		if err != nil || got != s {
			t.Fatalf("ToStruct identity")
		}
	})

	t.Run("ToString/ToBytes/ToJSON", func(t *testing.T) {
		s := mk(
			[]string{"a", "b"},
			[]value.Value{value.IntValue(1), nil},
		)
		out, _ := s.ToString()
		if out != `{"a":1,"b":null}` {
			t.Fatalf("ToString: %s", out)
		}
		b, _ := s.ToBytes()
		if string(b) != out {
			t.Fatalf("ToBytes: %s", b)
		}
		j, _ := s.ToJSON()
		if j != out {
			t.Fatalf("ToJSON: %s", j)
		}
	})

	t.Run("Format with nil renders NULL", func(t *testing.T) {
		s := mk([]string{"a", "b"}, []value.Value{nil, value.IntValue(2)})
		got := s.Format('t')
		if got != "(NULL, 2)" {
			t.Fatalf("Format: %s", got)
		}
	})

	t.Run("Interface anonymous coerces ints to floats", func(t *testing.T) {
		// keys all empty -> anonymous struct, coerced.
		s := mk([]string{"", ""}, []value.Value{value.IntValue(1), value.IntValue(2)})
		fields, ok := s.Interface().([]map[string]any)
		if !ok {
			t.Fatalf("Interface type: %T", s.Interface())
		}
		if len(fields) != 2 {
			t.Fatalf("Interface len: %d", len(fields))
		}
		if v, ok := fields[0][""].(float64); !ok || v != 1.0 {
			t.Fatalf("Interface[0]: %v (%T)", fields[0][""], fields[0][""])
		}
	})

	t.Run("Interface named keeps int64", func(t *testing.T) {
		s := mk([]string{"a", "b"}, []value.Value{value.IntValue(1), nil})
		fields, ok := s.Interface().([]map[string]any)
		if !ok {
			t.Fatalf("Interface type: %T", s.Interface())
		}
		if v, ok := fields[0]["a"].(int64); !ok || v != 1 {
			t.Fatalf("Interface[a]: %v (%T)", fields[0]["a"], fields[0]["a"])
		}
		if fields[1]["b"] != nil {
			t.Fatalf("Interface[b]: %v", fields[1]["b"])
		}
	})

	t.Run("CoerceIntsToFloats helper", func(t *testing.T) {
		// int64 -> float64
		if got := value.CoerceIntsToFloats(int64(7)); got != float64(7) {
			t.Fatalf("int64: %v (%T)", got, got)
		}
		// []any recurse
		arr := []any{int64(1), int64(2)}
		coerced, ok := value.CoerceIntsToFloats(arr).([]any)
		if !ok || len(coerced) != 2 {
			t.Fatalf("[]any: %v (%T)", coerced, coerced)
		}
		if coerced[0].(float64) != 1.0 {
			t.Fatalf("[]any[0]: %v", coerced[0])
		}
		// []map[string]any recurse
		mp := []map[string]any{{"a": int64(3)}}
		got, ok := value.CoerceIntsToFloats(mp).([]map[string]any)
		if !ok || got[0]["a"].(float64) != 3.0 {
			t.Fatalf("[]map: %v", got)
		}
		// other unchanged
		if got := value.CoerceIntsToFloats("hi"); got != "hi" {
			t.Fatalf("string passthrough: %v", got)
		}
	})
}
