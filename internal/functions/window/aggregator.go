package window

import (
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// WindowAggregator is the SQLite-side wrapper that drives a per-spec
// WINDOW_X implementation across the active frame. SQLite calls
// Step / Done; the aggregator parses option markers out of the
// argument list, threads the resulting *WindowFuncStatus through to
// the per-spec step function, and finally encodes the Done result
// back into the over-the-wire form.
type WindowAggregator struct {
	distinctMap map[string]struct{}
	agg         *WindowFuncAggregatedStatus
	step        func([]value.Value, *WindowFuncStatus, *WindowFuncAggregatedStatus) error
	done        func(*WindowFuncAggregatedStatus) (value.Value, error)
	once        sync.Once
}

func (a *WindowAggregator) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, opt := helper.ParseOptions(values...)
	values, windowOpt := parseWindowOptions(values...)
	a.once.Do(func() {
		a.agg.IgnoreNullsOpt = opt.IgnoreNulls
		a.agg.DistinctOpt = opt.Distinct
	})
	return a.step(values, windowOpt, a.agg)
}

func (a *WindowAggregator) Done() (any, error) {
	ret, err := a.done(a.agg)
	if err != nil {
		return nil, err
	}
	return value.EncodeValue(ret)
}

func newWindowAggregator(
	step func([]value.Value, *WindowFuncStatus, *WindowFuncAggregatedStatus) error,
	done func(*WindowFuncAggregatedStatus) (value.Value, error),
) *WindowAggregator {
	return &WindowAggregator{
		distinctMap: map[string]struct{}{},
		agg:         newWindowFuncAggregatedStatus(),
		step:        step,
		done:        done,
	}
}

// NewWindowAggregator is the exported counterpart of
// newWindowAggregator, used by per-category window implementations
// outside this package (e.g. text_analysis/TF_IDF).
func NewWindowAggregator(
	step func([]value.Value, *WindowFuncStatus, *WindowFuncAggregatedStatus) error,
	done func(*WindowFuncAggregatedStatus) (value.Value, error),
) *WindowAggregator {
	return newWindowAggregator(step, done)
}
