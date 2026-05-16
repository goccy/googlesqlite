package value

import (
	"math/big"
	"regexp"
	"time"
)

type SafeValue struct {
	Value Value
}

func (v *SafeValue) Add(arg Value) (Value, error) {
	ret, err := v.Value.Add(arg)
	if err != nil {
		return nil, nil
	}
	return ret, nil
}

func (v *SafeValue) Sub(arg Value) (Value, error) {
	ret, err := v.Value.Sub(arg)
	if err != nil {
		return nil, nil
	}
	return ret, nil
}

func (v *SafeValue) Mul(arg Value) (Value, error) {
	ret, err := v.Value.Mul(arg)
	if err != nil {
		return nil, nil
	}
	return ret, nil
}

func (v *SafeValue) Div(arg Value) (Value, error) {
	ret, err := v.Value.Div(arg)
	if err != nil {
		return nil, nil
	}
	return ret, nil
}

func (v *SafeValue) EQ(arg Value) (bool, error) {
	ret, err := v.Value.EQ(arg)
	if err != nil {
		return false, nil
	}
	return ret, nil
}

func (v *SafeValue) GT(arg Value) (bool, error) {
	ret, err := v.Value.GT(arg)
	if err != nil {
		return false, nil
	}
	return ret, nil
}

func (v *SafeValue) GTE(arg Value) (bool, error) {
	ret, err := v.Value.GTE(arg)
	if err != nil {
		return false, nil
	}
	return ret, nil
}

func (v *SafeValue) LT(arg Value) (bool, error) {
	ret, err := v.Value.LT(arg)
	if err != nil {
		return false, nil
	}
	return ret, nil
}

func (v *SafeValue) LTE(arg Value) (bool, error) {
	ret, err := v.Value.LTE(arg)
	if err != nil {
		return false, nil
	}
	return ret, nil
}

func (v *SafeValue) ToInt64() (int64, error) {
	ret, err := v.Value.ToInt64()
	if err != nil {
		return 0, nil
	}
	return ret, nil
}

func (v *SafeValue) ToString() (string, error) {
	ret, err := v.Value.ToString()
	if err != nil {
		return "", nil
	}
	return ret, nil
}

func (v *SafeValue) ToBytes() ([]byte, error) {
	ret, err := v.Value.ToBytes()
	if err != nil {
		return nil, nil
	}
	return ret, nil
}

func (v *SafeValue) ToFloat64() (float64, error) {
	ret, err := v.Value.ToFloat64()
	if err != nil {
		return 0, nil
	}
	return ret, nil
}

func (v *SafeValue) ToBool() (bool, error) {
	ret, err := v.Value.ToBool()
	if err != nil {
		return false, nil
	}
	return ret, nil
}

func (v *SafeValue) ToArray() (*ArrayValue, error) {
	ret, err := v.Value.ToArray()
	if err != nil {
		return &ArrayValue{}, nil
	}
	return ret, nil
}

func (v *SafeValue) ToStruct() (*StructValue, error) {
	ret, err := v.Value.ToStruct()
	if err != nil {
		return &StructValue{}, nil
	}
	return ret, nil
}

func (v *SafeValue) ToJSON() (string, error) {
	ret, err := v.Value.ToJSON()
	if err != nil {
		return "", nil
	}
	return ret, nil
}

func (v *SafeValue) ToTime() (time.Time, error) {
	ret, err := v.Value.ToTime()
	if err != nil {
		return time.Time{}, nil
	}
	return ret, nil
}

func (v *SafeValue) ToRat() (*big.Rat, error) {
	ret, err := v.Value.ToRat()
	if err != nil {
		return nil, nil
	}
	return ret, nil
}

func (v *SafeValue) Format(verb rune) string {
	return v.Value.Format(verb)
}

func (v *SafeValue) Interface() any {
	return v.Value.Interface()
}

var (
	dateRe     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	datetimeRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}$`)
	timeRe     = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}`)
)
