package value

import (
	"fmt"
	"math/big"
	"time"
)

// RangeValue is a half-open interval [Start, End) over a comparable
// element type. Per BigQuery, RANGE supports DATE, DATETIME, and
// TIMESTAMP element types. Either bound can be NULL — interpreted as
// "unbounded".
type RangeValue struct {
	Start Value // nil = unbounded
	End   Value // nil = unbounded
	// ElemHeader records the element ValueType so the decoder can
	// reconstruct typed bounds without redundant type inference.
	ElemHeader ValueType
}

func (r *RangeValue) Add(Value) (Value, error) {
	return nil, fmt.Errorf("unsupported add operator for range value")
}

func (r *RangeValue) Sub(Value) (Value, error) {
	return nil, fmt.Errorf("unsupported sub operator for range value")
}

func (r *RangeValue) Mul(Value) (Value, error) {
	return nil, fmt.Errorf("unsupported mul operator for range value")
}

func (r *RangeValue) Div(Value) (Value, error) {
	return nil, fmt.Errorf("unsupported div operator for range value")
}

func (r *RangeValue) EQ(other Value) (bool, error) {
	rr, ok := other.(*RangeValue)
	if !ok {
		return false, fmt.Errorf("RANGE EQ: other side is %T", other)
	}
	if (r.Start == nil) != (rr.Start == nil) {
		return false, nil
	}
	if (r.End == nil) != (rr.End == nil) {
		return false, nil
	}
	if r.Start != nil {
		eq, err := r.Start.EQ(rr.Start)
		if err != nil || !eq {
			return false, err
		}
	}
	if r.End != nil {
		eq, err := r.End.EQ(rr.End)
		if err != nil || !eq {
			return false, err
		}
	}
	return true, nil
}

func (r *RangeValue) GT(Value) (bool, error) {
	return false, fmt.Errorf("unsupported gt operator for range value")
}

func (r *RangeValue) GTE(Value) (bool, error) {
	return false, fmt.Errorf("unsupported gte operator for range value")
}

func (r *RangeValue) LT(Value) (bool, error) {
	return false, fmt.Errorf("unsupported lt operator for range value")
}

func (r *RangeValue) LTE(Value) (bool, error) {
	return false, fmt.Errorf("unsupported lte operator for range value")
}

func (r *RangeValue) ToInt64() (int64, error) {
	return 0, fmt.Errorf("unsupported int64 cast for range value")
}

func (r *RangeValue) ToString() (string, error) {
	startStr := "UNBOUNDED"
	if r.Start != nil {
		s, err := rangeBoundDisplay(r.Start)
		if err != nil {
			return "", err
		}
		startStr = s
	}
	endStr := "UNBOUNDED"
	if r.End != nil {
		s, err := rangeBoundDisplay(r.End)
		if err != nil {
			return "", err
		}
		endStr = s
	}
	return fmt.Sprintf("[%s, %s)", startStr, endStr), nil
}

// rangeBoundDisplay renders one bound of a RANGE in the form BigQuery
// shows it inside the half-open `[start, end)` literal:
//
//   - DATE  → `2006-01-02`               (Value.ToString is already this)
//   - DATETIME → `2006-01-02 15:04:05.999999`
//     (space separator, not `T`, to match BigQuery's RANGE display)
//   - TIMESTAMP → `2006-01-02 15:04:05.000000+00` in UTC
//
// Anything else falls through to Value.ToString. Keeping the override
// local to RangeValue avoids changing the global Datetime / Timestamp
// ToString contract — only the RANGE container renders differently.
func rangeBoundDisplay(v Value) (string, error) {
	switch x := v.(type) {
	case DatetimeValue:
		return time.Time(x).Format("2006-01-02 15:04:05.999999"), nil
	case TimestampValue:
		return time.Time(x).UTC().Format("2006-01-02 15:04:05.000000") + "+00", nil
	}
	return v.ToString()
}

func (r *RangeValue) ToBytes() ([]byte, error) {
	s, err := r.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (r *RangeValue) ToFloat64() (float64, error) {
	return 0, fmt.Errorf("unsupported float64 cast for range value")
}

func (r *RangeValue) ToBool() (bool, error) {
	return false, fmt.Errorf("unsupported bool cast for range value")
}

func (r *RangeValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("unsupported array cast for range value")
}

func (r *RangeValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("unsupported struct cast for range value")
}

func (r *RangeValue) ToJSON() (string, error) {
	s, err := r.ToString()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%q", s), nil
}

func (r *RangeValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("unsupported time cast for range value")
}

func (r *RangeValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("unsupported numeric cast for range value")
}

func (r *RangeValue) Format(verb rune) string {
	s, err := r.ToString()
	if err != nil {
		return "error"
	}
	switch verb {
	case 'T':
		return fmt.Sprintf("RANGE %q", s)
	}
	return s
}

func (r *RangeValue) Interface() any {
	s, err := r.ToString()
	if err != nil {
		return nil
	}
	return s
}
