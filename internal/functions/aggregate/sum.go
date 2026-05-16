package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type SUM struct {
	sum value.Value
}

func (f *SUM) Step(v value.Value, opt *helper.Option) error {
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
	return nil
}

func (f *SUM) Done() (value.Value, error) {
	return f.sum, nil
}
