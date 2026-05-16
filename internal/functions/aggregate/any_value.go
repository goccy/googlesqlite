package aggregate

import (
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type ANY_VALUE struct {
	once sync.Once
	opt  *helper.Option
	val  value.Value
}

func (f *ANY_VALUE) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	f.once.Do(func() {
		f.opt = opt
		f.val = v
	})
	return nil
}

func (f *ANY_VALUE) Done() (value.Value, error) {
	return f.val, nil
}
