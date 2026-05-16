package json

import (
	"fmt"
	"math"

	gjson "github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Strict and array variants of the JSON typed accessors.
// Strict (e.g. INT32, UINT32, BOOL_ARRAY) raise on type mismatch;
// LAX_* counterparts (LAX_INT32, LAX_BOOL_ARRAY, ...) return NULL
// on mismatch, mirroring the behaviour of LAX_INT64 / LAX_FLOAT64
// already shipped.
//
// Naming convention: BindXxx for the SQLite-side adapter, the
// Go-level helper lives next to it.

// ---------------- strict scalar ----------------

// The strict JSON scalar accessors read only the first argument; the
// analyzer may append a trailing wide-number-mode option, so they
// short-circuit NULL but must not enforce arity.

func BindInt32(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	raw, ok := jsonScalarFromValue(args[0])
	if !ok {
		return nil, fmt.Errorf("INT32: JSON value is not numeric")
	}
	v, ok := laxInt64(raw)
	if !ok {
		return nil, fmt.Errorf("INT32: cannot convert JSON value to INT32")
	}
	iv, _ := v.ToInt64()
	if iv < math.MinInt32 || iv > math.MaxInt32 {
		return nil, fmt.Errorf("INT32: value %d out of range", iv)
	}
	return value.IntValue(iv), nil
}

func BindUint32(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	raw, ok := jsonScalarFromValue(args[0])
	if !ok {
		return nil, fmt.Errorf("UINT32: JSON value is not numeric")
	}
	v, ok := laxInt64(raw)
	if !ok {
		return nil, fmt.Errorf("UINT32: cannot convert JSON value to UINT32")
	}
	iv, _ := v.ToInt64()
	if iv < 0 || iv > math.MaxUint32 {
		return nil, fmt.Errorf("UINT32: value %d out of range", iv)
	}
	return value.IntValue(iv), nil
}

func BindUint64(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	raw, ok := jsonScalarFromValue(args[0])
	if !ok {
		return nil, fmt.Errorf("UINT64: JSON value is not numeric")
	}
	v, ok := laxInt64(raw)
	if !ok {
		return nil, fmt.Errorf("UINT64: cannot convert JSON value to UINT64")
	}
	iv, _ := v.ToInt64()
	if iv < 0 {
		return nil, fmt.Errorf("UINT64: value %d is negative", iv)
	}
	return value.IntValue(iv), nil
}

func BindFloat(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	raw, ok := jsonScalarFromValue(args[0])
	if !ok {
		return nil, fmt.Errorf("FLOAT: JSON value is not numeric")
	}
	v, ok := laxFloat64(raw)
	if !ok {
		return nil, fmt.Errorf("FLOAT: cannot convert JSON value to FLOAT")
	}
	return v, nil
}

// ---------------- lax scalar (extra) ----------------

func BindLaxInt32(args ...value.Value) (value.Value, error) {
	return bindLax(args, func(raw any) (value.Value, bool) {
		v, ok := laxInt64(raw)
		if !ok {
			return nil, false
		}
		iv, _ := v.ToInt64()
		if iv < math.MinInt32 || iv > math.MaxInt32 {
			return nil, false
		}
		return value.IntValue(iv), true
	})
}

func BindLaxUint32(args ...value.Value) (value.Value, error) {
	return bindLax(args, func(raw any) (value.Value, bool) {
		v, ok := laxInt64(raw)
		if !ok {
			return nil, false
		}
		iv, _ := v.ToInt64()
		if iv < 0 || iv > math.MaxUint32 {
			return nil, false
		}
		return value.IntValue(iv), true
	})
}

func BindLaxUint64(args ...value.Value) (value.Value, error) {
	return bindLax(args, func(raw any) (value.Value, bool) {
		v, ok := laxInt64(raw)
		if !ok {
			return nil, false
		}
		iv, _ := v.ToInt64()
		if iv < 0 {
			return nil, false
		}
		return value.IntValue(iv), true
	})
}

func BindLaxFloat(args ...value.Value) (value.Value, error) {
	return bindLax(args, laxFloat64)
}

// ---------------- arrays (strict) ----------------

// arrayFromJsonValue decodes a JSON array (top-level) into a Go
// slice. NULL JSON / scalar JSON return (nil, false).
func arrayFromJsonValue(v value.Value) ([]any, bool) {
	jv, ok := v.(value.JsonValue)
	if !ok {
		return nil, false
	}
	body := string(jv)
	if body == "" || body == "null" {
		return nil, false
	}
	var arr []any
	if err := gjson.Unmarshal([]byte(body), &arr); err != nil {
		return nil, false
	}
	return arr, true
}

// strictArray runs `conv` on each element of the JSON array. If
// any element cannot be converted, the function errors. NULL
// elements pass through as NULL.
func strictArray(args []value.Value, conv func(any) (value.Value, bool), name string) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	arr, ok := arrayFromJsonValue(args[0])
	if !ok {
		return nil, fmt.Errorf("%s: JSON value is not an array", name)
	}
	out := make([]value.Value, 0, len(arr))
	for i, e := range arr {
		if e == nil {
			out = append(out, nil)
			continue
		}
		v, ok := conv(e)
		if !ok {
			return nil, fmt.Errorf("%s: element %d cannot be converted", name, i)
		}
		out = append(out, v)
	}
	return &value.ArrayValue{Values: out}, nil
}

