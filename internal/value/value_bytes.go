package value

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type BytesValue []byte

func (bv BytesValue) Add(v Value) (Value, error) {
	v2, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	return BytesValue(append([]byte(bv), v2...)), nil
}

func (bv BytesValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("sub operation is unsupported for bytes %v", bv)
}

func (bv BytesValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for bytes %v", bv)
}

func (bv BytesValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for bytes %v", bv)
}

func (bv BytesValue) EQ(v Value) (bool, error) {
	v2, err := v.ToBytes()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to bytes", v)
	}
	return bytes.Equal([]byte(bv), v2), nil
}

func (bv BytesValue) GT(v Value) (bool, error) {
	v2, err := v.ToBytes()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to bytes", v)
	}
	return string(bv) > string(v2), nil
}

func (bv BytesValue) GTE(v Value) (bool, error) {
	v2, err := v.ToBytes()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to bytes", v)
	}
	return string(bv) >= string(v2), nil
}

func (bv BytesValue) LT(v Value) (bool, error) {
	v2, err := v.ToBytes()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to bytes", v)
	}
	return string(bv) < string(v2), nil
}

func (bv BytesValue) LTE(v Value) (bool, error) {
	v2, err := v.ToBytes()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to bytes", v)
	}
	return string(bv) <= string(v2), nil
}

func (bv BytesValue) ToInt64() (int64, error) {
	if len(bv) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(string(bv), 10, 64)
}

func (bv BytesValue) ToString() (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(bv)), nil
}

func (bv BytesValue) ToBytes() ([]byte, error) {
	return []byte(bv), nil
}

func (bv BytesValue) ToFloat64() (float64, error) {
	if len(bv) == 0 {
		return 0, nil
	}
	return strconv.ParseFloat(string(bv), 64)
}

func (bv BytesValue) ToBool() (bool, error) {
	if len(bv) == 0 {
		return false, nil
	}
	return strconv.ParseBool(string(bv))
}

func (bv BytesValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert array from bytes: %v", bv)
}

func (bv BytesValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert struct from bytes: %v", bv)
}

func (bv BytesValue) ToJSON() (string, error) {
	v, err := bv.ToString()
	if err != nil {
		return "", err
	}
	return strconv.Quote(v), nil
}

func (bv BytesValue) ToTime() (time.Time, error) {
	raw := string(bv)
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
	return time.Time{}, fmt.Errorf("failed to convert time.Time from bytes: %s", bv)
}

func (bv BytesValue) ToRat() (*big.Rat, error) {
	r := new(big.Rat)
	r.SetString(string(bv))
	return r, nil
}

func printableChar(v byte) bool {
	if 0x20 <= v && v <= 0x7e {
		return true
	}
	return false
}

func (bv BytesValue) Format(verb rune) string {
	switch verb {
	case 't':
		var ret strings.Builder
		for _, b := range bv {
			if printableChar(b) {
				fmt.Fprintf(&ret, "%c", b)
			} else {
				fmt.Fprintf(&ret, "\\x%02x", b)
			}
		}
		return ret.String()
	case 'T':
		var ret strings.Builder
		ret.WriteString(`b"`)
		for _, b := range bv {
			if printableChar(b) {
				fmt.Fprintf(&ret, "%c", b)
			} else {
				fmt.Fprintf(&ret, "\\x%02x", b)
			}
		}
		ret.WriteString(`"`)
		return ret.String()
	}
	v, _ := bv.ToString()
	return v
}

func (bv BytesValue) Interface() any {
	return []byte(bv)
}
