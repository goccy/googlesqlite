package helper

import (
	"fmt"
	"sort"

	"github.com/goccy/googlesqlite/internal/value"
)

// Aggregator is the SQLite-side wrapper that drives a per-spec
// aggregate (Step / Done shape) over a single GROUP. SQLite hands
// each row's args to Step; we strip the encoded option markers
// (DISTINCT / IGNORE_NULLS / LIMIT / ORDER_BY), thread the resulting
// Option through to the per-spec step function, and finally encode
// the Done result back into the over-the-wire form.
//
// Living in the leaf aggrt package alongside Option keeps it
// importable by every aggregate-family sub-package (aggregate,
// approx_aggregate, stat_aggregate, hll) without going through
// internal.
type Aggregator struct {
	distinctMap map[string]struct{}
	distinctNil bool
	step        func([]value.Value, *Option) error
	done        func() (value.Value, error)
}

func (a *Aggregator) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, opt := ParseOptions(values...)
	if opt.IgnoreNulls {
		filtered := []value.Value{}
		for _, v := range values {
			if v == nil {
				continue
			}
			filtered = append(filtered, v)
		}
		values = filtered
		if len(values) == 0 {
			return nil
		}
	}
	if opt.Distinct {
		if len(values) < 1 {
			return fmt.Errorf("DISTINCT option required at least one argument")
		}
		if values[0] == nil {
			if a.distinctNil {
				return nil
			}
			a.distinctNil = true
		} else {
			key, err := values[0].ToString()
			if err != nil {
				return err
			}
			if _, exists := a.distinctMap[key]; exists {
				return nil
			}
			a.distinctMap[key] = struct{}{}
		}
	}
	return a.step(values, opt)
}

func (a *Aggregator) Done() (any, error) {
	ret, err := a.done()
	if err != nil {
		return nil, err
	}
	return value.EncodeValue(ret)
}

// NewAggregator wires up a fresh Aggregator with the given per-spec
// step / done callbacks. Returns the wrapper SQLite drives.
func NewAggregator(
	step func([]value.Value, *Option) error,
	done func() (value.Value, error),
) *Aggregator {
	return &Aggregator{
		distinctMap: map[string]struct{}{},
		step:        step,
		done:        done,
	}
}

// OrderedValue captures a single value plus the ORDER BY decoration
// that arrived with it. Aggregators that observe ORDER BY semantics
// (ARRAY_AGG, STRING_AGG, ARRAY_CONCAT_AGG, ...) buffer rows as
// OrderedValues and call SortAggregatedValues during Done.
type OrderedValue struct {
	OrderBy []*OrderBy
	Value   value.Value
}

// SortAggregatedValues stable-sorts the captured values per the
// ORDER BY directives carried inside each OrderedValue. The caller
// passes the surrounding Option so SortAggregatedValues can short-
// circuit when no ORDER BY was attached.
func SortAggregatedValues(values []*OrderedValue, opt *Option) []*OrderedValue {
	if opt != nil && len(opt.OrderBy) == 0 {
		return values
	}
	sort.Slice(values, func(i, j int) bool {
		for orderBy := 0; orderBy < len(values[0].OrderBy); orderBy++ {
			iV := values[i].OrderBy[orderBy].Value
			jV := values[j].OrderBy[orderBy].Value
			isAsc := values[0].OrderBy[orderBy].IsAsc
			if iV == nil {
				return isAsc
			}
			if jV == nil {
				return !isAsc
			}
			isEqual, _ := iV.EQ(jV)
			if isEqual {
				continue
			}
			if isAsc {
				cond, _ := iV.LT(jV)
				return cond
			}
			cond, _ := iV.GT(jV)
			return cond
		}
		return false
	})
	return values
}
