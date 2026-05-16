package window

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// Window-function (analytic) binders. Each constructor returns
// a fresh *WindowAggregator state object that SQLite drives
// through Step / Inverse / value.Value / Done as the active frame
// scans the partition.

func BindWindowAnyValue() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_ANY_VALUE{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowArrayAgg() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_ARRAY_AGG{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowAvg() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_AVG{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowCount() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_COUNT{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowCountStar() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_COUNT_STAR{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowCountIf() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_COUNTIF{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowLogicalOr() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_LOGICAL_OR{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowLogicalAnd() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_LOGICAL_AND{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowMax() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_MAX{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowMin() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_MIN{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowStringAgg() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_STRING_AGG{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				var delim string
				if len(args) > 1 {
					d, err := args[1].ToString()
					if err != nil {
						return err
					}
					delim = d
				}
				return fn.Step(args[0], delim, windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowSum() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_SUM{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowCorr() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_CORR{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], args[1], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowCovarPop() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_COVAR_POP{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], args[1], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowCovarSamp() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_COVAR_SAMP{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], args[1], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowStddevPop() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_STDDEV_POP{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowStddevSamp() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_STDDEV_SAMP{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowStddev() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_STDDEV{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowVarPop() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_VAR_POP{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowVarSamp() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_VAR_SAMP{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowVariance() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_VARIANCE{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowFirstValue() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_FIRST_VALUE{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowLastValue() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_LAST_VALUE{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowNthValue() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_NTH_VALUE{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				if args[1] == nil {
					return fmt.Errorf("NTH_VALUE: constant integer expression must be not null value")
				}
				num, err := args[1].ToInt64()
				if err != nil {
					return err
				}
				return fn.Step(args[0], num, windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowLead() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_LEAD{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				var offset int64 = 1
				if len(args) >= 2 {
					if args[1] == nil {
						return fmt.Errorf("LEAD: offset is must be not null value")
					}
					v, err := args[1].ToInt64()
					if err != nil {
						return err
					}
					offset = v
				}
				if offset < 0 {
					return fmt.Errorf("LEAD: offset is must be positive value %d", offset)
				}
				var defaultValue value.Value
				if len(args) == 3 {
					defaultValue = args[2]
				}
				return fn.Step(args[0], offset, defaultValue, windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowLag() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_LAG{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				var offset int64 = 1
				if len(args) >= 2 {
					if args[1] == nil {
						return fmt.Errorf("LAG: offset is must be not null value")
					}
					v, err := args[1].ToInt64()
					if err != nil {
						return err
					}
					offset = v
				}
				if offset < 0 {
					return fmt.Errorf("LAG: offset is must be positive value %d", offset)
				}
				var defaultValue value.Value
				if len(args) == 3 {
					defaultValue = args[2]
				}
				return fn.Step(args[0], offset, defaultValue, windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowPercentileCont() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_PERCENTILE_CONT{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], args[1], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowPercentileDisc() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_PERCENTILE_DISC{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(args[0], args[1], windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowRank() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_RANK{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowDenseRank() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_DENSE_RANK{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowPercentRank() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_PERCENT_RANK{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowCumeDist() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_CUME_DIST{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowNtile() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_NTILE{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				if args[0] == nil {
					return fmt.Errorf("NTILE: constant integer expression must be not null value")
				}
				num, err := args[0].ToInt64()
				if err != nil {
					return err
				}
				if num <= 0 {
					return fmt.Errorf("NTILE: constant integer expression must be positive value")
				}
				return fn.Step(num, windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}

func BindWindowRowNumber() func() *WindowAggregator {
	return func() *WindowAggregator {
		fn := &WINDOW_ROW_NUMBER{}
		return newWindowAggregator(
			func(args []value.Value, windowOpt *WindowFuncStatus, agg *WindowFuncAggregatedStatus) error {
				return fn.Step(windowOpt, agg)
			},
			func(agg *WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}
