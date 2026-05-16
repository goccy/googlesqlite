package hll

import (
	"sync"

	"github.com/DataDog/go-hll"
	"github.com/spaolacci/murmur3"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type HLL_COUNT_INIT struct {
	once sync.Once
	hll  *hll.Hll
}

func (f *HLL_COUNT_INIT) Step(input value.Value, precision int64, opt *helper.Option) (e error) {
	f.once.Do(func() {
		h, err := hll.NewHll(hll.Settings{Log2m: int(precision)})
		if err != nil {
			e = err
		}
		f.hll = &h
	})
	var v uint64
	switch input.(type) {
	case value.IntValue:
		s, err := input.ToString()
		if err != nil {
			return err
		}
		v = murmur3.Sum64([]byte(s))
	case *value.NumericValue:
		b, err := input.ToBytes()
		if err != nil {
			return err
		}
		v = murmur3.Sum64(b)
	case value.StringValue:
		s, err := input.ToString()
		if err != nil {
			return err
		}
		v = murmur3.Sum64([]byte(s))
	case value.BytesValue:
		b, err := input.ToBytes()
		if err != nil {
			return err
		}
		v = murmur3.Sum64(b)
	}
	f.hll.AddRaw(v)
	return nil
}

func (f *HLL_COUNT_INIT) Done() (value.Value, error) {
	if f.hll == nil {
		return nil, nil
	}
	return value.BytesValue(f.hll.ToBytes()), nil
}
