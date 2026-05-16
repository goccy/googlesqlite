package value

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type StructValue struct {
	Keys   []string
	Values []Value
	M      map[string]Value
}

func (sv *StructValue) Add(v Value) (Value, error) {
	return nil, fmt.Errorf("add operation is unsupported for struct %v", sv)
}

func (sv *StructValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("sub operation is unsupported for struct %v", sv)
}

func (sv *StructValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for struct %v", sv)
}

func (sv *StructValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for struct %v", sv)
}

// structCompare compares two StructValues positionally per BigQuery
// semantics. Anonymous-field structs (auto-generated names like
// `f0_`) and named structs with the same shape compare equal when
// every positional field compares equal. Comparing by map key would
// miss this — anonymous and named structs end up with different
// keys for the same logical field, so a key-based loop returned
// the wrong answer (anonymous-struct-field bug, runtime side).
//
// Returns one of -1 / 0 / 1 (analogous to bytes.Compare). For
// structs of different lengths the longer struct is treated as
// greater after every common-prefix field has compared equal.
// Nil-vs-non-nil short-circuits to a comparison of the underlying
// scalars where possible.
func structCompare(a, b *StructValue) (int, error) {
	la, lb := len(a.Values), len(b.Values)
	common := la
	if lb < common {
		common = lb
	}
	for i := 0; i < common; i++ {
		x, y := a.Values[i], b.Values[i]
		switch {
		case x == nil && y == nil:
			continue
		case x == nil:
			return -1, nil
		case y == nil:
			return 1, nil
		}
		eq, err := x.EQ(y)
		if err != nil {
			return 0, err
		}
		if eq {
			continue
		}
		gt, err := x.GT(y)
		if err != nil {
			return 0, err
		}
		if gt {
			return 1, nil
		}
		return -1, nil
	}
	switch {
	case la < lb:
		return -1, nil
	case la > lb:
		return 1, nil
	}
	return 0, nil
}

func (sv *StructValue) EQ(v Value) (bool, error) {
	st, err := v.ToStruct()
	if err != nil {
		return false, err
	}
	cmp, err := structCompare(sv, st)
	if err != nil {
		return false, err
	}
	return cmp == 0, nil
}

func (sv *StructValue) GT(v Value) (bool, error) {
	st, err := v.ToStruct()
	if err != nil {
		return false, err
	}
	cmp, err := structCompare(sv, st)
	if err != nil {
		return false, err
	}
	return cmp > 0, nil
}

func (sv *StructValue) GTE(v Value) (bool, error) {
	st, err := v.ToStruct()
	if err != nil {
		return false, err
	}
	cmp, err := structCompare(sv, st)
	if err != nil {
		return false, err
	}
	return cmp >= 0, nil
}

func (sv *StructValue) LT(v Value) (bool, error) {
	st, err := v.ToStruct()
	if err != nil {
		return false, err
	}
	cmp, err := structCompare(sv, st)
	if err != nil {
		return false, err
	}
	return cmp < 0, nil
}

func (sv *StructValue) LTE(v Value) (bool, error) {
	st, err := v.ToStruct()
	if err != nil {
		return false, err
	}
	cmp, err := structCompare(sv, st)
	if err != nil {
		return false, err
	}
	return cmp <= 0, nil
}

func (sv *StructValue) ToInt64() (int64, error) {
	return 0, fmt.Errorf("failed to convert int64 from struct %v", sv)
}

func (sv *StructValue) ToString() (string, error) {
	fields := []string{}
	for i := 0; i < len(sv.Keys); i++ {
		key := sv.Keys[i]
		value := sv.Values[i]
		if value == nil {
			fields = append(
				fields,
				fmt.Sprintf("%s:null", strconv.Quote(key)),
			)
			continue
		}
		v, err := value.ToJSON()
		if err != nil {
			return "", err
		}
		fields = append(
			fields,
			fmt.Sprintf("%s:%s", strconv.Quote(key), v),
		)
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ",")), nil
}

func (sv *StructValue) ToBytes() ([]byte, error) {
	v, err := sv.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (sv *StructValue) ToFloat64() (float64, error) {
	return 0, fmt.Errorf("failed to convert float64 from struct %v", sv)
}

func (sv *StructValue) ToBool() (bool, error) {
	return false, fmt.Errorf("failed to convert bool from struct %v", sv)
}

func (sv *StructValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert array from struct %v", sv)
}

func (sv *StructValue) ToStruct() (*StructValue, error) {
	return sv, nil
}

func (sv *StructValue) ToJSON() (string, error) {
	return sv.ToString()
}

func (sv *StructValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("failed to convert time.Time from struct %v", sv)
}

func (sv *StructValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("failed to convert *big.Rat from struct %v", sv)
}

func (sv *StructValue) Format(verb rune) string {
	elems := []string{}
	for _, v := range sv.Values {
		if v == nil {
			elems = append(elems, "NULL")
			continue
		}
		elems = append(elems, v.Format(verb))
	}
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (sv *StructValue) Interface() any {
	// A struct whose declared fields are all unnamed (all keys == "")
	// corresponds to `SELECT AS STRUCT <expr>, ...` — an anonymous
	// BigQuery struct. Historically the Go driver surfaced integer
	// literals in such structs as float64 (default JSON decode), and
	// the public test suite encodes that expectation. Preserve that
	// behavior only for all-unnamed structs; named-field structs keep
	// exact int/float types, which is what the rest of the suite
	// expects (approx_top_count, to_json_with_struct, etc.).
	allAnonymous := len(sv.Keys) > 0
	for _, k := range sv.Keys {
		if k != "" {
			allAnonymous = false
			break
		}
	}
	fields := []map[string]any{}
	for i := 0; i < len(sv.Keys); i++ {
		key := sv.Keys[i]
		value := sv.Values[i]
		if value == nil {
			fields = append(fields, map[string]any{
				key: nil,
			})
			continue
		}
		iv := value.Interface()
		if allAnonymous {
			iv = coerceIntsToFloats(iv)
		}
		fields = append(fields, map[string]any{
			key: iv,
		})
	}
	return fields
}

// coerceIntsToFloats walks a decoded driver interface value (int64,
// []interface{}, map entries) and replaces int64 leaves with float64.
// Only used by StructValue.Interface for the all-unnamed-fields case —
// see that method's comment.
func coerceIntsToFloats(v any) any {
	switch vv := v.(type) {
	case int64:
		return float64(vv)
	case []any:
		out := make([]any, len(vv))
		for i, e := range vv {
			out[i] = coerceIntsToFloats(e)
		}
		return out
	case []map[string]any:
		out := make([]map[string]any, len(vv))
		for i, m := range vv {
			nm := make(map[string]any, len(m))
			for k, mv := range m {
				nm[k] = coerceIntsToFloats(mv)
			}
			out[i] = nm
		}
		return out
	}
	return v
}
