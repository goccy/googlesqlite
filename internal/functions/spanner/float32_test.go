package spanner

import (
	"math"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestBindFloat32FromJson(t *testing.T) {
	t.Parallel()

	// Plain numeric JSON value.
	got, err := BindFloat32FromJson(value.StringValue("3.5"))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 3.5 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// `null` JSON value -> NULL Value.
	v, err := BindFloat32FromJson(value.StringValue("null"))
	if err != nil || v != nil {
		t.Fatalf("expected null, got %v %v", v, err)
	}

	// exact mode: rejects out-of-range numbers.
	if _, err := BindFloat32FromJson(value.StringValue("1e40"), value.StringValue("exact")); err == nil {
		t.Fatal("expected out-of-range error")
	}

	// round mode: rounds to +Inf for out-of-range values (float32 rounds 1e40 to +Inf).
	got, err = BindFloat32FromJson(value.StringValue("1e40"), value.StringValue("round"))
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsInf(mustFloat64(t, got), 1) {
		t.Fatalf("expected +Inf, got %f", mustFloat64(t, got))
	}

	// Non-numeric JSON -> error.
	if _, err := BindFloat32FromJson(value.StringValue(`"abc"`)); err == nil {
		t.Fatal("expected non-numeric error")
	}

	// Invalid JSON -> error.
	if _, err := BindFloat32FromJson(value.StringValue("not-json")); err == nil {
		t.Fatal("expected json parse error")
	}

	if v, _ := BindFloat32FromJson(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindFloat32FromJson(); err == nil {
		t.Fatal("expected arg count error")
	}
	if _, err := BindFloat32FromJson(value.StringValue("1"), value.StringValue("x"), value.StringValue("y")); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindFloat32ArrayFromJson(t *testing.T) {
	t.Parallel()

	got, err := BindFloat32ArrayFromJson(value.StringValue("[1, 2.5, null]"))
	if err != nil {
		t.Fatal(err)
	}
	arrV, ok := got.(*value.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", got)
	}
	if len(arrV.Values) != 3 {
		t.Fatalf("expected 3, got %d", len(arrV.Values))
	}
	if arrV.Values[2] != nil {
		t.Fatal("expected null element")
	}

	// Empty array -> empty ArrayValue.
	got, err = BindFloat32ArrayFromJson(value.StringValue("[]"))
	if err != nil {
		t.Fatal(err)
	}
	arrV, _ = got.(*value.ArrayValue)
	if len(arrV.Values) != 0 {
		t.Fatal("expected empty array")
	}

	// Non-numeric element -> error.
	if _, err := BindFloat32ArrayFromJson(value.StringValue(`["abc"]`)); err == nil {
		t.Fatal("expected non-numeric error")
	}
	// Invalid JSON -> error.
	if _, err := BindFloat32ArrayFromJson(value.StringValue("not-json")); err == nil {
		t.Fatal("expected parse error")
	}
	if v, _ := BindFloat32ArrayFromJson(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindFloat32ArrayFromJson(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindFloat64ArrayFromJson(t *testing.T) {
	t.Parallel()

	got, err := BindFloat64ArrayFromJson(value.StringValue("[1, 2.5]"))
	if err != nil {
		t.Fatal(err)
	}
	arrV, ok := got.(*value.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", got)
	}
	if len(arrV.Values) != 2 {
		t.Fatalf("expected 2, got %d", len(arrV.Values))
	}

	if _, err := BindFloat64ArrayFromJson(value.StringValue(`["abc"]`)); err == nil {
		t.Fatal("expected non-numeric error")
	}
	if v, _ := BindFloat64ArrayFromJson(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindFloat64ArrayFromJson(); err == nil {
		t.Fatal("expected arg count error")
	}
}
