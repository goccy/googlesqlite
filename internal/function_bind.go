package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/functions/window"
	"github.com/goccy/googlesqlite/internal/value"
)

type SQLiteFunction func(...any) (any, error)
type BindFunction func(...value.Value) (value.Value, error)
type aggregateBindFunction func() func() *helper.Aggregator
type windowBindFunction func() func() *window.WindowAggregator

type funcInfo struct {
	Name     string
	BindFunc BindFunction
	// NonDeterministic, when true, registers the underlying SQLite
	// function without the SQLITE_DETERMINISTIC flag. SQLite is then
	// required to re-evaluate the call for every row, instead of
	// constant-folding it. Required for functions like GENERATE_UUID
	// or CURRENT_TIMESTAMP whose result must vary per call.
	NonDeterministic bool
}

type aggregateFuncInfo struct {
	Name     string
	BindFunc aggregateBindFunction
}

type windowFuncInfo struct {
	Name     string
	BindFunc windowBindFunction
}

// BIT_CAST_TO_* are wired through Scalar1 in normalFuncs; no
// per-function bindXxx helper is needed.

// bindBool / bindInt64 are identity passthroughs: they must observe a
// NULL argument and return it unchanged, so they use Scalar1KeepNull
// (arity check only, no NULL short-circuit).
var bindBool = helper.Scalar1KeepNull(func(v value.Value) (value.Value, error) {
	return v, nil
})

var bindInt64 = helper.Scalar1KeepNull(func(v value.Value) (value.Value, error) {
	return v, nil
})

// bindDouble implements `FLOAT64(json_expr[, wide_number_mode])`.
// Strict: errors on non-numeric JSON. wide_number_mode controls
// rounding behaviour for numbers that don't fit FLOAT64 exactly;
// 'exact' raises, 'round' tolerates the precision loss. Default is
// 'round' per the BigQuery spec.
func bindDouble(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("FLOAT64: invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	mode := "round"
	if len(args) == 2 && args[1] != nil {
		s, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		switch s {
		case "exact", "round":
			mode = s
		default:
			return nil, fmt.Errorf("FLOAT64: unexpected wide_number_mode: %s", s)
		}
	}
	jv, ok := args[0].(value.JsonValue)
	if !ok {
		// Already-numeric pass-through: legacy callers feed FLOAT64
		// to coerce non-JSON values; let the conversion happen.
		f, err := args[0].ToFloat64()
		if err != nil {
			return nil, err
		}
		return value.FloatValue(f), nil
	}
	body := strings.TrimSpace(string(jv))
	if body == "" || body == "null" {
		return nil, nil
	}
	if body[0] == '"' || body[0] == '{' || body[0] == '[' || body == "true" || body == "false" {
		return nil, fmt.Errorf("FLOAT64: JSON value is not a number")
	}
	f, err := strconv.ParseFloat(body, 64)
	if err != nil {
		return nil, fmt.Errorf("FLOAT64: failed to parse JSON number %q: %w", body, err)
	}
	if mode == "exact" {
		// Re-serialise and compare; round-trip mismatch means we
		// lost precision.
		if strconv.FormatFloat(f, 'g', -1, 64) != body {
			return nil, fmt.Errorf("FLOAT64: number %q cannot be represented as FLOAT64 without loss", body)
		}
	}
	return value.FloatValue(f), nil
}

func bindDistinct(args ...value.Value) (value.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("DISTINCT: invalid number of arguments: got %d, want 0", len(args))
	}
	return helper.DISTINCT()
}

func bindIgnoreNulls(args ...value.Value) (value.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("IGNORE_NULLS: invalid number of arguments: got %d, want 0", len(args))
	}
	return helper.IGNORE_NULLS()
}

var bindWindowRowID = helper.Scalar1(func(v value.Value) (value.Value, error) {
	a0, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	return window.WINDOW_ROWID(a0)
})

func bindEvalJavaScript(args ...value.Value) (value.Value, error) {
	code, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	encodedType, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	var typ Type
	if err := json.Unmarshal([]byte(encodedType), &typ); err != nil {
		return nil, fmt.Errorf("EVAL_JAVASCRIPT: failed to decode type information from %s: %w", encodedType, err)
	}
	if len(args) == 2 {
		return EVAL_JAVASCRIPT(code, &typ, nil, nil)
	}
	argNames, err := args[2].ToArray()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(argNames.Values))
	for _, val := range argNames.Values {
		name, err := val.ToString()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return EVAL_JAVASCRIPT(code, &typ, names, args[3:])
}
