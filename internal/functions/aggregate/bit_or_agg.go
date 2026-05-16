package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type BIT_OR_AGG struct {
	val int64
}

func (f *BIT_OR_AGG) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	i64, err := v.ToInt64()
	if err != nil {
		return err
	}
	if f.val == -1 {
		f.val = i64
	} else {
		f.val |= i64
	}
	return nil
}

func (f *BIT_OR_AGG) Done() (value.Value, error) {
	return value.IntValue(f.val), nil
}
