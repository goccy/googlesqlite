package value

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type StringValue string

func (sv StringValue) Add(v Value) (Value, error) {
	v2, err := v.ToString()
	if err != nil {
		return nil, err
	}
	return StringValue(string(sv) + v2), nil
}

func (sv StringValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("sub operation is unsupported for string %v", sv)
}

func (sv StringValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for string %v", sv)
}

func (sv StringValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for string %v", sv)
}

func (sv StringValue) EQ(v Value) (bool, error) {
	v2, err := v.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to string", v)
	}
	return string(sv) == v2, nil
}

func (sv StringValue) GT(v Value) (bool, error) {
	v2, err := v.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to string", v)
	}
	return string(sv) > v2, nil
}

func (sv StringValue) GTE(v Value) (bool, error) {
	v2, err := v.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to string", v)
	}
	return string(sv) >= v2, nil
}

func (sv StringValue) LT(v Value) (bool, error) {
	v2, err := v.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to string", v)
	}
	return string(sv) < v2, nil
}

func (sv StringValue) LTE(v Value) (bool, error) {
	v2, err := v.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to string", v)
	}
	return string(sv) <= v2, nil
}

func (sv StringValue) ToInt64() (int64, error) {
	if sv == "" {
		return 0, nil
	}
	toParse := string(sv)
	base := 10
	if strings.Contains(strings.ToLower(toParse), "0x") {
		base = 0
	}
	return strconv.ParseInt(toParse, base, 64)
}

func (sv StringValue) ToString() (string, error) {
	return string(sv), nil
}

func (sv StringValue) ToBytes() ([]byte, error) {
	return []byte(string(sv)), nil
}

func (sv StringValue) ToFloat64() (float64, error) {
	if sv == "" {
		return 0, nil
	}
	return strconv.ParseFloat(string(sv), 64)
}

func (sv StringValue) ToBool() (bool, error) {
	if sv == "" {
		return false, nil
	}
	return strconv.ParseBool(string(sv))
}

func (sv StringValue) ToArray() (*ArrayValue, error) {
	if sv == "" {
		return nil, nil
	}
	return nil, fmt.Errorf("failed to convert array from string: %v", sv)
}

func (sv StringValue) ToStruct() (*StructValue, error) {
	if sv == "" {
		return nil, nil
	}
	return nil, fmt.Errorf("failed to convert struct from string: %v", sv)
}

func (sv StringValue) ToJSON() (string, error) {
	return strconv.Quote(string(sv)), nil
}

func (sv StringValue) ToTime() (time.Time, error) {
	raw := string(sv)
	switch {
	case isDate(raw):
		return parseDate(raw)
	case isDatetime(raw):
		return parseDatetime(raw)
	case isTime(raw):
		return parseTime(raw)
	case isTimestamp(raw):
		return parseTimestamp(raw, time.UTC)
	}
	return time.Time{}, fmt.Errorf("failed to convert %s to time.Time type", sv)
}

func (sv StringValue) ToRat() (*big.Rat, error) {
	r := new(big.Rat)
	r.SetString(string(sv))
	return r, nil
}

func (sv StringValue) Format(verb rune) string {
	switch verb {
	case 't':
		return string(sv)
	case 'T':
		return strconv.Quote(string(sv))
	}
	return string(sv)
}

func (sv StringValue) Interface() any {
	return string(sv)
}
