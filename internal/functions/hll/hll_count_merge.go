package hll

import (
	"github.com/DataDog/go-hll"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type HLL_COUNT_MERGE struct {
	hll *hll.Hll
}

func (f *HLL_COUNT_MERGE) Step(sketch []byte, opt *helper.Option) error {
	h, err := hll.FromBytes(sketch)
	if err != nil {
		return err
	}
	if f.hll == nil {
		f.hll = &h
	} else {
		f.hll.Union(h)
	}
	return nil
}

func (f *HLL_COUNT_MERGE) Done() (value.Value, error) {
	if f.hll == nil {
		return value.IntValue(0), nil
	}
	return value.IntValue(f.hll.Cardinality()), nil
}
