package value_test

import (
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/value"
)

// TestEncodeValue covers the EncodeValue path for every supported
// Value type, including the primitive short-circuits, the SafeValue
// passthrough, and the various JSON envelopes.
func TestEncodeValue(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		got, err := value.EncodeValue(nil)
		if err != nil || got != nil {
			t.Fatalf("nil: %v / err=%v", got, err)
		}
	})

	t.Run("IntValue passthrough", func(t *testing.T) {
		got, err := value.EncodeValue(value.IntValue(42))
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := got.(int64); !ok || v != 42 {
			t.Fatalf("got %v (%T)", got, got)
		}
	})

	t.Run("FloatValue passthrough", func(t *testing.T) {
		got, err := value.EncodeValue(value.FloatValue(3.5))
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := got.(float64); !ok || v != 3.5 {
			t.Fatalf("got %v (%T)", got, got)
		}
	})

	t.Run("BoolValue passthrough", func(t *testing.T) {
		got, err := value.EncodeValue(value.BoolValue(true))
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := got.(bool); !ok || !v {
			t.Fatalf("got %v (%T)", got, got)
		}
	})

	t.Run("SafeValue passthrough", func(t *testing.T) {
		got, err := value.EncodeValue(&value.SafeValue{Value: value.IntValue(7)})
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := got.(int64); !ok || v != 7 {
			t.Fatalf("got %v (%T)", got, got)
		}
	})

	t.Run("StringValue envelope", func(t *testing.T) {
		got, err := value.EncodeValue(value.StringValue("hello"))
		if err != nil {
			t.Fatal(err)
		}
		s, ok := got.(string)
		if !ok {
			t.Fatalf("got %T", got)
		}
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		var layout value.ValueLayout
		if err := json.Unmarshal(b, &layout); err != nil {
			t.Fatal(err)
		}
		if layout.Header != value.StringValueType || layout.Body != "hello" {
			t.Fatalf("layout: %+v", layout)
		}
	})

	t.Run("ValueLayoutFromValue exhaustive", func(t *testing.T) {
		// Just confirm every supported Value type produces a layout.
		r := new(big.Rat)
		r.SetInt64(1)
		cases := map[string]value.Value{
			"string":   value.StringValue("x"),
			"bytes":    value.BytesValue([]byte("abc")),
			"numeric":  &value.NumericValue{Rat: r},
			"bignum":   &value.NumericValue{Rat: r, IsBigNumeric: true},
			"date":     value.DateValue(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
			"dt":       value.DatetimeValue(time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)),
			"time":     value.TimeValue(time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)),
			"ts":       value.TimestampValue(time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)),
			"interval": &value.IntervalValue{IntervalValue: &bigquery.IntervalValue{Days: 1}},
			"json":     value.JsonValue(`{"x":1}`),
		}
		for name, v := range cases {
			t.Run(name, func(t *testing.T) {
				layout, err := value.ValueLayoutFromValue(v)
				if err != nil {
					t.Fatal(err)
				}
				if layout.Header == "" {
					t.Fatalf("missing header for %s", name)
				}
			})
		}
	})

	t.Run("ArrayValue with NULL element", func(t *testing.T) {
		av := &value.ArrayValue{Values: []value.Value{value.IntValue(1), nil, value.IntValue(3)}}
		layout, err := value.ValueLayoutFromValue(av)
		if err != nil {
			t.Fatal(err)
		}
		if layout.Header != value.ArrayValueType {
			t.Fatalf("header: %s", layout.Header)
		}
	})

	t.Run("StructValue", func(t *testing.T) {
		sv := &value.StructValue{
			Keys:   []string{"a"},
			Values: []value.Value{value.IntValue(1)},
			M:      map[string]value.Value{"a": value.IntValue(1)},
		}
		layout, err := value.ValueLayoutFromValue(sv)
		if err != nil {
			t.Fatal(err)
		}
		if layout.Header != value.StructValueType {
			t.Fatalf("header: %s", layout.Header)
		}
	})

	t.Run("Geography", func(t *testing.T) {
		g, err := value.GeographyFromWKT("POINT (1 2)")
		if err != nil {
			t.Fatal(err)
		}
		layout, err := value.ValueLayoutFromValue(g)
		if err != nil {
			t.Fatal(err)
		}
		if layout.Header != value.GeographyValueType {
			t.Fatalf("header: %s", layout.Header)
		}
	})

	t.Run("Range bounded and unbounded", func(t *testing.T) {
		// fully bounded
		r := &value.RangeValue{
			Start:      value.DateValue(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:        value.DateValue(time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC)),
			ElemHeader: value.DateValueType,
		}
		layout, err := value.ValueLayoutFromValue(r)
		if err != nil {
			t.Fatal(err)
		}
		if layout.Header != value.RangeValueType {
			t.Fatalf("header: %s", layout.Header)
		}
		// unbounded both ends
		r2 := &value.RangeValue{ElemHeader: value.DateValueType}
		if _, err := value.ValueLayoutFromValue(r2); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unsupported type returns error", func(t *testing.T) {
		// SafeValue is handled in EncodeValue's primitive switch but
		// not in valueLayoutFromValue.
		if _, err := value.ValueLayoutFromValue(&value.SafeValue{Value: value.IntValue(1)}); err == nil {
			t.Fatal("expected error for SafeValue layout")
		}
	})
}

