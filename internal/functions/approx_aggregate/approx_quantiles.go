package approx_aggregate

import (
	"sort"
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
	// Quantiles are read off the sorted input, NULLs first (they only
	// reach here under RESPECT NULLS).
	sort.SliceStable(f.values, func(i, j int) bool {
		a, b := f.values[i], f.values[j]
		if a == nil || b == nil {
			return a == nil && b != nil
		}
		lt, err := a.LT(b)
		return err == nil && lt
	})
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
