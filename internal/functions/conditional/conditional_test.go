package conditional

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestBindCoalesceFirstNonNull asserts COALESCE returns the first
// non-NULL argument, matching the GoogleSQL conditional-expressions
// spec.
func TestBindCoalesceFirstNonNull(t *testing.T) {
	got, err := BindCoalesce(nil, value.IntValue(7), value.IntValue(8))
	if err != nil {
		t.Fatalf("BindCoalesce: %v", err)
	}
	v, _ := got.ToInt64()
	if v != 7 {
		t.Fatalf("want 7, got %d", v)
	}
}

func TestBindCoalesceAllNull(t *testing.T) {
	got, err := BindCoalesce(nil, nil, nil)
	if err != nil {
		t.Fatalf("BindCoalesce: %v", err)
	}
	if got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

func TestBindCoalesceSingleArg(t *testing.T) {
	got, err := BindCoalesce(value.StringValue("only"))
	if err != nil {
		t.Fatalf("BindCoalesce: %v", err)
	}
	s, _ := got.ToString()
	if s != "only" {
		t.Fatalf("want 'only', got %q", s)
	}
}

func TestBindCoalesceZeroArgs(t *testing.T) {
	if _, err := BindCoalesce(); err == nil {
		t.Fatalf("expected error for zero arguments")
	}
}

func TestBindIfTrue(t *testing.T) {
	got, err := BindIf(value.BoolValue(true), value.IntValue(1), value.IntValue(2))
	if err != nil {
		t.Fatalf("BindIf: %v", err)
	}
	v, _ := got.ToInt64()
	if v != 1 {
		t.Fatalf("want 1, got %d", v)
	}
}

func TestBindIfFalse(t *testing.T) {
	got, err := BindIf(value.BoolValue(false), value.IntValue(1), value.IntValue(2))
	if err != nil {
		t.Fatalf("BindIf: %v", err)
	}
	v, _ := got.ToInt64()
	if v != 2 {
		t.Fatalf("want 2, got %d", v)
	}
}

// TestBindIfNullCondition treats a NULL condition as false per
// GoogleSQL semantics — the false_result branch is returned.
func TestBindIfNullCondition(t *testing.T) {
	got, err := BindIf(nil, value.IntValue(1), value.IntValue(2))
	if err != nil {
		t.Fatalf("BindIf: %v", err)
	}
	v, _ := got.ToInt64()
	if v != 2 {
		t.Fatalf("want 2 for NULL cond, got %d", v)
	}
}

func TestBindIfWrongArgCount(t *testing.T) {
	if _, err := BindIf(value.BoolValue(true), value.IntValue(1)); err == nil {
		t.Fatalf("expected error for 2 args")
	}
	if _, err := BindIf(); err == nil {
		t.Fatalf("expected error for 0 args")
	}
}

func TestBindIfNullExprNonNull(t *testing.T) {
	got, err := BindIfNull(value.StringValue("a"), value.StringValue("b"))
	if err != nil {
		t.Fatalf("BindIfNull: %v", err)
	}
	s, _ := got.ToString()
	if s != "a" {
		t.Fatalf("want 'a', got %q", s)
	}
}

func TestBindIfNullExprNull(t *testing.T) {
	got, err := BindIfNull(nil, value.StringValue("fallback"))
	if err != nil {
		t.Fatalf("BindIfNull: %v", err)
	}
	s, _ := got.ToString()
	if s != "fallback" {
		t.Fatalf("want 'fallback', got %q", s)
	}
}

func TestBindIfNullWrongArgCount(t *testing.T) {
	if _, err := BindIfNull(value.IntValue(1)); err == nil {
		t.Fatalf("expected error for 1 arg")
	}
}

// TestBindNullIfEqual: when expr == expr_to_match the result is NULL,
// per GoogleSQL spec.
func TestBindNullIfEqual(t *testing.T) {
	got, err := BindNullIf(value.IntValue(5), value.IntValue(5))
	if err != nil {
		t.Fatalf("BindNullIf: %v", err)
	}
	if got != nil {
		t.Fatalf("want nil for equal args, got %v", got)
	}
}

func TestBindNullIfDifferent(t *testing.T) {
	got, err := BindNullIf(value.IntValue(5), value.IntValue(6))
	if err != nil {
		t.Fatalf("BindNullIf: %v", err)
	}
	v, _ := got.ToInt64()
	if v != 5 {
		t.Fatalf("want 5, got %d", v)
	}
}

// TestBindNullIfExprNull: NULLIF(NULL, x) returns NULL.
func TestBindNullIfExprNull(t *testing.T) {
	got, err := BindNullIf(nil, value.IntValue(5))
	if err != nil {
		t.Fatalf("BindNullIf: %v", err)
	}
	if got != nil {
		t.Fatalf("want nil, got %v", got)
	}
}

func TestBindNullIfStrings(t *testing.T) {
	got, err := BindNullIf(value.StringValue("hi"), value.StringValue("bye"))
	if err != nil {
		t.Fatalf("BindNullIf: %v", err)
	}
	s, _ := got.ToString()
	if s != "hi" {
		t.Fatalf("want 'hi', got %q", s)
	}
}

func TestBindNullIfWrongArgCount(t *testing.T) {
	if _, err := BindNullIf(value.IntValue(1)); err == nil {
		t.Fatalf("expected error for 1 arg")
	}
}
