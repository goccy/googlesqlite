package aggregate

import (
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type ARRAY struct {
	once   sync.Once
	opt    *helper.Option
	values []*helper.OrderedValue
}

func (f *ARRAY) Step(v value.Value, opt *helper.Option) error {
	f.once.Do(func() { f.opt = opt })
	f.values = append(f.values, &helper.OrderedValue{
		Value: v,
	})
	return nil
}

func (f *ARRAY) Done() (value.Value, error) {
	values := make([]value.Value, 0, len(f.values))
	for _, v := range f.values {
		values = append(values, v.Value)
	}
	return &value.ArrayValue{
		Values: values,
	}, nil
}
