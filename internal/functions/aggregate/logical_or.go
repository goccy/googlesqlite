package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// LOGICAL_OR reports whether any non-NULL input is TRUE. v stays nil
// until the first non-NULL Step so that LOGICAL_OR over zero input
// rows reports SQL NULL rather than a synthetic FALSE.
type LOGICAL_OR struct {
	v value.Value
}

func (f *LOGICAL_OR) Step(cond value.Value, opt *helper.Option) error {
	if cond == nil {
		return nil
	}
	b, err := cond.ToBool()
	if err != nil {
		return err
	}
	if f.v == nil {
		f.v = value.BoolValue(b)
		return nil
	}
	cur, err := f.v.ToBool()
	if err != nil {
		return err
	}
	f.v = value.BoolValue(cur || b)
	return nil
}

func (f *LOGICAL_OR) Done() (value.Value, error) {
	return f.v, nil
}
