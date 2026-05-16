package hll

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// HyperLogLog++ function binders.
// HLL_COUNT.INIT / .MERGE / .MERGE_PARTIAL run as aggregates;
// HLL_COUNT.EXTRACT is a scalar (BindHllCountExtract).

func BindHllCountInit() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &HLL_COUNT_INIT{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				precision := int64(15)
				if len(args) == 2 {
					if args[1] == nil {
						return nil
					}
					v, err := args[1].ToInt64()
					if err != nil {
						return err
					}
					precision = v
				}
				return fn.Step(args[0], precision, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindHllCountMerge() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &HLL_COUNT_MERGE{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if args[0] == nil {
					return nil
				}
				b, err := args[0].ToBytes()
				if err != nil {
					return err
				}
				return fn.Step(b, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindHllCountMergePartial() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &HLL_COUNT_MERGE_PARTIAL{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if args[0] == nil {
					return nil
				}
				b, err := args[0].ToBytes()
				if err != nil {
					return err
				}
				return fn.Step(b, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindHllCountExtract(args ...value.Value) (value.Value, error) {
	if args[0] == nil {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	return HLL_COUNT_EXTRACT(b)
}
