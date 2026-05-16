package value_test

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestDecodeValue + TestConvertArgs round-trips every supported Value
// shape through EncodeValue -> DecodeValue and exercises the fall-back
// branches that hand back a plain StringValue.
func TestDecodeValue(t *testing.T) {
	t.Parallel()

	roundtrip := func(t *testing.T, v value.Value) value.Value {
		t.Helper()
		enc, err := value.EncodeValue(v)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		dec, err := value.DecodeValue(enc)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		return dec
	}

	t.Run("nil", func(t *testing.T) {
		got, err := value.DecodeValue(nil)
		if err != nil || got != nil {
			t.Fatalf("nil: %v / err=%v", got, err)
		}
	})

	t.Run("int64 passthrough", func(t *testing.T) {
		got, err := value.DecodeValue(int64(42))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.IntValue(42) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("float64 passthrough", func(t *testing.T) {
		got, err := value.DecodeValue(float64(3.5))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.FloatValue(3.5) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("bool passthrough", func(t *testing.T) {
		got, err := value.DecodeValue(true)
		if err != nil {
			t.Fatal(err)
		}
		if got != value.BoolValue(true) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("unknown raw type returns error", func(t *testing.T) {
		if _, err := value.DecodeValue(complex64(0)); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-base64 string becomes StringValue", func(t *testing.T) {
		got, err := value.DecodeValue("not-base64!!!")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.StringValue); !ok {
			t.Fatalf("expected StringValue, got %T", got)
		}
	})

	t.Run("base64 of non-JSON becomes StringValue", func(t *testing.T) {
		// b64("hi") decodes successfully but isn't a JSON layout.
		got, err := value.DecodeValue("aGk=")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got.(value.StringValue); !ok {
			t.Fatalf("expected StringValue, got %T", got)
		}
	})

	t.Run("StringValue roundtrip", func(t *testing.T) {
		got := roundtrip(t, value.StringValue("hello"))
		s, _ := got.ToString()
		if s != "hello" {
			t.Fatalf("got %s", s)
		}
	})

	t.Run("BytesValue roundtrip", func(t *testing.T) {
		got := roundtrip(t, value.BytesValue([]byte{1, 2, 3}))
		b, _ := got.ToBytes()
		if len(b) != 3 || b[0] != 1 {
			t.Fatalf("got %v", b)
		}
	})

	t.Run("DateValue roundtrip", func(t *testing.T) {
		got := roundtrip(t, value.DateValue(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)))
		s, _ := got.ToString()
		if s != "2020-01-02" {
			t.Fatalf("got %s", s)
		}
	})

	t.Run("DatetimeValue roundtrip", func(t *testing.T) {
		got := roundtrip(t, value.DatetimeValue(time.Date(2020, 1, 2, 12, 30, 45, 0, time.UTC)))
		if _, ok := got.(value.DatetimeValue); !ok {
			t.Fatalf("type: %T", got)
		}
	})

	t.Run("TimeValue roundtrip", func(t *testing.T) {
		got := roundtrip(t, value.TimeValue(time.Date(0, 1, 1, 12, 30, 45, 0, time.UTC)))
		if _, ok := got.(value.TimeValue); !ok {
			t.Fatalf("type: %T", got)
		}
	})

	t.Run("TimestampValue roundtrip", func(t *testing.T) {
		got := roundtrip(t, value.TimestampValue(time.Date(2020, 1, 2, 12, 30, 45, 0, time.UTC)))
		if _, ok := got.(value.TimestampValue); !ok {
			t.Fatalf("type: %T", got)
		}
	})

	t.Run("JsonValue roundtrip", func(t *testing.T) {
		got := roundtrip(t, value.JsonValue(`{"x":1}`))
		s, _ := got.ToString()
		if s != `{"x":1}` {
			t.Fatalf("got %s", s)
		}
	})

	t.Run("ArrayValue roundtrip", func(t *testing.T) {
		av := &value.ArrayValue{Values: []value.Value{value.IntValue(1), value.StringValue("two"), nil}}
		got := roundtrip(t, av)
		arr, ok := got.(*value.ArrayValue)
		if !ok {
			t.Fatalf("type: %T", got)
		}
		if len(arr.Values) != 3 {
			t.Fatalf("len: %d", len(arr.Values))
		}
	})

	t.Run("StructValue roundtrip", func(t *testing.T) {
		sv := &value.StructValue{
			Keys:   []string{"a", "b"},
			Values: []value.Value{value.IntValue(1), value.StringValue("hi")},
			M: map[string]value.Value{
				"a": value.IntValue(1), "b": value.StringValue("hi"),
			},
		}
		got := roundtrip(t, sv)
		st, ok := got.(*value.StructValue)
		if !ok {
			t.Fatalf("type: %T", got)
		}
		if len(st.Keys) != 2 {
			t.Fatalf("keys: %v", st.Keys)
		}
	})

	t.Run("RangeValue roundtrip with both bounds", func(t *testing.T) {
		r := &value.RangeValue{
			Start:      value.DateValue(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:        value.DateValue(time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC)),
			ElemHeader: value.DateValueType,
		}
		got := roundtrip(t, r)
		rr, ok := got.(*value.RangeValue)
		if !ok {
			t.Fatalf("type: %T", got)
		}
		if rr.Start == nil || rr.End == nil {
			t.Fatalf("missing bounds: %+v", rr)
		}
	})

	t.Run("Geography roundtrip", func(t *testing.T) {
		g, _ := value.GeographyFromWKT("POINT (1 2)")
		got := roundtrip(t, g)
		gv, ok := got.(*value.GeographyValue)
		if !ok {
			t.Fatalf("type: %T", got)
		}
		s, _ := gv.ToWKT()
		if s != "POINT (1 2)" {
			t.Fatalf("WKT: %s", s)
		}
	})
}

// TestConvertArgs exercises the variadic decoder used by the FFI
// callback path.
func TestConvertArgs(t *testing.T) {
	t.Parallel()

	got, err := value.ConvertArgs(int64(1), float64(2.5), nil, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("len: %d", len(got))
	}
	if got[0] != value.IntValue(1) {
		t.Fatalf("got[0]: %v", got[0])
	}
	if got[1] != value.FloatValue(2.5) {
		t.Fatalf("got[1]: %v", got[1])
	}
	if got[2] != nil {
		t.Fatalf("got[2]: %v", got[2])
	}
	if _, ok := got[3].(value.StringValue); !ok {
		t.Fatalf("got[3]: %T", got[3])
	}
}

// TestConvertArgsPropagatesError makes sure an unsupported arg type
// surfaces an error rather than panicking.
func TestConvertArgsPropagatesError(t *testing.T) {
	t.Parallel()
	if _, err := value.ConvertArgs(complex64(0)); err == nil {
		t.Fatal("expected error")
	}
}

// TestDecodeValueArrayCoercesFloatLiterals decodes an array whose body
// contains a JSON literal with a decimal point, exercising the
// float-branch of coerceJSONNumber.
func TestDecodeValueArrayCoercesFloatLiterals(t *testing.T) {
	t.Parallel()

	av := &value.ArrayValue{Values: []value.Value{value.FloatValue(1.5)}}
	enc, err := value.EncodeValue(av)
	if err != nil {
		t.Fatal(err)
	}
	got, err := value.DecodeValue(enc)
	if err != nil {
		t.Fatal(err)
	}
	arr, ok := got.(*value.ArrayValue)
	if !ok {
		t.Fatalf("type: %T", got)
	}
	if len(arr.Values) != 1 {
		t.Fatalf("len: %d", len(arr.Values))
	}
	fv, ok := arr.Values[0].(value.FloatValue)
	if !ok || float64(fv) != 1.5 {
		t.Fatalf("element: %v (%T)", arr.Values[0], arr.Values[0])
	}
}

// TestDecodeValueRangeWithUnboundedStart drives the range-decoder
// branch where the start bound is JSON-null.
func TestDecodeValueRangeWithUnboundedStart(t *testing.T) {
	t.Parallel()

	r := &value.RangeValue{
		Start:      nil,
		End:        value.DateValue(time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC)),
		ElemHeader: value.DateValueType,
	}
	enc, err := value.EncodeValue(r)
	if err != nil {
		t.Fatal(err)
	}
	got, err := value.DecodeValue(enc)
	if err != nil {
		t.Fatal(err)
	}
	rr, ok := got.(*value.RangeValue)
	if !ok {
		t.Fatalf("type: %T", got)
	}
	if rr.Start != nil {
		t.Fatalf("start: %v", rr.Start)
	}
	if rr.End == nil {
		t.Fatal("end nil")
	}
}
