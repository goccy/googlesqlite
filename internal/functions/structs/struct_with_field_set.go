package structs

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// STRUCT_WITH_FIELD_SET returns a copy of the input struct with the
// field at the given (0-based) index replaced by newValue. Used by
// UPDATE rewrites of `col.field = value` — SQL/SQLite cannot assign
// to a function call, so the formatter rewrites those updates into
// a whole-column replacement that calls this function.
//
// The struct's Keys are preserved; only the Values entry at idx
// changes. Out-of-range idx returns the input unchanged.
func STRUCT_WITH_FIELD_SET(v value.Value, idx int, newValue value.Value) (value.Value, error) {
	sv, err := v.ToStruct()
	if err != nil {
		return nil, err
	}
	if sv == nil {
		return nil, nil
	}
	if idx < 0 || idx >= len(sv.Values) {
		return sv, nil
	}
	keys := append([]string(nil), sv.Keys...)
	values := append([]value.Value(nil), sv.Values...)
	values[idx] = newValue
	out := &value.StructValue{
		Keys:   keys,
		Values: values,
		M:      make(map[string]value.Value, len(keys)),
	}
	for i, k := range keys {
		out.M[k] = values[i]
	}
	return out, nil
}

func BindStructWithFieldSet(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("googlesqlite_struct_with_field_set: expected 3 args, got %d", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	idx, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	return STRUCT_WITH_FIELD_SET(args[0], int(idx), args[2])
}
