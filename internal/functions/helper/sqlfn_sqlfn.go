package helper

// Package sqlfn collects the scalar-function binder helpers that let
// call sites declare arity-typed wrappers around their semantic
// functions instead of repeating the same `existsNull(args)` /
// argument-fanout boilerplate in every `bindXxx` helper.
//
// The helpers are intentionally type-light: googlesqlite's runtime
// value type is the (non-parameterized) `value.Value` interface, so
// the wrappers trade Go-side type parameters for a small set of
// arity-typed shapes. Each helper:
//
//   - asserts the expected arg count (so an analyzer-bridge mismatch
//     surfaces as a clear runtime error rather than a silent panic);
//   - propagates NULL by returning (nil, nil) the moment any arg is
//     nil (the GoogleSQL default for scalar functions);
//   - forwards the typed args to the caller's pure semantic function.
//
// Variants ending in `KeepNull` skip the NULL-propagation step for
// the handful of functions (COALESCE, IFNULL, IS_NULL, …) that need
// to observe NULLs themselves.

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// BindFunction is the signature ncruces' adapter expects from us:
// take an arg list of (possibly NULL) googlesqlite Values and return
// either a Value or an error.
type BindFunction = func(...value.Value) (value.Value, error)

// Scalar1 wraps a 1-arg semantic function. Returns NULL when the
// single argument is NULL.
func Scalar1(fn func(value.Value) (value.Value, error)) BindFunction {
	return func(args ...value.Value) (value.Value, error) {
		if err := scalarArity(args, 1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return nil, nil
		}
		return fn(args[0])
	}
}

// Scalar2 wraps a 2-arg semantic function. Returns NULL when either
// argument is NULL.
func Scalar2(fn func(a, b value.Value) (value.Value, error)) BindFunction {
	return func(args ...value.Value) (value.Value, error) {
		if err := scalarArity(args, 2); err != nil {
			return nil, err
		}
		if args[0] == nil || args[1] == nil {
			return nil, nil
		}
		return fn(args[0], args[1])
	}
}

// Scalar3 wraps a 3-arg semantic function. Returns NULL when any
// argument is NULL.
func Scalar3(fn func(a, b, c value.Value) (value.Value, error)) BindFunction {
	return func(args ...value.Value) (value.Value, error) {
		if err := scalarArity(args, 3); err != nil {
			return nil, err
		}
		if args[0] == nil || args[1] == nil || args[2] == nil {
			return nil, nil
		}
		return fn(args[0], args[1], args[2])
	}
}

// ScalarN wraps an arbitrary-arity semantic function that should
// short-circuit when any argument is NULL.
func ScalarN(fn func(...value.Value) (value.Value, error)) BindFunction {
	return func(args ...value.Value) (value.Value, error) {
		for _, a := range args {
			if a == nil {
				return nil, nil
			}
		}
		return fn(args...)
	}
}

// Scalar1KeepNull wraps a 1-arg semantic function that needs to
// observe NULL itself (e.g. IS_NULL).
func Scalar1KeepNull(fn func(value.Value) (value.Value, error)) BindFunction {
	return func(args ...value.Value) (value.Value, error) {
		if err := scalarArity(args, 1); err != nil {
			return nil, err
		}
		return fn(args[0])
	}
}

// Scalar2KeepNull wraps a 2-arg semantic function that needs to
// observe NULL itself (e.g. IS_DISTINCT_FROM).
func Scalar2KeepNull(fn func(a, b value.Value) (value.Value, error)) BindFunction {
	return func(args ...value.Value) (value.Value, error) {
		if err := scalarArity(args, 2); err != nil {
			return nil, err
		}
		return fn(args[0], args[1])
	}
}

// ScalarNKeepNull wraps a variadic semantic function that needs to
// observe NULLs itself (e.g. COALESCE).
func ScalarNKeepNull(fn func(...value.Value) (value.Value, error)) BindFunction {
	return func(args ...value.Value) (value.Value, error) {
		return fn(args...)
	}
}

// ExistsNull reports whether any of the args is the runtime NULL
// (a nil value.Value). Per-spec helpers that handle their own arity
// can use this to short-circuit NULL propagation.
func ExistsNull(args []value.Value) bool {
	for _, a := range args {
		if a == nil {
			return true
		}
	}
	return false
}

func scalarArity(args []value.Value, want int) error {
	if len(args) != want {
		return fmt.Errorf("invalid number of arguments: got %d, want %d", len(args), want)
	}
	return nil
}
