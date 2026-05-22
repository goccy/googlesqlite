package value

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/goccy/googlesqlite/internal/intervalvalue"
)

type IntervalValue struct {
	*intervalvalue.IntervalValue
}

func (iv *IntervalValue) Add(v Value) (Value, error) {
	return nil, fmt.Errorf("unsupported add operator for interval value")
}

func (iv *IntervalValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("unsupported sub operator for interval value")
}

func (iv *IntervalValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("unsupported mul operator for interval value")
}

func (iv *IntervalValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("unsupported div operator for interval value")
}

func (iv *IntervalValue) EQ(v Value) (bool, error) {
	return false, fmt.Errorf("unsupported eq operator for interval value")
}

func (iv *IntervalValue) GT(v Value) (bool, error) {
	return false, fmt.Errorf("unsupported gt operator for interval value")
}

func (iv *IntervalValue) GTE(v Value) (bool, error) {
	return false, fmt.Errorf("unsupporte gte operator for interval value")
}

func (iv *IntervalValue) LT(v Value) (bool, error) {
	return false, fmt.Errorf("unsupported lt operator for interval value")
}

func (iv *IntervalValue) LTE(v Value) (bool, error) {
	return false, fmt.Errorf("unsupported lte operator for interval value")
}

func (iv *IntervalValue) ToInt64() (int64, error) {
	return 0, fmt.Errorf("unsupported int64 cast for interval value")
}

func (iv *IntervalValue) ToString() (string, error) {
	if iv.Years == 0 && iv.Months < 0 {
		return "-" + iv.String(), nil
	}
	return iv.String(), nil
}

func (iv *IntervalValue) ToBytes() ([]byte, error) {
	s, err := iv.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (iv *IntervalValue) ToFloat64() (float64, error) {
	return 0, fmt.Errorf("unsupported float64 cast for interval value")
}

func (iv *IntervalValue) ToBool() (bool, error) {
	return false, fmt.Errorf("unsupported bool cast for interval value")
}

func (iv *IntervalValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("unsupported array cast for interval value")
}

func (iv *IntervalValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("unsupported struct cast for interval value")
}

func (iv *IntervalValue) ToJSON() (string, error) {
	s, err := iv.ToString()
	if err != nil {
		return "", err
	}
	return strconv.Quote(s), nil
}

func (iv *IntervalValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("unsupported time cast for interval value")
}

func (iv *IntervalValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("unsupported numeric cast for interval value")
}

func (iv *IntervalValue) Format(verb rune) string {
	s, err := iv.ToString()
	if err != nil {
		return ""
	}
	return s
}

func (iv *IntervalValue) Interface() any {
	s, err := iv.ToString()
	if err != nil {
		return nil
	}
	return s
}
