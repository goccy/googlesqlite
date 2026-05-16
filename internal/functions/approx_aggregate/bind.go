package approx_aggregate

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Approximate-aggregate function binders.

func BindApproxCountDistinct() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &APPROX_COUNT_DISTINCT{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return fn.Step(args[0], opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindApproxQuantiles() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &APPROX_QUANTILES{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if args[1] == nil {
					return fmt.Errorf("APPROX_QUANTILES: number must be not null")
				}
				num, err := args[1].ToInt64()
				if err != nil {
					return err
				}
				return fn.Step(args[0], num, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindApproxTopCount() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &APPROX_TOP_COUNT{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if args[1] == nil {
					return fmt.Errorf("APPROX_TOP_COUNT: number must be not null")
				}
				num, err := args[1].ToInt64()
				if err != nil {
					return err
				}
				return fn.Step(args[0], num, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindApproxTopSum() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &APPROX_TOP_SUM{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if args[2] == nil {
					return fmt.Errorf("APPROX_TOP_SUM: number must be not null")
				}
				num, err := args[2].ToInt64()
				if err != nil {
					return err
				}
				return fn.Step(args[0], args[1], num, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}
