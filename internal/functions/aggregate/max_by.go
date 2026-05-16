package aggregate

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// MAX_BY(value_expr, key_expr) returns value_expr from the row that
// has the maximum non-null key_expr. NULL keys are skipped. Ties
// take the first-encountered row.
type MAX_BY struct {
	initialized bool
	bestVal     value.Value
	bestKey     value.Value
}

func (f *MAX_BY) Step(v, k value.Value, opt *helper.Option) error {
	if k == nil {
		return nil
	}
	if !f.initialized {
		f.bestVal = v
		f.bestKey = k
		f.initialized = true
		return nil
	}
	cond, err := k.GT(f.bestKey)
	if err != nil {
		return err
	}
	if cond {
		f.bestVal = v
		f.bestKey = k
	}
	return nil
}

func (f *MAX_BY) Done() (value.Value, error) {
	return f.bestVal, nil
}

// MIN_BY mirrors MAX_BY with the opposite ordering.
type MIN_BY struct {
	initialized bool
	bestVal     value.Value
	bestKey     value.Value
}

func (f *MIN_BY) Step(v, k value.Value, opt *helper.Option) error {
	if k == nil {
		return nil
	}
	if !f.initialized {
		f.bestVal = v
		f.bestKey = k
		f.initialized = true
		return nil
	}
	cond, err := k.LT(f.bestKey)
	if err != nil {
		return err
	}
	if cond {
		f.bestVal = v
		f.bestKey = k
	}
	return nil
}

func (f *MIN_BY) Done() (value.Value, error) {
	return f.bestVal, nil
}
