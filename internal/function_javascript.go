package internal

import (
	"fmt"
	"math/big"
	"time"

	"github.com/dop251/goja"
	googlesql "github.com/goccy/go-googlesql"

	"github.com/goccy/googlesqlite/internal/value"
)

func EVAL_JAVASCRIPT(code string, retType *Type, argNames []string, args []value.Value) (value.Value, error) {
	vm := goja.New()
	for i := range args {
		var v any
		if args[i] != nil {
			structV, ok := args[i].(*value.StructValue)
			if ok {
				v = structV.M
			} else {
				v = args[i].Interface()
			}
		}
		if err := vm.Set(argNames[i], v); err != nil {
			return nil, fmt.Errorf(
				"failed to set argument variable for %s as %v",
				argNames[i],
				args[i],
			)
		}
	}
	evalCode := fmt.Sprintf(`
function googlesqlite_javascript_func() { %s }
googlesqlite_javascript_func();
`, code)
	ret, err := vm.RunString(evalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate javascript code %s: %w", code, err)
	}
	typ, err := retType.ToGoogleSQLType()
	if err != nil {
		return nil, fmt.Errorf("failed to get return type: %w", err)
	}
	value, err := castJavaScriptValue(typ, ret)
	if err != nil {
		return nil, fmt.Errorf("failed to convert googlesqlite value from %v: %w", ret, err)
	}
	return value, nil
}

func castJavaScriptValue(t googlesql.Googlesql_TypeNode, v goja.Value) (value.Value, error) {
	if v == nil {
		return nil, nil
	}
	// Googlesql_TypeNode carries Kind directly.
	switch m1(t.Kind()) {
	case googlesql.TypeKindTypeInt32, googlesql.TypeKindTypeInt64, googlesql.TypeKindTypeUint32, googlesql.TypeKindTypeUint64:
		return value.IntValue(v.ToInteger()), nil
	case googlesql.TypeKindTypeBool:
		return value.BoolValue(v.ToBoolean()), nil
	case googlesql.TypeKindTypeFloat, googlesql.TypeKindTypeDouble:
		return value.FloatValue(v.ToFloat()), nil
	case googlesql.TypeKindTypeString, googlesql.TypeKindTypeEnum:
		return value.StringValue(v.ToString().String()), nil
	case googlesql.TypeKindTypeBytes:
		return value.BytesValue(v.ToString().String()), nil
	case googlesql.TypeKindTypeDate:
		t, err := value.ParseDate(v.ToString().String())
		if err != nil {
			return nil, err
		}
		return value.DateValue(t), nil
	case googlesql.TypeKindTypeDatetime:
		t, err := value.ParseDatetime(v.ToString().String())
		if err != nil {
			return nil, err
		}
		return value.DatetimeValue(t), nil
	case googlesql.TypeKindTypeTime:
		t, err := value.ParseTime(v.ToString().String())
		if err != nil {
			return nil, err
		}
		return value.TimeValue(t), nil
	case googlesql.TypeKindTypeTimestamp:
		t, err := value.ParseTimestamp(v.ToString().String(), time.UTC)
		if err != nil {
			return nil, err
		}
		return value.TimestampValue(t), nil
	case googlesql.TypeKindTypeInterval:
		return value.ParseInterval(v.ToString().String())
	case googlesql.TypeKindTypeNumeric:
		r := new(big.Rat)
		r.SetString(v.ToNumber().String())
		return &value.NumericValue{Rat: r}, nil
	case googlesql.TypeKindTypeBignumeric:
		r := new(big.Rat)
		r.SetString(v.ToNumber().String())
		return &value.NumericValue{Rat: r}, nil
	case googlesql.TypeKindTypeJson:
		return value.JsonValue(v.ToString().String()), nil
	case googlesql.TypeKindTypeArray:
		// ArrayType.ElementType isn't exposed via the bridge; cast
		// each element through without type info.
		var ret value.ArrayValue
		for _, vv := range v.Export().([]any) {
			base, err := ValueFromGoValue(vv)
			if err != nil {
				return nil, err
			}
			ret.Values = append(ret.Values, base)
		}
		return &ret, nil
	case googlesql.TypeKindTypeStruct:
		base, err := ValueFromGoValue(v.Export())
		if err != nil {
			return nil, err
		}
		return CastValue(t, base)
	case googlesql.TypeKindTypeGeography:
		base, err := ValueFromGoValue(v.Export())
		if err != nil {
			return nil, err
		}
		return CastValue(t, base)
	}
	return nil, fmt.Errorf("unsupported cast %v from JavaScript value", m1(t.Kind()))
}
