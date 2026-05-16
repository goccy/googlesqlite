package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type LOGICAL_AND struct {
	v bool
}

func (f *LOGICAL_AND) Step(cond value.Value, opt *helper.Option) error {
	if cond == nil {
		return nil
	}
	b, err := cond.ToBool()
	if err != nil {
		return err
	}
	if !b {
		f.v = false
	}
	return nil
}

func (f *LOGICAL_AND) Done() (value.Value, error) {
	return value.BoolValue(f.v), nil
}
