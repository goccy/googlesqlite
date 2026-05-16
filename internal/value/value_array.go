package value

import (
	"fmt"
	"math/big"
	"strings"
	"time"
)

type ArrayValue struct {
	Values []Value
}

func (av *ArrayValue) Has(v Value) (bool, error) {
	for _, val := range av.Values {
		cond, err := val.EQ(v)
		if err != nil {
			return false, err
		}
		if cond {
			return true, nil
		}
	}
	return false, nil
}

func (av *ArrayValue) Add(v Value) (Value, error) {
	return nil, fmt.Errorf("add operation is unsupported for array %v", av)
}

func (av *ArrayValue) Sub(v Value) (Value, error) {
	return nil, fmt.Errorf("sub operation is unsupported for array %v", av)
}

func (av *ArrayValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for array %v", av)
}

func (av *ArrayValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for array %v", av)
}

func (av *ArrayValue) EQ(v Value) (bool, error) {
	arr, err := v.ToArray()
	if err != nil {
		return false, err
	}
	if len(arr.Values) != len(av.Values) {
		return false, nil
	}
	for idx, value := range av.Values {
		cond, err := arr.Values[idx].EQ(value)
		if err != nil {
			return false, err
		}
		if !cond {
			return false, nil
		}
	}
	return true, nil
}

func (av *ArrayValue) GT(v Value) (bool, error) {
	arr, err := v.ToArray()
	if err != nil {
		return false, err
	}
	if len(arr.Values) != len(av.Values) {
		return false, nil
	}
	for idx, value := range av.Values {
		cond, err := arr.Values[idx].GT(value)
		if err != nil {
			return false, err
		}
		if !cond {
			return false, nil
		}
	}
	return true, nil
}

func (av *ArrayValue) GTE(v Value) (bool, error) {
	arr, err := v.ToArray()
	if err != nil {
		return false, err
	}
	if len(arr.Values) != len(av.Values) {
		return false, nil
	}
	for idx, value := range av.Values {
		cond, err := arr.Values[idx].GTE(value)
		if err != nil {
			return false, err
		}
		if !cond {
			return false, nil
		}
	}
	return true, nil
}

func (av *ArrayValue) LT(v Value) (bool, error) {
	arr, err := v.ToArray()
	if err != nil {
		return false, err
	}
	if len(arr.Values) != len(av.Values) {
		return false, nil
	}
	for idx, value := range av.Values {
		cond, err := arr.Values[idx].LT(value)
		if err != nil {
			return false, err
		}
		if !cond {
			return false, nil
		}
	}
	return true, nil
}

func (av *ArrayValue) LTE(v Value) (bool, error) {
	arr, err := v.ToArray()
	if err != nil {
		return false, err
	}
	if len(arr.Values) != len(av.Values) {
		return false, nil
	}
	for idx, value := range av.Values {
		cond, err := arr.Values[idx].LTE(value)
		if err != nil {
			return false, err
		}
		if !cond {
			return false, nil
		}
	}
	return true, nil
}

func (av *ArrayValue) ToInt64() (int64, error) {
	return 0, fmt.Errorf("failed to convert int64 from array %v", av)
}

func (av *ArrayValue) ToString() (string, error) {
	elems := []string{}
	for _, v := range av.Values {
		if v == nil {
			elems = append(elems, "null")
			continue
		}
		elem, err := v.ToJSON()
		if err != nil {
			return "", err
		}
		elems = append(elems, elem)
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ",")), nil
}

func (av *ArrayValue) ToBytes() ([]byte, error) {
	v, err := av.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (av *ArrayValue) ToFloat64() (float64, error) {
	return 0, fmt.Errorf("failed to convert float64 from array %v", av)
}

func (av *ArrayValue) ToBool() (bool, error) {
	return false, fmt.Errorf("failed to convert bool from array %v", av)
}

func (av *ArrayValue) ToArray() (*ArrayValue, error) {
	return av, nil
}

func (av *ArrayValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert struct from array %v", av)
}

func (av *ArrayValue) ToJSON() (string, error) {
	return av.ToString()
}

func (av *ArrayValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("failed to convert time.Time from array %v", av)
}

func (av *ArrayValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("failed to convert *big.Rat from array %v", av)
}

func (av *ArrayValue) Format(verb rune) string {
	elems := []string{}
	for _, v := range av.Values {
		if v == nil {
			elems = append(elems, "NULL")
			continue
		}
		elems = append(elems, v.Format(verb))
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}

func (av *ArrayValue) Interface() any {
	var arr []any
	for _, v := range av.Values {
		if v == nil {
			arr = append(arr, nil)
		} else {
			arr = append(arr, v.Interface())
		}
	}
	return arr
}
