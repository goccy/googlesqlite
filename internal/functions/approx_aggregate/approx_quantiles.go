package approx_aggregate

import (
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type APPROX_QUANTILES struct {
	once   sync.Once
	values []value.Value
	num    int64
}

func (f *APPROX_QUANTILES) Step(v value.Value, num int64, opt *helper.Option) error {
	f.once.Do(func() {
		f.num = num
	})
	f.values = append(f.values, v)
	return nil
}

func (f *APPROX_QUANTILES) Done() (value.Value, error) {
	if len(f.values) == 0 {
		return nil, nil
	}
	if f.num == 0 {
		return &value.ArrayValue{Values: []value.Value{f.values[0]}}, nil
	}
	if f.num == 1 {
		return &value.ArrayValue{Values: []value.Value{f.values[0], f.values[len(f.values)-1]}}, nil
	}
	ratio := float64(100) / float64(f.num)
	length := float64(len(f.values))
	quantiles := []value.Value{}
	for i := float64(0); i < 100; i += ratio {
		fIdx := length * (i / 100)
		idx := int64(fIdx)
		if float64(idx) < fIdx {
			idx += 1
		}
		if idx > 0 {
			quantiles = append(quantiles, f.values[idx-1])
		} else {
			quantiles = append(quantiles, f.values[idx])
		}
	}
	quantiles = append(quantiles, f.values[len(f.values)-1])
	return &value.ArrayValue{Values: quantiles}, nil
}
