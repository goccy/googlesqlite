package spanner

import (
	"fmt"
	"math"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"
)

// BindFloat32FromJson coerces a JSON value to FLOAT32. The optional
// `wide_number_mode` STRING gates whether out-of-range numbers raise
// (`exact`) or round to the nearest representable FLOAT32 (`round`).
func BindFloat32FromJson(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("FLOAT32: invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	mode := "exact"
	if len(args) == 2 && args[1] != nil {
		s, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		mode = s
	}
	raw, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	num, ok := v.(float64)
	if !ok {
		return nil, fmt.Errorf("FLOAT32: JSON value is not numeric")
	}
	if mode == "exact" && (num > math.MaxFloat32 || num < -math.MaxFloat32) {
		return nil, fmt.Errorf("FLOAT32: value %g out of FLOAT32 range", num)
	}
	return value.FloatValue(float64(float32(num))), nil
}

// BindFloat32ArrayFromJson coerces a JSON array to ARRAY<FLOAT32>.
func BindFloat32ArrayFromJson(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("FLOAT32_ARRAY: invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	var arr []any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil, err
	}
	out := make([]value.Value, 0, len(arr))
	for _, x := range arr {
		if x == nil {
			out = append(out, nil)
			continue
		}
		f, ok := x.(float64)
		if !ok {
			return nil, fmt.Errorf("FLOAT32_ARRAY: array element is not numeric")
		}
		out = append(out, value.FloatValue(float64(float32(f))))
	}
	return &value.ArrayValue{Values: out}, nil
}

// BindFloat64ArrayFromJson coerces a JSON array to ARRAY<FLOAT64>.
func BindFloat64ArrayFromJson(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("FLOAT64_ARRAY: invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	var arr []any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil, err
	}
	out := make([]value.Value, 0, len(arr))
	for _, x := range arr {
		if x == nil {
			out = append(out, nil)
			continue
		}
		f, ok := x.(float64)
		if !ok {
			return nil, fmt.Errorf("FLOAT64_ARRAY: array element is not numeric")
		}
		out = append(out, value.FloatValue(f))
	}
	return &value.ArrayValue{Values: out}, nil
}
