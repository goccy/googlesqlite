package value

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// jsonUnmarshalPreserveNumbers decodes JSON into v with json.Number so
// integer literals don't collapse to float64 when landing in an
// interface{} slot. The post-pass rewrites json.Number nodes to int64
// or float64 based on the source shape (presence of '.' or 'e').
func jsonUnmarshalPreserveNumbers(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

// coerceJSONNumber converts a json.Number into int64 (no decimal /
// exponent) or float64 (otherwise). Non-json.Number values pass
// through untouched.
func coerceJSONNumber(v any) any {
	n, ok := v.(json.Number)
	if !ok {
		return v
	}
	s := string(n)
	if !strings.ContainsAny(s, ".eE") {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return v
}

// ConvertArgs decodes a sequence of SQLite-side Go values (the form
// the ncruces driver hands us in Step / scalar callbacks) into a slice
// of googlesqlite Values. It is a thin loop around DecodeValue.
func ConvertArgs(args ...any) ([]Value, error) {
	values := make([]Value, 0, len(args))
	for _, arg := range args {
		v, err := DecodeValue(arg)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

// DecodeValue is the inverse of EncodeValue: given the over-the-wire
// form (Go primitive or base64-encoded ValueLayout JSON) produce the
// matching googlesqlite Value.
func DecodeValue(v any) (Value, error) {
	if isNullValue(v) {
		return nil, nil
	}
	v = coerceJSONNumber(v)
	switch vv := v.(type) {
	case int64:
		return IntValue(vv), nil
	case float64:
		return FloatValue(vv), nil
	case bool:
		return BoolValue(vv), nil
	}
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected value type: %T", v)
	}
	// Try the canonical base64-of-JSON envelope first. If either step
	// fails the input is most likely a raw SQL string literal (e.g.
	// `'int32'`, `'date'`, `'bytes'`) that the formatter inlined
	// without the envelope. Fall back to a plain StringValue so the
	// receiver can still call .ToString() on it.
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return StringValue(s), nil
	}
	var layout ValueLayout
	if err := json.Unmarshal(decoded, &layout); err != nil {
		return StringValue(s), nil
	}
	if layout.Header == "" {
		return StringValue(s), nil
	}
	return decodeFromValueLayout(&layout)
}

func decodeFromValueLayout(layout *ValueLayout) (Value, error) {
	switch layout.Header {
	case StringValueType:
		return StringValue(layout.Body), nil
	case BytesValueType:
		decoded, err := base64.StdEncoding.DecodeString(layout.Body)
		if err != nil {
			return nil, err
		}
		return BytesValue(decoded), nil
	case NumericValueType:
		r := new(big.Rat)
		r.SetString(layout.Body)
		return &NumericValue{Rat: r}, nil
	case BigNumericValueType:
		r := new(big.Rat)
		r.SetString(layout.Body)
		return &NumericValue{Rat: r, IsBigNumeric: true}, nil
	case DateValueType:
		t, err := parseDate(layout.Body)
		if err != nil {
			return nil, err
		}
		return DateValue(t), nil
	case DatetimeValueType:
		t, err := parseDatetime(layout.Body)
		if err != nil {
			return nil, err
		}
		return DatetimeValue(t), nil
	case TimeValueType:
		t, err := parseTime(layout.Body)
		if err != nil {
			return nil, err
		}
		return TimeValue(t), nil
	case TimestampValueType:
		microsec, err := strconv.ParseInt(layout.Body, 10, 64)
		microSecondsInSecond := int64(time.Second) / int64(time.Microsecond)
		sec := microsec / microSecondsInSecond
		remainder := microsec - (sec * microSecondsInSecond)
		if err != nil {
			return nil, fmt.Errorf("failed to parse unixmicro for timestamp value %s: %w", layout.Body, err)
		}
		return TimestampValue(time.Unix(sec, remainder*int64(time.Microsecond))), nil
	case IntervalValueType:
		return parseInterval(layout.Body)
	case JsonValueType:
		return JsonValue(layout.Body), nil
	case ArrayValueType:
		var arr []any
		if err := jsonUnmarshalPreserveNumbers([]byte(layout.Body), &arr); err != nil {
			return nil, fmt.Errorf("failed to decode array body: %w", err)
		}
		ret := &ArrayValue{
			Values: make([]Value, 0, len(arr)),
		}
		for _, elem := range arr {
			val, err := DecodeValue(elem)
			if err != nil {
				return nil, err
			}
			ret.Values = append(ret.Values, val)
		}
		return ret, nil
	case StructValueType:
		var structLayout StructValueLayout
		if err := jsonUnmarshalPreserveNumbers([]byte(layout.Body), &structLayout); err != nil {
			return nil, err
		}
		m := map[string]Value{}
		values := make([]Value, 0, len(structLayout.Values))
		for i, data := range structLayout.Values {
			val, err := DecodeValue(data)
			if err != nil {
				return nil, err
			}
			m[structLayout.Keys[i]] = val
			values = append(values, val)
		}
		ret := &StructValue{}
		ret.Keys = structLayout.Keys
		ret.Values = values
		ret.M = m
		return ret, nil
	case GeographyValueType:
		body := layout.Body
		inverted := false
		if strings.HasPrefix(body, "INVERTED ") {
			inverted = true
			body = body[len("INVERTED "):]
		}
		ret, err := GeographyFromWKT(body)
		if err != nil {
			return nil, fmt.Errorf("decodeFromValueLayout failed: %w", err)
		}
		if inverted {
			ret.MarkInverted()
		}
		return ret, nil
	case RangeValueType:
		var rl struct {
			Elem  string `json:"elem"`
			Start any    `json:"start"`
			End   any    `json:"end"`
		}
		if err := jsonUnmarshalPreserveNumbers([]byte(layout.Body), &rl); err != nil {
			return nil, fmt.Errorf("failed to decode range body: %w", err)
		}
		ret := &RangeValue{ElemHeader: ValueType(rl.Elem)}
		if rl.Start != nil {
			s, err := DecodeValue(rl.Start)
			if err != nil {
				return nil, err
			}
			ret.Start = s
		}
		if rl.End != nil {
			e, err := DecodeValue(rl.End)
			if err != nil {
				return nil, err
			}
			ret.End = e
		}
		return ret, nil
	}
	return nil, fmt.Errorf("unexpected value header: %s", layout.Header)
}
