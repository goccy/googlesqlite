package internal

import (
	"database/sql/driver"
	"fmt"
	"strings"

	googlesql "github.com/goccy/go-googlesql"

	"github.com/goccy/googlesqlite/internal/value"
)

// TypedValue is a query argument whose GoogleSQL type is declared by
// the caller instead of inferred from its Go value. Inference cannot
// recover struct field order from a Go map, nor the type of a NULL
// field or of an empty array's elements; a caller that knows the
// declared type (for example a BigQuery parameterType) passes it here.
//
// Value is cast to Type: struct fields are matched by name,
// case-insensitively, and a field missing from Value is NULL. For a
// named parameter Type is also the type the analyzer declares; a
// positional parameter stays undeclared, like every other positional
// parameter.
type TypedValue struct {
	Type  *Type
	Value any
}

// checkTypedNamedValue encodes tv.Value against tv.Type and stores the
// result in value. The TypedValue wrapper is kept, with Value replaced
// by the encoded layout, so Analyze can still read the declared type;
// Analyze unwraps it (unwrapTypedArgs) before anything else sees the
// argument.
func checkTypedNamedValue(value *driver.NamedValue, tv TypedValue) error {
	if tv.Type == nil {
		return fmt.Errorf("TypedValue: nil Type")
	}
	v, err := ValueFromGoValue(tv.Value)
	if err != nil {
		return err
	}
	cast, err := castToDeclaredType(tv.Type, v)
	if err != nil {
		return fmt.Errorf("TypedValue: %w", err)
	}
	var encoded any
	if cast != nil {
		if encoded, err = EncodeValue(cast); err != nil {
			return err
		}
	}
	value.Value = TypedValue{Type: tv.Type, Value: encoded}
	return nil
}

// checkPreparedNamedValue is CheckNamedValue for a prepared statement.
// Its parameter types were fixed when it was prepared, so a TypedValue
// contributes only its encoded value.
func checkPreparedNamedValue(value *driver.NamedValue) error {
	if err := CheckNamedValue(value); err != nil {
		return err
	}
	if tv, ok := value.Value.(TypedValue); ok {
		value.Value = tv.Value
	}
	return nil
}

// castToDeclaredType walks t alongside v: arrays cast element-wise,
// structs are rebuilt in t's field order (fields matched by name,
// case-insensitively; absent fields become NULL, undeclared fields are
// an error), scalars go through CastValue.
func castToDeclaredType(t *Type, v value.Value) (value.Value, error) {
	if v == nil {
		return nil, nil
	}
	switch {
	case t.IsArray():
		arr, err := v.ToArray()
		if err != nil {
			return nil, err
		}
		out := &value.ArrayValue{}
		for _, elem := range arr.Values {
			cast, err := castToDeclaredType(t.ElementType, elem)
			if err != nil {
				return nil, err
			}
			out.Values = append(out.Values, cast)
		}
		return out, nil
	case t.IsStruct():
		st, err := v.ToStruct()
		if err != nil {
			return nil, err
		}
		byName := make(map[string]value.Value, len(st.Keys))
		for i, k := range st.Keys {
			byName[strings.ToLower(k)] = st.Values[i]
		}
		out := &value.StructValue{M: map[string]value.Value{}}
		for _, f := range t.FieldTypes {
			name := strings.ToLower(f.Name)
			cast, err := castToDeclaredType(f.Type, byName[name])
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", f.Name, err)
			}
			delete(byName, name)
			out.Keys = append(out.Keys, f.Name)
			out.Values = append(out.Values, cast)
			out.M[f.Name] = cast
		}
		for name := range byName {
			return nil, fmt.Errorf("field %s is not in the declared STRUCT type", name)
		}
		return out, nil
	}
	gt, err := t.ToGoogleSQLType()
	if err != nil {
		return nil, err
	}
	return CastValue(gt, v)
}

// unwrapTypedArgs returns args with every TypedValue replaced by its
// encoded value, plus the declared types keyed by lower-cased name.
func unwrapTypedArgs(args []driver.NamedValue) ([]driver.NamedValue, map[string]*Type) {
	var declared map[string]*Type
	var out []driver.NamedValue
	for i, a := range args {
		tv, ok := a.Value.(TypedValue)
		if !ok {
			continue
		}
		if out == nil {
			out = append([]driver.NamedValue(nil), args...)
			declared = map[string]*Type{}
		}
		out[i].Value = tv.Value
		if a.Name != "" {
			declared[strings.ToLower(a.Name)] = tv.Type
		}
	}
	if out == nil {
		return args, nil
	}
	return out, declared
}

// googleSQLTypeForDecodedValue derives a full GoogleSQL type from a
// decoded value, recursing into arrays and structs. It fails when any
// struct field is NULL, since that carries no type.
func googleSQLTypeForDecodedValue(factory *googlesql.TypeFactory, v value.Value) (googlesql.Googlesql_TypeNode, bool) {
	switch dv := v.(type) {
	case nil:
		return nil, false
	case *value.ArrayValue:
		for _, elem := range dv.Values {
			if elem == nil {
				continue
			}
			et, ok := googleSQLTypeForDecodedValue(factory, elem)
			if !ok {
				return nil, false
			}
			at, err := factory.MakeArrayType2(et)
			if err != nil || at == nil {
				return nil, false
			}
			return at, true
		}
		// No element to inspect: keep the historical ARRAY<STRING>.
		return makeArrayType(factory, googlesql.TypeKindTypeString)
	case *value.StructValue:
		fields := make([]*googlesql.StructField, 0, len(dv.Keys))
		for i, k := range dv.Keys {
			ft, ok := googleSQLTypeForDecodedValue(factory, dv.Values[i])
			if !ok {
				return nil, false
			}
			fields = append(fields, &googlesql.StructField{Name: k, Type_: ft})
		}
		st, err := factory.MakeStructType2(fields)
		if err != nil || st == nil {
			return nil, false
		}
		return st, true
	}
	return googleSQLTypeForScalarValue(factory, v)
}
