package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Aggregate-function binders. Each constructor returns a fresh
// *helper.Aggregator state object that SQLite drives via Step / Done
// per group during query execution.

func BindArray() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &ARRAY{}
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

func BindAnyValue() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &ANY_VALUE{}
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

func BindArrayAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &ARRAY_AGG{}
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

func BindArrayConcatAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &ARRAY_CONCAT_AGG{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if args[0] == nil {
					return nil
				}
				array, err := args[0].ToArray()
				if err != nil {
					return err
				}
				return fn.Step(array, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindAvg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &AVG{}
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

func BindCount() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &COUNT{}
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

func BindCountStar() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &COUNT_STAR{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return fn.Step(opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindBitAndAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &BIT_AND_AGG{value.IntValue(-1)}
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

func BindBitOrAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &BIT_OR_AGG{-1}
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

func BindBitXorAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &BIT_XOR_AGG{1}
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

func BindCountIf() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &COUNTIF{}
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

func BindLogicalAnd() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &LOGICAL_AND{true}
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

func BindLogicalOr() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &LOGICAL_OR{}
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

func BindMax() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &MAX{}
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

func BindMin() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &MIN{}
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

func BindMaxBy() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &MAX_BY{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if len(args) < 2 {
					return nil
				}
				return fn.Step(args[0], args[1], opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindMinBy() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &MIN_BY{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if len(args) < 2 {
					return nil
				}
				return fn.Step(args[0], args[1], opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindStringAgg() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &STRING_AGG{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if len(args) == 1 {
					return fn.Step(args[0], "", opt)
				}
				delim, err := args[1].ToString()
				if err != nil {
					return err
				}
				return fn.Step(args[0], delim, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindSum() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &SUM{}
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
