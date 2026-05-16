package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type COUNT_STAR struct {
	count int64
}

func (f *COUNT_STAR) Step(opt *helper.Option) error {
	f.count++
	return nil
}

func (f *COUNT_STAR) Done() (value.Value, error) {
	return value.IntValue(f.count), nil
}
