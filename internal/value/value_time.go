package value

import (
	"fmt"
	"math/big"
	"time"
)

type TimeValue time.Time

func (t TimeValue) Add(v Value) (Value, error) {
	return nil, fmt.Errorf("add operation is unsupported for time %v", t)
}

func (t TimeValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("sub operation is unsupported for time %v", t)
}

func (t TimeValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for time %v", t)
}

func (t TimeValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for time %v", t)
}

func (t TimeValue) EQ(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Equal(v2), nil
}

func (t TimeValue) GT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).After(v2), nil
}

func (t TimeValue) GTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Equal(v2) || time.Time(t).After(v2), nil
}

func (t TimeValue) LT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Before(v2), nil
}

func (t TimeValue) LTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Equal(v2) || time.Time(t).Before(v2), nil
}

func (t TimeValue) ToInt64() (int64, error) {
	return time.Time(t).Unix(), nil
}

func (t TimeValue) ToString() (string, error) {
	return time.Time(t).Format("15:04:05.999999"), nil
}

func (t TimeValue) ToBytes() ([]byte, error) {
	v, err := t.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (t TimeValue) ToFloat64() (float64, error) {
	return float64(time.Time(t).Unix()), nil
}

func (t TimeValue) ToBool() (bool, error) {
	return false, fmt.Errorf("failed to convert %v to bool type", t)
}

func (t TimeValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert %v to array type", t)
}

func (t TimeValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert %v to struct type", t)
}

func (t TimeValue) ToJSON() (string, error) {
	return t.ToString()
}

func (t TimeValue) ToTime() (time.Time, error) {
	return time.Time(t), nil
}

func (t TimeValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("failed to convert *big.Rat from time %v", t)
}

func (t TimeValue) Format(verb rune) string {
	formatted := time.Time(t).Format("15:04:05.999999")
	switch verb {
	case 't':
		return formatted
	case 'T':
		return fmt.Sprintf(`TIME %q`, formatted)
	}
	return formatted
}

func (t TimeValue) Interface() any {
	return time.Time(t).Format("15:04:05.999999")
}
