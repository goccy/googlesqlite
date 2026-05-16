package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type AVG struct {
	sum value.Value
	num int64
}

func (f *AVG) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	if f.sum == nil {
		f.sum = v
	} else {
		added, err := f.sum.Add(v)
		if err != nil {
			return err
		}
		f.sum = added
	}
	f.num++
	return nil
}

func (f *AVG) Done() (value.Value, error) {
	if f.sum == nil {
		return nil, nil
	}
	base, err := f.sum.ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(base / float64(f.num)), nil
}
