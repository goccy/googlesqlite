package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// LOGICAL_AND reports whether every non-NULL input is TRUE. v stays nil
// until the first non-NULL Step so that LOGICAL_AND over zero input
// rows reports SQL NULL rather than a synthetic TRUE.
type LOGICAL_AND struct {
	v value.Value
}

func (f *LOGICAL_AND) Step(cond value.Value, opt *helper.Option) error {
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
	f.v = value.BoolValue(cur && b)
	return nil
}

func (f *LOGICAL_AND) Done() (value.Value, error) {
	return f.v, nil
}
