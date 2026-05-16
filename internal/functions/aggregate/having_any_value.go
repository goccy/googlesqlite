package aggregate

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

type HAVING_ANY_VALUE struct {
	values []value.Value
	keys   []value.Value
	want   string
}

func (f *HAVING_ANY_VALUE) Step(val, having value.Value, modifier string, _ *helper.Option) error {
	if f.want == "" {
		f.want = modifier
	}
	f.values = append(f.values, val)
	f.keys = append(f.keys, having)
	return nil
}

func (f *HAVING_ANY_VALUE) Done() (value.Value, error) {
	var bestKey value.Value
	wantMin := f.want == "MIN"
	for _, k := range f.keys {
		if k == nil {
			continue
		}
		if bestKey == nil {
			bestKey = k
			continue
		}
		var cond bool
		var err error
		if wantMin {
			cond, err = k.LT(bestKey)
		} else {
			cond, err = k.GT(bestKey)
		}
		if err != nil {
			return nil, err
		}
		if cond {
			bestKey = k
		}
	}
	if bestKey == nil {
		return nil, nil
	}
	for i, k := range f.keys {
		if k == nil {
			continue
		}
		eq, err := k.EQ(bestKey)
		if err != nil {
			return nil, err
		}
		if !eq {
			continue
		}
		if f.values[i] == nil {
			continue
		}
		return f.values[i], nil
	}
	return nil, nil
}

func BindHavingAnyValue() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		fn := &HAVING_ANY_VALUE{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if len(args) != 3 {
					return fmt.Errorf("googlesqlite_having_any_value: expected 3 args, got %d", len(args))
				}
				modifier := ""
				if args[2] != nil {
					s, err := args[2].ToString()
					if err != nil {
						return err
					}
					modifier = s
				}
				return fn.Step(args[0], args[1], modifier, opt)
			},
			func() (value.Value, error) {
				return fn.Done()
			},
		)
	}
}
