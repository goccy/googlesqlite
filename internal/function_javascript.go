//go:build !js

package internal

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	"modernc.org/quickjs"

	"github.com/goccy/googlesqlite/internal/value"
)

// EVAL_JAVASCRIPT runs a JavaScript UDF body with the given arguments and
// converts the result to the declared return type.
//
// The JS engine is modernc.org/quickjs — a pure-Go ccgo transpilation of
// QuickJS. It is used here instead of a reflection-based engine (goja)
// on purpose: goja reaches reflect.Value.MethodByName, which flips the Go
// linker into conservative dead-code mode and retains every method of
// every interface-converted type (tens of MB of unused go-googlesql
// accessors). quickjs does not, so the linker can prune normally.
func EVAL_JAVASCRIPT(code string, retType *Type, argNames []string, args []value.Value) (value.Value, error) {
	vm, err := quickjs.NewVM()
	if err != nil {
		return nil, fmt.Errorf("failed to create javascript VM: %w", err)
	}
	defer vm.Close()

	// Bind each argument as a global. The values are injected as JSON
	// literals — JSON is a subset of JavaScript — which avoids exposing
	// Go objects to the engine via reflection.
	var prelude strings.Builder
	for i := range args {
		var v any
		if args[i] != nil {
			if structV, ok := args[i].(*value.StructValue); ok {
				v = structV.M
			} else {
				v = args[i].Interface()
			}
		}
		enc, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to encode argument %s as %v: %w", argNames[i], args[i], err)
		}
		fmt.Fprintf(&prelude, "var %s = %s;\n", argNames[i], enc)
	}
	evalCode := prelude.String() + fmt.Sprintf(
		"function googlesqlite_javascript_func() { %s }\ngooglesqlite_javascript_func();",
		code,
	)
	ret, err := vm.Eval(evalCode, quickjs.EvalGlobal)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate javascript code %s: %w", code, err)
	}
	typ, err := retType.ToGoogleSQLType()
	if err != nil {
		return nil, fmt.Errorf("failed to get return type: %w", err)
	}
	v, err := castJavaScriptValue(typ, ret)
	if err != nil {
		return nil, fmt.Errorf("failed to convert googlesqlite value from %v: %w", ret, err)
	}
	return v, nil
}

// castJavaScriptValue converts the result of VM.Eval — a native Go value
// (nil / string / int / bool / float64 / *big.Int / *quickjs.Object /
// quickjs.Undefined) — to a googlesqlite value of the declared type.
func castJavaScriptValue(t googlesql.Googlesql_TypeNode, v any) (value.Value, error) {
	if v == nil {
		return nil, nil
	}
	if _, ok := v.(quickjs.Undefined); ok {
		return nil, nil
	}
	switch m1(t.Kind()) {
	case googlesql.TypeKindTypeInt32, googlesql.TypeKindTypeInt64, googlesql.TypeKindTypeUint32, googlesql.TypeKindTypeUint64:
		return value.IntValue(jsToInt64(v)), nil
	case googlesql.TypeKindTypeBool:
		return value.BoolValue(jsToBool(v)), nil
	case googlesql.TypeKindTypeFloat, googlesql.TypeKindTypeDouble:
		return value.FloatValue(jsToFloat64(v)), nil
	case googlesql.TypeKindTypeString, googlesql.TypeKindTypeEnum:
		return value.StringValue(jsToString(v)), nil
	case googlesql.TypeKindTypeBytes:
		return value.BytesValue(jsToString(v)), nil
	case googlesql.TypeKindTypeDate:
		d, err := value.ParseDate(jsToString(v))
		if err != nil {
			return nil, err
		}
		return value.DateValue(d), nil
	case googlesql.TypeKindTypeDatetime:
		d, err := value.ParseDatetime(jsToString(v))
		if err != nil {
			return nil, err
		}
		return value.DatetimeValue(d), nil
	case googlesql.TypeKindTypeTime:
		d, err := value.ParseTime(jsToString(v))
		if err != nil {
			return nil, err
		}
		return value.TimeValue(d), nil
	case googlesql.TypeKindTypeTimestamp:
		d, err := value.ParseTimestamp(jsToString(v), time.UTC)
		if err != nil {
			return nil, err
		}
		return value.TimestampValue(d), nil
	case googlesql.TypeKindTypeInterval:
		return value.ParseInterval(jsToString(v))
	case googlesql.TypeKindTypeNumeric, googlesql.TypeKindTypeBignumeric:
		return &value.NumericValue{Rat: jsToRat(v)}, nil
	case googlesql.TypeKindTypeJson:
		return value.JsonValue(jsToString(v)), nil
	case googlesql.TypeKindTypeArray:
		exported, err := jsExport(v)
		if err != nil {
			return nil, err
		}
		arr, ok := exported.([]any)
		if !ok {
			return nil, fmt.Errorf("expected a JavaScript array, got %T", exported)
		}
		var ret value.ArrayValue
		for _, vv := range arr {
			base, err := ValueFromGoValue(vv)
			if err != nil {
				return nil, err
			}
			ret.Values = append(ret.Values, base)
		}
		return &ret, nil
	case googlesql.TypeKindTypeStruct, googlesql.TypeKindTypeGeography:
		exported, err := jsExport(v)
		if err != nil {
			return nil, err
		}
		base, err := ValueFromGoValue(exported)
		if err != nil {
			return nil, err
		}
		return CastValue(t, base)
	}
	return nil, fmt.Errorf("unsupported cast %v from JavaScript value", m1(t.Kind()))
}

// jsExport converts a quickjs object result to its natural Go shape
// ([]any for arrays, map[string]any for objects) via JSON. Scalars are
// returned unchanged.
func jsExport(v any) (any, error) {
	obj, ok := v.(*quickjs.Object)
	if !ok {
		return v, nil
	}
	raw, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func jsToInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	case bool:
		if x {
			return 1
		}
		return 0
	case *big.Int:
		return x.Int64()
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	}
	return 0
}

func jsToFloat64(v any) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case float64:
		return x
	case bool:
		if x {
			return 1
		}
		return 0
	case *big.Int:
		f := new(big.Float).SetInt(x)
		r, _ := f.Float64()
		return r
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}

func jsToBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case string:
		return x != ""
	}
	return v != nil
}

func jsToString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	case *big.Int:
		return x.String()
	}
	return fmt.Sprint(v)
}

func jsToRat(v any) *big.Rat {
	r := new(big.Rat)
	switch x := v.(type) {
	case int:
		r.SetInt64(int64(x))
	case int64:
		r.SetInt64(x)
	case float64:
		r.SetFloat64(x)
	case *big.Int:
		r.SetInt(x)
	case string:
		r.SetString(x)
	}
	return r
}
