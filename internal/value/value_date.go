package value

import (
	"fmt"
	"math/big"
	"time"

	"github.com/goccy/googlesqlite/internal/intervalvalue"
)

type DateValue time.Time

func (d DateValue) AddDateWithInterval(v int, interval string) (Value, error) {
	switch interval {
	case "WEEK":
		return DateValue(time.Time(d).AddDate(0, 0, v*7)), nil
	case "MONTH":
		return DateValue(time.Time(d).AddDate(0, v, 0)), nil
	case "YEAR":
		return DateValue(time.Time(d).AddDate(v, 0, 0)), nil
	default:
		return DateValue(time.Time(d).AddDate(0, 0, v)), nil
	}
}

func (d DateValue) Add(v Value) (Value, error) {
	src := time.Time(d)
	switch vv := v.(type) {
	case *IntervalValue:
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
	case IntValue:
		return DateValue(time.Time(d).AddDate(0, 0, int(vv))), nil
	}
	return nil, fmt.Errorf("failed to use add operator for date and %T type", v)
}

func (d DateValue) Sub(v Value) (Value, error) {
	src := time.Time(d)
	switch vv := v.(type) {
	case *IntervalValue:
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
	case IntValue:
		return DateValue(time.Time(d).AddDate(0, 0, -int(vv))), nil
	}
	dst, err := v.ToTime()
	if err != nil {
		return nil, err
	}
	duration := time.Time(d).Sub(dst)
	days := duration / (24 * time.Hour)
	return &IntervalValue{
		IntervalValue: &intervalvalue.IntervalValue{
			Days: int32(days),
		},
	}, nil
}

func (d DateValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for date %v", d)
}

func (d DateValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for date %v", d)
}

func (d DateValue) EQ(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Equal(v2), nil
}

func (d DateValue) GT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).After(v2), nil
}

func (d DateValue) GTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Equal(v2) || time.Time(d).After(v2), nil
}

func (d DateValue) LT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Before(v2), nil
}

func (d DateValue) LTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(d).Equal(v2) || time.Time(d).Before(v2), nil
}

func (d DateValue) ToInt64() (int64, error) {
	return time.Time(d).Unix(), nil
}

func (d DateValue) ToString() (string, error) {
	return time.Time(d).Format("2006-01-02"), nil
}

func (d DateValue) ToBytes() ([]byte, error) {
	v, err := d.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (d DateValue) ToFloat64() (float64, error) {
	return float64(time.Time(d).Unix()), nil
}

func (d DateValue) ToBool() (bool, error) {
	return false, fmt.Errorf("failed to convert %v to bool type", d)
}

func (d DateValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert %v to array type", d)
}

func (d DateValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert %v to struct type", d)
}

func (d DateValue) ToJSON() (string, error) {
	return d.ToString()
}

func (d DateValue) ToTime() (time.Time, error) {
	return time.Time(d), nil
}

func (d DateValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("failed to convert *big.Rat from date %v", d)
}

func (d DateValue) Format(verb rune) string {
	formatted := time.Time(d).Format("2006-01-02")
	switch verb {
	case 't':
		return formatted
	case 'T':
		return fmt.Sprintf(`DATE %q`, formatted)
	}
	return formatted
}

func (d DateValue) Interface() any {
	return time.Time(d).Format("2006-01-02")
}

const (
	datetimeFormat = "2006-01-02T15:04:05.999999"
)
