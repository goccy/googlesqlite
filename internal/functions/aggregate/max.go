package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type MAX struct {
	initialized bool
	max         value.Value
}

func (f *MAX) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	if f.initialized {
		cond, err := v.GT(f.max)
		if err != nil {
			return err
		}
		if cond {
			f.max = v
		}
	} else {
		f.max = v
		f.initialized = true
	}
	return nil
}

func (f *MAX) Done() (value.Value, error) {
	return f.max, nil
}
