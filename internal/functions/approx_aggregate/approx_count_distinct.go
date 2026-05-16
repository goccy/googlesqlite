package approx_aggregate

import (
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type APPROX_COUNT_DISTINCT struct {
	once     sync.Once
	valueMap map[string]struct{}
}

func (f *APPROX_COUNT_DISTINCT) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	f.once.Do(func() { f.valueMap = map[string]struct{}{} })
	val, err := v.ToString()
	if err != nil {
		return err
	}
	f.valueMap[val] = struct{}{}
	return nil
}

func (f *APPROX_COUNT_DISTINCT) Done() (value.Value, error) {
	return value.IntValue(len(f.valueMap)), nil
}
