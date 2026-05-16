package stat_aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Statistical-aggregate function binders. Each constructor returns
// a fresh *helper.Aggregator that SQLite drives via Step / Done per
// group during query execution.

func BindCorr() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &CORR{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return fn.Step(args[0], args[1], opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindCovarPop() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &COVAR_POP{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return fn.Step(args[0], args[1], opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindCovarSamp() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &COVAR_SAMP{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return fn.Step(args[0], args[1], opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}

func BindStddevPop() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &STDDEV_POP{}
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

func BindStddevSamp() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &STDDEV_SAMP{}
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

func BindStddev() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &STDDEV{}
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

func BindVarPop() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &VAR_POP{}
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

func BindVarSamp() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &VAR_SAMP{}
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

func BindVariance() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &VARIANCE{}
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
