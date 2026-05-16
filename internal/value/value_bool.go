package value

import (
	"fmt"
	"math/big"
	"time"
)

type BoolValue bool

func (bv BoolValue) Add(v Value) (Value, error) {
	return nil, fmt.Errorf("add operation is unsupported for bool %v", bv)
}

func (bv BoolValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("sub operation is unsupported for bool %v", bv)
}

func (bv BoolValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for bool %v", bv)
}

func (bv BoolValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for bool %v", bv)
}

func (bv BoolValue) EQ(v Value) (bool, error) {
	v2, err := v.ToBool()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to bool", v)
	}
	return bool(bv) == v2, nil
}

func (bv BoolValue) GT(v Value) (bool, error) {
	return false, fmt.Errorf("gt operation is unsupported for bool %v", bv)
}

func (bv BoolValue) GTE(v Value) (bool, error) {
	return false, fmt.Errorf("gte operation is unsupported for bool %v", bv)
}

func (bv BoolValue) LT(v Value) (bool, error) {
	return false, fmt.Errorf("lt operation is unsupported for bool %v", bv)
}

func (bv BoolValue) LTE(v Value) (bool, error) {
	return false, fmt.Errorf("lte operation is unsupported for bool %v", bv)
}

func (bv BoolValue) ToInt64() (int64, error) {
	if bv {
		return 1, nil
	}
	return 0, nil
}

func (bv BoolValue) ToString() (string, error) {
	return fmt.Sprint(bv), nil
}

func (bv BoolValue) ToBytes() ([]byte, error) {
	return []byte(fmt.Sprint(bv)), nil
}

func (bv BoolValue) ToFloat64() (float64, error) {
	if bv {
		return 1, nil
	}
	return 0, nil
}

func (bv BoolValue) ToBool() (bool, error) {
	return bool(bv), nil
}

func (bv BoolValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert bool from array: %v", bv)
}

func (bv BoolValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert bool from struct: %v", bv)
}

func (bv BoolValue) ToJSON() (string, error) {
	return fmt.Sprint(bv), nil
}

func (bv BoolValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("failed to convert bool from time.Time: %v", bv)
}

func (bv BoolValue) ToRat() (*big.Rat, error) {
	r := new(big.Rat)
	if bv {
		r.SetInt64(1)
		return r, nil
	}
	r.SetInt64(0)
	return r, nil
}

func (bv BoolValue) Format(verb rune) string {
	return fmt.Sprint(bv)
}

func (bv BoolValue) Interface() any {
	return bool(bv)
}
