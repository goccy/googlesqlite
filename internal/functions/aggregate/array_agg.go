package aggregate

import (
	"fmt"
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type ARRAY_AGG struct {
	once   sync.Once
	opt    *helper.Option
	values []*helper.OrderedValue
}

func (f *ARRAY_AGG) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return fmt.Errorf("ARRAY_AGG: input value must be not null")
	}
	f.once.Do(func() { f.opt = opt })
	f.values = append(f.values, &helper.OrderedValue{
		OrderBy: opt.OrderBy,
		Value:   v,
	})
	return nil
}

func (f *ARRAY_AGG) Done() (value.Value, error) {
	f.values = helper.SortAggregatedValues(f.values, f.opt)
	if f.opt != nil && f.opt.Limit != nil {
		minLen := min(*f.opt.Limit, int64(len(f.values)))
		f.values = f.values[:minLen]
	}
	values := make([]value.Value, 0, len(f.values))
	for _, v := range f.values {
		values = append(values, v.Value)
	}
	return &value.ArrayValue{
		Values: values,
	}, nil
}
