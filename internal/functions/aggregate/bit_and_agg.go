package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type BIT_AND_AGG struct {
	val value.Value
}

func (f *BIT_AND_AGG) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	i64, err := v.ToInt64()
	if err != nil {
		return err
	}
	if f.val == nil {
		f.val = value.IntValue(i64)
	} else {
		curI64, err := f.val.ToInt64()
		if err != nil {
			return err
		}
		f.val = value.IntValue(curI64 & i64)
	}
	return nil
}

func (f *BIT_AND_AGG) Done() (value.Value, error) {
	return f.val, nil
}
