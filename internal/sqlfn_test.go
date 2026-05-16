package internal

import (
	"errors"
	"testing"

	"github.com/goccy/googlesqlite/internal/functions/helper"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestScalar1NullPropagation(t *testing.T) {
	bind := helper.Scalar1(func(v value.Value) (value.Value, error) {
		t.Fatalf("body should not run on NULL input")
		return nil, nil
	})
	got, err := bind(nil)
	if err != nil || got != nil {
		t.Fatalf("Scalar1 NULL: got=%v err=%v, want (nil, nil)", got, err)
	}
}

func TestScalar1Forwards(t *testing.T) {
	bind := helper.Scalar1(func(v value.Value) (value.Value, error) {
		return v, nil
	})
	got, err := bind(value.IntValue(7))
	if err != nil {
		t.Fatal(err)
	}
	iv, ok := got.(value.IntValue)
	if !ok || iv != 7 {
		t.Fatalf("Scalar1 forward: got=%v want IntValue(7)", got)
	}
}

func TestScalar1ArityCheck(t *testing.T) {
	bind := helper.Scalar1(func(value.Value) (value.Value, error) { return nil, nil })
	if _, err := bind(); err == nil {
		t.Fatal("Scalar1 with 0 args should error")
	}
	if _, err := bind(value.IntValue(1), value.IntValue(2)); err == nil {
		t.Fatal("Scalar1 with 2 args should error")
	}
}

func TestScalar2NullPropagation(t *testing.T) {
	called := false
	bind := helper.Scalar2(func(a, b value.Value) (value.Value, error) {
		called = true
		return a, nil
	})
	for _, args := range [][]value.Value{
		{nil, value.IntValue(1)},
		{value.IntValue(1), nil},
		{nil, nil},
	} {
		got, err := bind(args...)
		if got != nil || err != nil {
			t.Fatalf("Scalar2 NULL %v: got=%v err=%v", args, got, err)
		}
	}
	if called {
		t.Fatal("Scalar2 body fired on NULL")
	}
}

func TestScalar3Forwards(t *testing.T) {
	bind := helper.Scalar3(func(a, b, c value.Value) (value.Value, error) {
		ai, _ := a.ToInt64()
		bi, _ := b.ToInt64()
		ci, _ := c.ToInt64()
		return value.IntValue(ai + bi + ci), nil
	})
	got, err := bind(value.IntValue(1), value.IntValue(2), value.IntValue(3))
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := got.ToInt64(); v != 6 {
		t.Fatalf("Scalar3 sum: got %v want 6", v)
	}
}

func TestScalarNNullPropagation(t *testing.T) {
	called := false
	bind := helper.ScalarN(func(args ...value.Value) (value.Value, error) {
		called = true
		return args[0], nil
	})
	got, err := bind(value.IntValue(1), nil, value.IntValue(3))
	if got != nil || err != nil {
		t.Fatalf("ScalarN NULL: got=%v err=%v", got, err)
	}
	if called {
		t.Fatal("ScalarN body fired on NULL")
	}
}

func TestKeepNullVariantsObserveNull(t *testing.T) {
	bind := helper.Scalar1KeepNull(func(v value.Value) (value.Value, error) {
		if v == nil {
			return value.BoolValue(true), nil
		}
		return value.BoolValue(false), nil
	})
	got, err := bind(nil)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := got.ToBool(); !b {
		t.Fatalf("Scalar1KeepNull on NULL should observe nil; got %v", got)
	}
}

func TestErrorPropagation(t *testing.T) {
	want := errors.New("boom")
	bind := helper.Scalar1(func(value.Value) (value.Value, error) {
		return nil, want
	})
	if _, err := bind(value.IntValue(1)); !errors.Is(err, want) {
		t.Fatalf("error not propagated: %v", err)
	}
}
