package value

import (
	"fmt"
	"math/big"
	"strings"
	"time"
)

type NumericValue struct {
	*big.Rat
	IsBigNumeric bool
}

func (nv *NumericValue) Add(v Value) (Value, error) {
	z := new(big.Rat)
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return nil, err
	}
	nv.Rat = z.Add(x, y)
	return nv, nil
}

func (nv *NumericValue) Sub(v Value) (Value, error) {
	z := new(big.Rat)
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return nil, err
	}
	zy := new(big.Rat)
	nv.Rat = z.Add(x, zy.Neg(y))
	return nv, nil
}

func (nv *NumericValue) Mul(v Value) (Value, error) {
	z := new(big.Rat)
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return nil, err
	}
	nv.Rat = z.Mul(x, y)
	return nv, nil
}

func (nv *NumericValue) Div(v Value) (ret Value, e error) {
	defer func() {
		if r := recover(); r != nil {
			// big.Rat.Inv panics with a plain string on division by
			// zero, not an error value, so route both kinds through
			// fmt.Errorf rather than asserting the concrete type.
			e = fmt.Errorf("NumericValue.Div: %v", r)
		}
	}()
	z := new(big.Rat)
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return nil, err
	}
	zy := new(big.Rat)
	nv.Rat = z.Mul(x, zy.Inv(y))
	return nv, nil
}

func (nv *NumericValue) EQ(v Value) (bool, error) {
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return false, err
	}
	return x.Cmp(y) == 0, nil
}

func (nv *NumericValue) GT(v Value) (bool, error) {
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return false, err
	}
	return x.Cmp(y) > 0, nil
}

func (nv *NumericValue) GTE(v Value) (bool, error) {
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return false, err
	}
	return x.Cmp(y) >= 0, nil
}

func (nv *NumericValue) LT(v Value) (bool, error) {
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return false, err
	}
	return x.Cmp(y) < 0, nil
}

func (nv *NumericValue) LTE(v Value) (bool, error) {
	x := nv.Rat
	y, err := v.ToRat()
	if err != nil {
		return false, err
	}
	return x.Cmp(y) <= 0, nil
}

func (nv *NumericValue) ToInt64() (int64, error) {
	return nv.Rat.Num().Int64(), nil
}

func (nv *NumericValue) toString() string {
	var v string
	if nv.IsBigNumeric {
		v = nv.FloatString(38)
	} else {
		v = nv.FloatString(9)
	}
	v = strings.TrimRight(v, "0")
	v = strings.TrimRight(v, ".")
	return v
}

func (nv *NumericValue) ToString() (string, error) {
	return nv.toString(), nil
}

func (nv *NumericValue) ToBytes() ([]byte, error) {
	return []byte(nv.toString()), nil
}

func (nv *NumericValue) ToFloat64() (float64, error) {
	f, _ := nv.Float64()
	return f, nil
}

func (nv *NumericValue) ToBool() (bool, error) {
	v := nv.Rat.Num().Int64()
	switch v {
	case 1:
		return true, nil
	case 0:
		return false, nil
	}
	return false, fmt.Errorf("failed to convert numeric value to bool type")
}

func (nv *NumericValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert array from numeric value")
}

func (nv *NumericValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert struct from numeric value")
}

func (nv *NumericValue) ToJSON() (string, error) {
	return nv.toString(), nil
}

func (nv *NumericValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("failed to convert time.Time from numeric value")
}

func (nv *NumericValue) ToRat() (*big.Rat, error) {
	return nv.Rat, nil
}

func (nv *NumericValue) Format(verb rune) string {
	return nv.toString()
}

func (nv *NumericValue) Interface() any {
	return nv.String()
}
