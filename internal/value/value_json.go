package value

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/goccy/go-json"
)

type JsonValue string

func (jv JsonValue) Add(v Value) (Value, error) {
	return nil, fmt.Errorf("add operation is unsupported for json %v", jv)
}

func (jv JsonValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("sub operation is unsupported for json %v", jv)
}

func (jv JsonValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for json %v", jv)
}

func (jv JsonValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for json %v", jv)
}

func (jv JsonValue) EQ(v Value) (bool, error) {
	return false, fmt.Errorf("eq operation is unsupported for json %v", jv)
}

func (jv JsonValue) GT(v Value) (bool, error) {
	return false, fmt.Errorf("gt operation is unsupported for json %v", jv)
}

func (jv JsonValue) GTE(v Value) (bool, error) {
	return false, fmt.Errorf("gte operation is unsupported for json %v", jv)
}

func (jv JsonValue) LT(v Value) (bool, error) {
	return false, fmt.Errorf("lt operation is unsupported for json %v", jv)
}

func (jv JsonValue) LTE(v Value) (bool, error) {
	return false, fmt.Errorf("lte operation is unsupported for json %v", jv)
}

func (jv JsonValue) ToInt64() (int64, error) {
	return strconv.ParseInt(string(jv), 0, 64)
}

func (jv JsonValue) ToString() (string, error) {
	return string(jv), nil
}

func (jv JsonValue) ToBytes() ([]byte, error) {
	return []byte(string(jv)), nil
}

func (jv JsonValue) ToFloat64() (float64, error) {
	return strconv.ParseFloat(string(jv), 64)
}

func (jv JsonValue) ToBool() (bool, error) {
	return strconv.ParseBool(string(jv))
}

func (jv JsonValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert json from array: %v", jv)
}

func (jv JsonValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert json from struct: %v", jv)
}

func (jv JsonValue) ToJSON() (string, error) {
	return string(jv), nil
}

func (jv JsonValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("failed to convert json from time.Time: %v", jv)
}

func (jv JsonValue) ToRat() (*big.Rat, error) {
	i64, err := strconv.ParseInt(string(jv), 0, 64)
	if err != nil {
		return nil, err
	}
	r := new(big.Rat)
	r.SetInt64(i64)
	return r, nil
}

func (jv JsonValue) Format(verb rune) string {
	return string(jv)
}

func (jv JsonValue) Interface() any {
	var v any
	if err := json.Unmarshal([]byte(jv), &v); err != nil {
		return nil
	}
	return v
}

func (jv JsonValue) reflectTypeToJsonType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	case reflect.Pointer:
		return jv.reflectTypeToJsonType(t.Elem())
	}
	return "unknown"
}

func (jv JsonValue) Type() string {
	if string(jv) == "null" {
		return "null"
	}
	rv := reflect.ValueOf(jv.Interface())
	return jv.reflectTypeToJsonType(rv.Type())
}
