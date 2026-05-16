package stat_aggregate

import (
	"gonum.org/v1/gonum/stat"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type COVAR_POP struct {
	x []float64
	y []float64
}

func (f *COVAR_POP) Step(x, y value.Value, opt *helper.Option) error {
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

func (f *COVAR_POP) Done() (value.Value, error) {
	if len(f.x) == 0 || len(f.y) == 0 {
		return nil, nil
	}
	// gonum's stat.Covariance divides by (n-1) — that is the *sample*
	// covariance. COVAR_POP must divide by n. Convert by scaling.
	n := float64(len(f.x))
	if n == 0 {
		return nil, nil
	}
	if n == 1 {
		// Population covariance is well-defined for a single sample
		// (= 0); sample covariance is undefined / NaN.
		return value.FloatValue(0), nil
	}
	sample := stat.Covariance(f.x, f.y, nil)
	return value.FloatValue(sample * (n - 1) / n), nil
}
