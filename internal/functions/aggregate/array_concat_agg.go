package aggregate

import (
	"fmt"
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type ARRAY_CONCAT_AGG struct {
	once   sync.Once
	opt    *helper.Option
	values []*helper.OrderedValue
}

func (f *ARRAY_CONCAT_AGG) Step(v *value.ArrayValue, opt *helper.Option) error {
	if v == nil {
		return fmt.Errorf("ARRAY_CONCAT_AGG: NULL value unsupported")
	}
	f.once.Do(func() { f.opt = opt })
	f.values = append(f.values, &helper.OrderedValue{
		OrderBy: opt.OrderBy,
		Value:   v,
	})
	return nil
}

func (f *ARRAY_CONCAT_AGG) Done() (value.Value, error) {
	f.values = helper.SortAggregatedValues(f.values, f.opt)

	if f.opt != nil && f.opt.Limit != nil {
		minLen := min(*f.opt.Limit, int64(len(f.values)))
		f.values = f.values[:minLen]
	}

	var values []value.Value
	for _, v := range f.values {
		a, err := v.Value.ToArray()
		if err != nil {
			return nil, err
		}
		values = append(values, a.Values...)
	}

	return &value.ArrayValue{
		Values: values,
	}, nil
}