// laxArray runs `conv` on each element; on failure the element
// becomes NULL rather than aborting.
func laxArray(args []value.Value, conv func(any) (value.Value, bool)) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	arr, ok := arrayFromJsonValue(args[0])
	if !ok {
		return nil, nil
	}
	out := make([]value.Value, 0, len(arr))
	for _, e := range arr {
		if e == nil {
			out = append(out, nil)
			continue
		}
		v, ok := conv(e)
		if !ok {
			out = append(out, nil)
			continue
		}
		out = append(out, v)
	}
	return &value.ArrayValue{Values: out}, nil
}

// helpers reusing the lax* converters
func convInt32(raw any) (value.Value, bool) {
	v, ok := laxInt64(raw)
	if !ok {
		return nil, false
	}
	iv, _ := v.ToInt64()
	if iv < math.MinInt32 || iv > math.MaxInt32 {
		return nil, false
	}
	return value.IntValue(iv), true
}

func convUint32(raw any) (value.Value, bool) {
	v, ok := laxInt64(raw)
	if !ok {
		return nil, false
	}
	iv, _ := v.ToInt64()
	if iv < 0 || iv > math.MaxUint32 {
		return nil, false
	}
	return value.IntValue(iv), true
}

func convUint64(raw any) (value.Value, bool) {
	v, ok := laxInt64(raw)
	if !ok {
		return nil, false
	}
	iv, _ := v.ToInt64()
	if iv < 0 {
		return nil, false
	}
	return value.IntValue(iv), true
}

// ---------------- strict array Bindings ----------------

func BindBoolArray(args ...value.Value) (value.Value, error) {
	return strictArray(args, laxBool, "BOOL_ARRAY")
}
func BindInt32Array(args ...value.Value) (value.Value, error) {
	return strictArray(args, convInt32, "INT32_ARRAY")
}
func BindInt64Array(args ...value.Value) (value.Value, error) {
	return strictArray(args, laxInt64, "INT64_ARRAY")
}
func BindUint32Array(args ...value.Value) (value.Value, error) {
	return strictArray(args, convUint32, "UINT32_ARRAY")
}
func BindUint64Array(args ...value.Value) (value.Value, error) {
	return strictArray(args, convUint64, "UINT64_ARRAY")
}
func BindFloatArray(args ...value.Value) (value.Value, error) {
	return strictArray(args, laxFloat64, "FLOAT_ARRAY")
}
func BindDoubleArray(args ...value.Value) (value.Value, error) {
	return strictArray(args, laxFloat64, "DOUBLE_ARRAY")
}
func BindStringArray(args ...value.Value) (value.Value, error) {
	return strictArray(args, laxString, "STRING_ARRAY")
}

// ---------------- lax array Bindings ----------------

func BindLaxBoolArray(args ...value.Value) (value.Value, error) {
	return laxArray(args, laxBool)
}
func BindLaxInt32Array(args ...value.Value) (value.Value, error) {
	return laxArray(args, convInt32)
}
func BindLaxInt64Array(args ...value.Value) (value.Value, error) {
	return laxArray(args, laxInt64)
}
func BindLaxUint32Array(args ...value.Value) (value.Value, error) {
	return laxArray(args, convUint32)
}
func BindLaxUint64Array(args ...value.Value) (value.Value, error) {
	return laxArray(args, convUint64)
}
func BindLaxFloatArray(args ...value.Value) (value.Value, error) {
	return laxArray(args, laxFloat64)
}
func BindLaxDoubleArray(args ...value.Value) (value.Value, error) {
	return laxArray(args, laxFloat64)
}
func BindLaxStringArray(args ...value.Value) (value.Value, error) {
	return laxArray(args, laxString)
}
