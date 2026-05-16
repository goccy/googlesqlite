package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type COUNTIF struct {
	count value.Value
}

func (f *COUNTIF) Step(cond value.Value, opt *helper.Option) error {
	if cond == nil {
		return nil
	}
	b, err := cond.ToBool()
	if err != nil {
		return err
	}
	if b {
		if f.count == nil {
			f.count = value.IntValue(1)
		} else {
			added, err := f.count.Add(value.IntValue(1))
			if err != nil {
				return err
			}
			f.count = added
		}
	}
	return nil
}

func (f *COUNTIF) Done() (value.Value, error) {
	if f.count == nil {
		return value.IntValue(0), nil
	}
	return f.count, nil
}
