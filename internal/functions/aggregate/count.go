package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type COUNT struct {
	count value.Value
}

func (f *COUNT) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	if f.count == nil {
		f.count = value.IntValue(1)
	} else {
		added, err := f.count.Add(value.IntValue(1))
		if err != nil {
			return err
		}
		f.count = added
	}
	return nil
}

func (f *COUNT) Done() (value.Value, error) {
	if f.count == nil {
		return value.IntValue(0), nil
	}
	return f.count, nil
}
