package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// BIT_OR_AGG accumulates the bitwise OR of every non-NULL input. val
// stays nil until the first non-NULL Step so that BIT_OR over zero
// input rows reports SQL NULL rather than a synthetic 0 — and so a
// genuine input of -1 is not mistaken for "no value yet".
type BIT_OR_AGG struct {
	val value.Value
}

func (f *BIT_OR_AGG) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	i64, err := v.ToInt64()
	if err != nil {
		return err
	}
	if f.val == nil {
		f.val = value.IntValue(i64)
		return nil
	}
	curI64, err := f.val.ToInt64()
	if err != nil {
		return err
	}
	f.val = value.IntValue(curI64 | i64)
	return nil
}

func (f *BIT_OR_AGG) Done() (value.Value, error) {
	return f.val, nil
}
