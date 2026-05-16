package hll

import (
	"github.com/DataDog/go-hll"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type HLL_COUNT_MERGE_PARTIAL struct {
	hll *hll.Hll
}

func (f *HLL_COUNT_MERGE_PARTIAL) Step(sketch []byte, opt *helper.Option) error {
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

func (f *HLL_COUNT_MERGE_PARTIAL) Done() (value.Value, error) {
	if f.hll == nil {
		return nil, nil
	}
	return value.BytesValue(f.hll.ToBytes()), nil
}
