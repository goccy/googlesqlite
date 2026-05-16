package spanner

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestBindAddDateSubDate(t *testing.T) {
	t.Parallel()

	d := value.DateValue(time.Date(2023, 6, 14, 0, 0, 0, 0, time.UTC))

	got, err := BindAddDate(d, value.IntValue(10))
	if err != nil {
		t.Fatal(err)
	}
	dv, ok := got.(value.DateValue)
	if !ok {
		t.Fatalf("expected DateValue, got %T", got)
	}
	want := time.Date(2023, 6, 24, 0, 0, 0, 0, time.UTC)
	if !time.Time(dv).Equal(want) {
		t.Fatalf("got %v want %v", time.Time(dv), want)
	}

	got, err = BindSubDate(d, value.IntValue(5))
	if err != nil {
		t.Fatal(err)
	}
	dv, ok = got.(value.DateValue)
	if !ok {
		t.Fatalf("expected DateValue, got %T", got)
	}
	want = time.Date(2023, 6, 9, 0, 0, 0, 0, time.UTC)
	if !time.Time(dv).Equal(want) {
		t.Fatalf("got %v want %v", time.Time(dv), want)
	}

	// First-argument type must be DATE.
	if _, err := BindAddDate(value.IntValue(1), value.IntValue(1)); err == nil {
		t.Fatal("expected type error")
	}
	if _, err := BindSubDate(value.IntValue(1), value.IntValue(1)); err == nil {
		t.Fatal("expected type error")
	}

	// Null propagation.
	if v, _ := BindAddDate(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindSubDate(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}

	// Arg count errors.
	if _, err := BindAddDate(d); err == nil {
		t.Fatal("expected arg count error")
	}
	if _, err := BindSubDate(d); err == nil {
		t.Fatal("expected arg count error")
	}
}
