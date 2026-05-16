package value

import (
	"fmt"
	"math/big"
	"time"
)

type IntValue int64

func (iv IntValue) Add(v Value) (Value, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	return IntValue(int64(iv) + v2), nil
}

func (iv IntValue) Sub(v Value) (Value, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	return IntValue(int64(iv) - v2), nil
}

func (iv IntValue) Mul(v Value) (Value, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	return IntValue(int64(iv) * v2), nil
}

func (iv IntValue) Div(v Value) (Value, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	if v2 == 0 {
		return nil, fmt.Errorf("zero divided error ( %d / 0 )", iv)
	}
	return IntValue(int64(iv) / v2), nil
}

func (iv IntValue) EQ(v Value) (bool, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to int64", v)
	}
	return int64(iv) == v2, nil
}

func (iv IntValue) GT(v Value) (bool, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to int64", v)
	}
	return int64(iv) > v2, nil
}

func (iv IntValue) GTE(v Value) (bool, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to int64", v)
	}
	return int64(iv) >= v2, nil
}

func (iv IntValue) LT(v Value) (bool, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to int64", v)
	}
	return int64(iv) < v2, nil
}

func (iv IntValue) LTE(v Value) (bool, error) {
	v2, err := v.ToInt64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to int64", v)
	}
	return int64(iv) <= v2, nil
}

func (iv IntValue) ToInt64() (int64, error) {
	return int64(iv), nil
}

func (iv IntValue) ToString() (string, error) {
	return fmt.Sprint(iv), nil
}

func (iv IntValue) ToBytes() ([]byte, error) {
	return []byte(fmt.Sprint(iv)), nil
}

func (iv IntValue) ToFloat64() (float64, error) {
	return float64(iv), nil
}

func (iv IntValue) ToBool() (bool, error) {
	switch iv {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("failed to convert %d to bool type", iv)
	}
}

func (iv IntValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert %d to array type", iv)
}

func (iv IntValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert %d to struct type", iv)
}

func (iv IntValue) ToJSON() (string, error) {
	return fmt.Sprint(iv), nil
}

func (iv IntValue) ToTime() (time.Time, error) {
	v := int64(iv)
	if v > time.Unix(0, 0).Unix()*int64(time.Millisecond) {
		return TimestampFromInt64Value(v)
	}
	return DateFromInt64Value(v)
}

func (iv IntValue) ToRat() (*big.Rat, error) {
	r := new(big.Rat)
	r.SetInt64(int64(iv))
	return r, nil
}

func (iv IntValue) Format(verb rune) string {
	return fmt.Sprint(iv)
}

func (iv IntValue) Interface() any {
	return int64(iv)
}
