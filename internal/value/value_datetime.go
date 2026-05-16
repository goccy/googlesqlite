package value

import (
	"fmt"
	"math/big"
	"time"

	"cloud.google.com/go/bigquery"
)

type DatetimeValue time.Time

func (d DatetimeValue) Add(v Value) (Value, error) {
	src := time.Time(d)
	if vv, ok := v.(*IntervalValue); ok {
		return DatetimeValue(time.Date(
			src.Year()+int(vv.Years),
			time.Month(int(src.Month())+int(vv.Months)),
			src.Day()+int(vv.Days),
			src.Hour()+int(vv.Hours),
			src.Minute()+int(vv.Minutes),
			src.Second()+int(vv.Seconds),
			src.Nanosecond()+int(vv.SubSecondNanos),
			src.Location(),
		)), nil
	}
	return nil, fmt.Errorf("failed to use add operator for datetime and %T type", v)
}

func (d DatetimeValue) Sub(v Value) (Value, error) {
	src := time.Time(d)
	if vv, ok := v.(*IntervalValue); ok {
		return DatetimeValue(time.Date(
			src.Year()-int(vv.Years),
			time.Month(int(src.Month())-int(vv.Months)),
			src.Day()-int(vv.Days),
			src.Hour()-int(vv.Hours),
			src.Minute()-int(vv.Minutes),
			src.Second()-int(vv.Seconds),
			src.Nanosecond()-int(vv.SubSecondNanos),
			src.Location(),
		)), nil
	}
	dst, err := v.ToTime()
	if err != nil {
		return nil, err
	}
	duration := src.Sub(dst)
	return &IntervalValue{IntervalValue: bigquery.IntervalValueFromDuration(duration)}, nil
}

func (d DatetimeValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for datetime %v", d)
}

func (d DatetimeValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for datetime %v", d)
}

func (d DatetimeValue) EQ(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Equal(v2), nil
}

func (d DatetimeValue) GT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).After(v2), nil
}

func (d DatetimeValue) GTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Equal(v2) || time.Time(d).After(v2), nil
}

func (d DatetimeValue) LT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Before(v2), nil
}

func (d DatetimeValue) LTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Equal(v2) || time.Time(d).Before(v2), nil
}

func (d DatetimeValue) ToInt64() (int64, error) {
	return time.Time(d).Unix(), nil
}

func (d DatetimeValue) ToString() (string, error) {
	return time.Time(d).Format(datetimeFormat), nil
}

func (d DatetimeValue) ToBytes() ([]byte, error) {
	v, err := d.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (d DatetimeValue) ToFloat64() (float64, error) {
	return float64(time.Time(d).Unix()), nil
}

func (d DatetimeValue) ToBool() (bool, error) {
	return false, fmt.Errorf("failed to convert %v to bool type", d)
}

func (d DatetimeValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert %v to array type", d)
}

func (d DatetimeValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert %v to struct type", d)
}

func (d DatetimeValue) ToJSON() (string, error) {
	return d.ToString()
}

func (d DatetimeValue) ToTime() (time.Time, error) {
	return time.Time(d), nil
}

func (d DatetimeValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("failed to convert *big.Rat from datetime %v", d)
}

func (d DatetimeValue) Format(verb rune) string {
	formatted := time.Time(d).Format(datetimeFormat)
	switch verb {
	case 't':
		return formatted
	case 'T':
		return fmt.Sprintf(`DATETIME %q`, formatted)
	}
	return formatted
}

func (d DatetimeValue) Interface() any {
	return time.Time(d).Format(datetimeFormat)
}
