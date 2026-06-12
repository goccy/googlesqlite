package value

import (
	"fmt"
	"math/big"
	"time"

	"github.com/goccy/googlesqlite/internal/intervalvalue"
)

type TimestampValue time.Time

func (t TimestampValue) AddValueWithPart(v int64, part string) (Value, error) {
	switch part {
	case "MICROSECOND":
		return TimestampValue(time.Time(t).Add(time.Duration(v) * time.Microsecond)), nil
	case "MILLISECOND":
		return TimestampValue(time.Time(t).Add(time.Duration(v) * time.Millisecond)), nil
	case "SECOND":
		return TimestampValue(time.Time(t).Add(time.Duration(v) * time.Second)), nil
	case "MINUTE":
		return TimestampValue(time.Time(t).Add(time.Duration(v) * time.Minute)), nil
	case "HOUR":
		return TimestampValue(time.Time(t).Add(time.Duration(v) * time.Hour)), nil
	case "DAY":
		return TimestampValue(time.Time(t).Add(time.Duration(v) * time.Hour * 24)), nil
	default:
		return nil, fmt.Errorf("unknown part value for timestamp: %s", part)
	}
}

func (t TimestampValue) Add(v Value) (Value, error) {
	src := time.Time(t)
	if vv, ok := v.(*IntervalValue); ok {
		return TimestampValue(time.Date(
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
	return nil, fmt.Errorf("failed to use add operator for timestamp and %T type", v)
}

func (t TimestampValue) Sub(v Value) (Value, error) {
	src := time.Time(t)
	if vv, ok := v.(*IntervalValue); ok {
		return TimestampValue(time.Date(
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
	return &IntervalValue{IntervalValue: intervalvalue.IntervalValueFromDuration(duration)}, nil
}

func (t TimestampValue) Mul(v Value) (Value, error) {
	return nil, fmt.Errorf("mul operation is unsupported for timestamp %v", t)
}

func (t TimestampValue) Div(v Value) (Value, error) {
	return nil, fmt.Errorf("div operation is unsupported for timestamp %v", t)
}

func (t TimestampValue) EQ(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Equal(v2), nil
}

func (t TimestampValue) GT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).After(v2), nil
}

func (t TimestampValue) GTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Equal(v2) || time.Time(t).After(v2), nil
}

func (t TimestampValue) LT(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Before(v2), nil
}

func (t TimestampValue) LTE(v Value) (bool, error) {
	v2, err := v.ToTime()
	if err != nil {
		return false, fmt.Errorf("failed to convert %v to time.Time", v)
	}
	return time.Time(t).Equal(v2) || time.Time(t).Before(v2), nil
}

func (t TimestampValue) ToInt64() (int64, error) {
	return time.Time(t).Unix(), nil
}

func (t TimestampValue) ToString() (string, error) {
	return time.Time(t).Format(time.RFC3339Nano), nil
}

func (t TimestampValue) ToBytes() ([]byte, error) {
	v, err := t.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (t TimestampValue) ToFloat64() (float64, error) {
	return float64(time.Time(t).Unix()), nil
}

func (t TimestampValue) ToBool() (bool, error) {
	return false, fmt.Errorf("failed to convert %v to bool type", t)
}

func (t TimestampValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("failed to convert %v to array type", t)
}

func (t TimestampValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("failed to convert %v to struct type", t)
}

func (t TimestampValue) ToJSON() (string, error) {
	return t.ToString()
}

func (t TimestampValue) ToTime() (time.Time, error) {
	return time.Time(t), nil
}

func (t TimestampValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("failed to convert *big.Rat from timestamp %v", t)
}

func (t TimestampValue) Format(verb rune) string {
	const timestampPrintableFormat = "2006-01-02 15:04:05"
	formatted := time.Time(t).UTC().Format(timestampPrintableFormat) + "+00"
	switch verb {
	case 't':
		return formatted
	case 'T':
		return fmt.Sprintf(`TIMESTAMP %q`, formatted)
	}
	return formatted
}

func (t TimestampValue) Interface() any {
	return time.Time(t).Format(time.RFC3339)
}
