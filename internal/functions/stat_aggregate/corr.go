package stat_aggregate

import (
	"gonum.org/v1/gonum/stat"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type CORR struct {
	x []float64
	y []float64
}

func (f *CORR) Step(x, y value.Value, opt *helper.Option) error {
	if x == nil || y == nil {
		return nil
	}
	vx, err := x.ToFloat64()
	if err != nil {
		return err
	}
	vy, err := y.ToFloat64()
	if err != nil {
		return err
	}
	f.x = append(f.x, vx)
	f.y = append(f.y, vy)
	return nil
}

func (f *CORR) Done() (value.Value, error) {
	if len(f.x) == 0 || len(f.y) == 0 {
		return nil, nil
	}
	return value.FloatValue(stat.Correlation(f.x, f.y, nil)), nil
}
