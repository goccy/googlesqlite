package stat_aggregate

import (
	"gonum.org/v1/gonum/stat"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type VAR_POP struct {
	v []float64
}

func (f *VAR_POP) Step(v value.Value, opt *helper.Option) error {
	if v == nil {
		return nil
	}
	f64, err := v.ToFloat64()
	if err != nil {
		return err
	}
	f.v = append(f.v, f64)
	return nil
}

func (f *VAR_POP) Done() (value.Value, error) {
	if len(f.v) == 0 {
		return nil, nil
	}
	_, variance := stat.PopMeanVariance(f.v, nil)
	return value.FloatValue(variance), nil
}
