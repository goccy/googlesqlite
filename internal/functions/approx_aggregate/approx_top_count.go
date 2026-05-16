package approx_aggregate

import (
	"fmt"
	"sort"
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type APPROX_TOP_COUNT struct {
	once     sync.Once
	valueMap map[value.Value]*value.StructValue
	num      int64
}

func (f *APPROX_TOP_COUNT) Step(v value.Value, num int64, opt *helper.Option) error {
	f.once.Do(func() {
		f.valueMap = map[value.Value]*value.StructValue{}
		f.num = num
	})
	val, exists := f.valueMap[v]
	if exists {
		cur, _ := val.Values[1].ToInt64()
		val.Values[1] = value.IntValue(cur + 1)
		val.M["count"] = value.IntValue(cur + 1)
	} else {
		f.valueMap[v] = &value.StructValue{
			Keys:   []string{"value", "count"},
			Values: []value.Value{v, value.IntValue(1)},
			M: map[string]value.Value{
				"value": v,
				"count": value.IntValue(1),
			},
		}
	}
	return nil
}

func (f *APPROX_TOP_COUNT) Done() (value.Value, error) {
	if len(f.valueMap) == 0 {
		return nil, nil
	}
	if int64(len(f.valueMap)) < f.num {
		return nil, fmt.Errorf("APPROX_TOP_COUNT: required number is larger than number of input values")
	}
	values := make([]*value.StructValue, 0, len(f.valueMap))
	for _, v := range f.valueMap {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		cond, _ := values[i].Values[1].GT(values[j].Values[1])
		return cond
	})
	ret := &value.ArrayValue{}
	for _, v := range values[:f.num] {
		ret.Values = append(ret.Values, v)
	}
	return ret, nil
}
