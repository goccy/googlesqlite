package approx_aggregate

import (
	"fmt"
	"sort"
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type APPROX_TOP_SUM struct {
	once     sync.Once
	valueMap map[value.Value]*value.StructValue
	num      int64
}

func (f *APPROX_TOP_SUM) Step(v, weight value.Value, num int64, opt *helper.Option) error {
	f.once.Do(func() {
		f.valueMap = map[value.Value]*value.StructValue{}
		f.num = num
	})
	val, exists := f.valueMap[v]
	if exists {
		if weight != nil {
			var sum value.Value
			if val.Values[1] == nil {
				sum = weight
			} else {
				added, err := val.Values[1].Add(weight)
				if err != nil {
					return err
				}
				sum = added
			}
			val.Values[1] = sum
			val.M["sum"] = sum
		}
	} else {
		f.valueMap[v] = &value.StructValue{
			Keys:   []string{"value", "sum"},
			Values: []value.Value{v, weight},
			M: map[string]value.Value{
				"value": v,
				"sum":   weight,
			},
		}
	}
	return nil
}

func (f *APPROX_TOP_SUM) Done() (value.Value, error) {
	if len(f.valueMap) == 0 {
		return nil, nil
	}
	if int64(len(f.valueMap)) < f.num {
		return nil, fmt.Errorf("APPROX_TOP_SUM: required number is larger than number of input values")
	}
	values := make([]*value.StructValue, 0, len(f.valueMap))
	for _, v := range f.valueMap {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].Values[1] == nil {
			return false
		}
		if values[j].Values[1] == nil {
			return true
		}
		cond, _ := values[i].Values[1].GT(values[j].Values[1])
		return cond
	})
	ret := &value.ArrayValue{}
	for _, v := range values[:f.num] {
		ret.Values = append(ret.Values, v)
	}
	return ret, nil
}