// TestValueFromGoValue covers the reflection-based Go -> Value
// converter for primitives, slices, maps, structs, time, pointers, and
// the interface-recursion path.
func TestValueFromGoValue(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		got, err := value.ValueFromGoValue(nil)
		if err != nil || got != nil {
			t.Fatalf("nil: %v / err=%v", got, err)
		}
	})

	t.Run("primitives", func(t *testing.T) {
		cases := map[string]any{
			"int":    int(7),
			"int8":   int8(7),
			"int16":  int16(7),
			"int32":  int32(7),
			"int64":  int64(7),
			"uint":   uint(7),
			"uint8":  uint8(7), // single byte (NOT a []byte)
			"uint16": uint16(7),
			"uint32": uint32(7),
			"uint64": uint64(7),
			"f32":    float32(1.5),
			"f64":    float64(1.5),
			"bool":   true,
			"string": "abc",
		}
		for name, v := range cases {
			t.Run(name, func(t *testing.T) {
				got, err := value.ValueFromGoValue(v)
				if err != nil {
					t.Fatal(err)
				}
				if got == nil {
					t.Fatal("got nil")
				}
			})
		}
	})

	t.Run("[]byte -> BytesValue", func(t *testing.T) {
		got, err := value.ValueFromGoValue([]byte("hi"))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.BytesValue); !ok {
			t.Fatalf("expected BytesValue, got %T", got)
		}
	})

	t.Run("[]int -> ArrayValue", func(t *testing.T) {
		got, err := value.ValueFromGoValue([]int{1, 2, 3})
		if err != nil {
			t.Fatal(err)
		}
		av, ok := got.(*value.ArrayValue)
		if !ok {
			t.Fatalf("expected ArrayValue, got %T", got)
		}
		if len(av.Values) != 3 {
			t.Fatalf("len: %d", len(av.Values))
		}
	})

	t.Run("map -> StructValue", func(t *testing.T) {
		got, err := value.ValueFromGoValue(map[string]int{"a": 1})
		if err != nil {
			t.Fatal(err)
		}
		sv, ok := got.(*value.StructValue)
		if !ok {
			t.Fatalf("expected StructValue, got %T", got)
		}
		if len(sv.Keys) != 1 || sv.Keys[0] != "a" {
			t.Fatalf("keys: %v", sv.Keys)
		}
	})

	t.Run("struct -> StructValue", func(t *testing.T) {
		type S struct {
			A int
			B string
		}
		got, err := value.ValueFromGoValue(S{A: 1, B: "x"})
		if err != nil {
			t.Fatal(err)
		}
		sv, ok := got.(*value.StructValue)
		if !ok {
			t.Fatalf("expected StructValue, got %T", got)
		}
		if len(sv.Keys) != 2 {
			t.Fatalf("keys: %v", sv.Keys)
		}
	})

	t.Run("time.Time -> TimestampValue", func(t *testing.T) {
		got, err := value.ValueFromGoValue(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.TimestampValue); !ok {
			t.Fatalf("expected TimestampValue, got %T", got)
		}
	})

	t.Run("pointer recurses", func(t *testing.T) {
		x := 7
		got, err := value.ValueFromGoValue(&x)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.IntValue); !ok {
			t.Fatalf("expected IntValue, got %T", got)
		}
	})

	t.Run("any-typed nil inside interface", func(t *testing.T) {
		var v any
		got, err := value.ValueFromGoValue(v)
		if err != nil || got != nil {
			t.Fatalf("nil any: %v / err=%v", got, err)
		}
	})

	t.Run("unsupported kind returns error", func(t *testing.T) {
		// A function value has Kind == Func, which isn't handled.
		if _, err := value.ValueFromGoValue(func() {}); err == nil {
			t.Fatal("expected error")
		}
	})
}
