package value

import (
	"fmt"
	"math/big"
	"time"
)

type FloatValue float64

func (fv FloatValue) Add(v Value) (Value, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return nil, err
	}
	return FloatValue(float64(fv) + v2), nil
}

func (fv FloatValue) Sub(v Value) (Value, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return nil, err
	}
	return FloatValue(float64(fv) - v2), nil
}

func (fv FloatValue) Mul(v Value) (Value, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return nil, err
	}
	return FloatValue(float64(fv) * v2), nil
}

func (fv FloatValue) Div(v Value) (Value, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return nil, err
	}
	if v2 == 0 {
		return nil, fmt.Errorf("zero divided error ( %f / 0 )", fv)
	}
	return FloatValue(float64(fv) / v2), nil
}

func (fv FloatValue) EQ(v Value) (bool, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to float64", v)
	}
	return float64(fv) == v2, nil
}

func (fv FloatValue) GT(v Value) (bool, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to float64", v)
	}
	return float64(fv) > v2, nil
}

func (fv FloatValue) GTE(v Value) (bool, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to float64", v)
	}
	return float64(fv) >= v2, nil
}

func (fv FloatValue) LT(v Value) (bool, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to float64", v)
	}
	return float64(fv) < v2, nil
}

func (fv FloatValue) LTE(v Value) (bool, error) {
	v2, err := v.ToFloat64()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to float64", v)
	}
	return float64(fv) <= v2, nil
}

func (fv FloatValue) ToInt64() (int64, error) {
	return int64(fv), nil
}

func (fv FloatValue) ToString() (string, error) {
	return fmt.Sprint(fv), nil
}

func (fv FloatValue) ToBytes() ([]byte, error) {
	return []byte(fmt.Sprint(fv)), nil
}

func (fv FloatValue) ToFloat64() (float64, error) {
	return float64(fv), nil
}

func (fv FloatValue) ToBool() (bool, error) {
	switch fmt.Sprint(fv) {
	case "1":
		return true, nil
	case "0":
		return false, nil
	}
	return false, fmt.Errorf("failed to convert %f to bool type", fv)
}

func (fv FloatValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert array from float64: %v", fv)
}

func (fv FloatValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert struct from float64: %v", fv)
}

func (fv FloatValue) ToJSON() (string, error) {
	return fmt.Sprint(fv), nil
}

func (fv FloatValue) ToTime() (time.Time, error) {
	return TimestampFromFloatValue(float64(fv))
}

func (fv FloatValue) ToRat() (*big.Rat, error) {
	r := new(big.Rat)
	r.SetFloat64(float64(fv))
	return r, nil
}

func (fv FloatValue) Format(verb rune) string {
	return fmt.Sprint(fv)
}

func (fv FloatValue) Interface() any {
	return float64(fv)
}
