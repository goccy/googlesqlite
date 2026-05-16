package value

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	"github.com/goccy/go-json"
)

// EncodeValue serializes a Value into the over-the-wire form the
// SQLite layer accepts. INT64 / FLOAT64 / BOOL pass through to their
// Go primitives; richer types (BYTES, NUMERIC, ARRAY, STRUCT, ...)
// are JSON-encoded under a small (header, body) envelope and then
// base64-encoded so SQLite stores them as a flat string.
func EncodeValue(v Value) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch vv := v.(type) {
	case IntValue:
		return v.ToInt64()
	case FloatValue:
		return v.ToFloat64()
	case BoolValue:
		return v.ToBool()
	case *SafeValue:
		return EncodeValue(vv.Value)
	}
	layout, err := valueLayoutFromValue(v)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(layout)
	if err != nil {
		return nil, fmt.Errorf("failed to encode value: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// ValueFromGoValue converts a plain Go value (the kind a SQL driver
// hands us across the FFI boundary) into a googlesqlite Value.
func ValueFromGoValue(v any) (Value, error) {
	if isNullValue(v) {
		return nil, nil
	}
	return valueFromGoReflectValue(reflect.ValueOf(v))
}

func valueFromGoReflectValue(v reflect.Value) (Value, error) {
	kind := v.Type().Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return IntValue(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return IntValue(int64(v.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return FloatValue(v.Float()), nil
	case reflect.Bool:
		return BoolValue(v.Bool()), nil
	case reflect.String:
		return StringValue(v.String()), nil
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return BytesValue(v.Bytes()), nil
		}
		ret := &ArrayValue{}
		for i := 0; i < v.Len(); i++ {
			elem, err := valueFromGoReflectValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			ret.Values = append(ret.Values, elem)
		}
		return ret, nil
	case reflect.Map:
		ret := &StructValue{M: map[string]Value{}}
		iter := v.MapRange()
		for iter.Next() {
			key, err := valueFromGoReflectValue(iter.Key())
			if err != nil {
				return nil, err
			}
			k, err := key.ToString()
			if err != nil {
				return nil, err
			}
			val, err := valueFromGoReflectValue(iter.Value())
			if err != nil {
				return nil, err
			}
			ret.Keys = append(ret.Keys, k)
			ret.Values = append(ret.Values, val)
			ret.M[k] = val
		}
		return ret, nil
	case reflect.Struct:
		t, ok := v.Interface().(time.Time)
		if ok {
			return TimestampValue(t), nil
		}
		ret := &StructValue{M: map[string]Value{}}
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			key := typ.Field(i).Name
			val, err := valueFromGoReflectValue(v.Field(i))
			if err != nil {
				return nil, err
			}
			ret.Keys = append(ret.Keys, key)
			ret.Values = append(ret.Values, val)
			ret.M[key] = val
		}
		return ret, nil
	case reflect.Pointer:
		return valueFromGoReflectValue(v.Elem())
	case reflect.Interface:
		vv := v.Interface()
		if isNullValue(vv) {
			return nil, nil
		}
		return valueFromGoReflectValue(reflect.ValueOf(vv))
	}
	return nil, fmt.Errorf("cannot convert %s type to googlesqlite value type", kind)
}

// ValueLayoutFromValue is the public entry point for the layout
// builder used by the over-the-wire encoder and by literal rendering.
func ValueLayoutFromValue(v Value) (*ValueLayout, error) {
	return valueLayoutFromValue(v)
}

func valueLayoutFromValue(v Value) (*ValueLayout, error) {
	switch vv := v.(type) {
	case StringValue:
		return &ValueLayout{
			Header: StringValueType,
			Body:   string(vv),
		}, nil
	case BytesValue:
		return &ValueLayout{
			Header: BytesValueType,
			Body:   base64.StdEncoding.EncodeToString([]byte(vv)),
		}, nil
	case *NumericValue:
		b, err := vv.MarshalText()
		if err != nil {
			return nil, err
		}
		if vv.IsBigNumeric {
			return &ValueLayout{
				Header: BigNumericValueType,
				Body:   string(b),
			}, nil
		}
		return &ValueLayout{
			Header: NumericValueType,
			Body:   string(b),
		}, nil
	case DateValue:
		body, err := vv.ToString()
		if err != nil {
			return nil, err
		}
		return &ValueLayout{
			Header: DateValueType,
			Body:   body,
		}, nil
	case DatetimeValue:
		body, err := vv.ToString()
		if err != nil {
			return nil, err
		}
		return &ValueLayout{
			Header: DatetimeValueType,
			Body:   body,
		}, nil
	case TimeValue:
		body, err := vv.ToString()
		if err != nil {
			return nil, err
		}
		return &ValueLayout{
			Header: TimeValueType,
			Body:   body,
		}, nil
	case TimestampValue:
		return &ValueLayout{
			Header: TimestampValueType,
			Body:   fmt.Sprint(time.Time(vv).UnixMicro()),
		}, nil
	case *IntervalValue:
		s, err := vv.ToString()
		if err != nil {
			return nil, err
		}
		return &ValueLayout{
			Header: IntervalValueType,
			Body:   s,
		}, nil
	case JsonValue:
		return &ValueLayout{
			Header: JsonValueType,
			Body:   string(vv),
		}, nil
	case *ArrayValue:
		values := make([]any, 0, len(vv.Values))
		for _, v := range vv.Values {
			if v == nil {
				values = append(values, nil)
				continue
			}
			val, err := EncodeValue(v)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		body, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		return &ValueLayout{
			Header: ArrayValueType,
			Body:   string(body),
		}, nil
	case *StructValue:
		values := make([]any, 0, len(vv.Values))
		for _, v := range vv.Values {
			val, err := EncodeValue(v)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		body, err := json.Marshal(&StructValueLayout{
			Keys:   vv.Keys,
			Values: values,
		})
		if err != nil {
			return nil, err
		}
		return &ValueLayout{
			Header: StructValueType,
			Body:   string(body),
		}, nil
	case *GeographyValue:
		s, err := vv.ToWKT()
		if err != nil {
			return nil, err
		}
		// Inverted polygons (the `oriented => TRUE`, CW-shell case
		// where the interior is the complement of the ring) are not
		// expressible in plain WKT, so prefix the body with an
		// `INVERTED ` sentinel that the decoder strips and translates
		// back into a MarkInverted call.
		if vv.Inverted() {
			s = "INVERTED " + s
		}
		return &ValueLayout{
			Header: GeographyValueType,
			Body:   s,
		}, nil
	case *RangeValue:
		// Encode as { "elem": "<header>", "start": <encoded_or_null>,
		// "end": <encoded_or_null> } so the decoder can rebuild the
		// typed bounds without re-inferring the element type.
		var startEnc, endEnc any
		if vv.Start != nil {
			s, err := EncodeValue(vv.Start)
			if err != nil {
				return nil, err
			}
			startEnc = s
		}
		if vv.End != nil {
			e, err := EncodeValue(vv.End)
			if err != nil {
				return nil, err
			}
			endEnc = e
		}
		body, err := json.Marshal(map[string]any{
			"elem":  string(vv.ElemHeader),
			"start": startEnc,
			"end":   endEnc,
		})
		if err != nil {
			return nil, err
		}
		return &ValueLayout{
			Header: RangeValueType,
			Body:   string(body),
		}, nil
	}
	return nil, fmt.Errorf("unexpected value type for layout: %T", v)
}
