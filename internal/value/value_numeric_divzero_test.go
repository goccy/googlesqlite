package value_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestNumericValueDivByZero covers the panic-recovery branch of
// NumericValue.Div. big.Rat.Inv(zero) panics with a plain string;
// the defer/recover in Div was repaired to handle non-error panics
// through fmt.Errorf("%v", r). The contract: Div by zero must return
// a non-nil error rather than crashing the goroutine.
//
// Authoritative behaviour: GoogleSQL specifies "division by zero" as
// a runtime error for SAFE.DIVIDE / DIV / `/` on NUMERIC and
// BIGNUMERIC types (docs/third_party/googlesql-docs/mathematical_functions.md
// "DIV" section). Returning an error from NumericValue.Div is how the
// runtime surfaces that condition.
func TestNumericValueDivByZero(t *testing.T) {
	t.Parallel()

	mk := func(n int64) *value.NumericValue {
		r := new(big.Rat)
		r.SetInt64(n)
		return &value.NumericValue{Rat: r}
	}

	_, err := mk(1).Div(mk(0))
	if err == nil {
		t.Fatalf("Div(0) returned nil err; want non-nil")
	}
	// The error string must mention the recovered panic message so
	// callers can diagnose the underlying cause.
	if !strings.Contains(err.Error(), "NumericValue.Div") {
		t.Fatalf("Div(0) err = %q; want substring NumericValue.Div", err.Error())
	}
}
