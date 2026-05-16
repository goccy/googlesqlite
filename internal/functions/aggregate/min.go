package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type MIN struct {
	initialized bool
	min         value.Value
}

func (f *MIN) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	if f.initialized {
		cond, err := v.LT(f.min)
		if err != nil {
			return err
		}
		if cond {
			f.min = v
		}
	} else {
		f.min = v
		f.initialized = true
	}
	return nil
}

func (f *MIN) Done() (value.Value, error) {
	return f.min, nil
}
