// Package aggrt holds the aggregator-runtime types and helpers shared
// between the internal Aggregator wrapper, the window sub-package, and
// per-aggregate function implementations. It is a leaf package: it
// depends on internal/value but is depended upon by internal/ and by
// internal/function/window/.
package helper

import (
	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"
)

// FuncOption is one of the encoded option markers
// (DISTINCT / IGNORE_NULLS / LIMIT / ORDER_BY) that the analyzer wraps
// around aggregate / window argument lists. parseAggregateOptions
// strips them from the runtime arg list and merges them into an
// AggregatorOption.
type FuncOption struct {
	Type  FuncOptionType `json:"type"`
	Value any            `json:"value"`
}

func (o *FuncOption) UnmarshalJSON(b []byte) error {
	type funcOption FuncOption

	var v funcOption
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	o.Type = v.Type
	switch v.Type {
	case OptionDistinct:
	case OptionIgnoreNulls:
	case OptionLimit:
		var val struct {
			Value int64 `json:"value"`
		}
		if err := json.Unmarshal(b, &val); err != nil {
			return err
		}
		o.Value = val.Value
	case OptionOrderBy:
		var val struct {
			Value *OrderBy `json:"value"`
		}
		if err := json.Unmarshal(b, &val); err != nil {
			return err
		}
		o.Value = val.Value
	}
	return nil
}

type FuncOptionType string

const (
	OptionUnknown     FuncOptionType = "aggregate_unknown"
	OptionDistinct    FuncOptionType = "aggregate_distinct"
	OptionLimit       FuncOptionType = "aggregate_limit"
	OptionOrderBy     FuncOptionType = "aggregate_order_by"
	OptionIgnoreNulls FuncOptionType = "aggregate_ignore_nulls"
)

// DISTINCT, IGNORE_NULLS, LIMIT, ORDER_BY emit the encoded marker
// values that the analyzer-rewritten SQL passes alongside the real
// aggregate arguments.
func DISTINCT() (value.Value, error) {
	b, _ := json.Marshal(&FuncOption{
		Type: OptionDistinct,
	})
	return value.StringValue(string(b)), nil
}

func LIMIT(limit int64) (value.Value, error) {
	b, _ := json.Marshal(&FuncOption{
		Type:  OptionLimit,
		Value: limit,
	})
	return value.StringValue(string(b)), nil
}

func IGNORE_NULLS() (value.Value, error) {
	b, _ := json.Marshal(&FuncOption{
		Type: OptionIgnoreNulls,
	})
	return value.StringValue(string(b)), nil
}

// OrderBy is the per-arg ordering directive packed into an
// ORDER_BY(<value>, <isAsc>) marker. The Value field is decoded from
// its over-the-wire form by UnmarshalJSON below using the canonical
// value codec.
type OrderBy struct {
	Value value.Value `json:"value"`
	IsAsc bool        `json:"isAsc"`
}

func (a *OrderBy) UnmarshalJSON(b []byte) error {
	var v struct {
		Value any  `json:"value"`
		IsAsc bool `json:"isAsc"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	val, err := value.ValueFromGoValue(v.Value)
	if err != nil {
		return err
	}
	a.Value = val
	a.IsAsc = v.IsAsc
	return nil
}

func ORDER_BY(v value.Value, isAsc bool) (value.Value, error) {
	b, _ := json.Marshal(&FuncOption{
		Type: OptionOrderBy,
		Value: &OrderBy{
			Value: v,
			IsAsc: isAsc,
		},
	})
	return value.StringValue(string(b)), nil
}

// Option holds the aggregate-call-level flags parsed out of the
// FuncOption markers in a step's argument list.
type Option struct {
	Distinct    bool
	IgnoreNulls bool
	Limit       *int64
	OrderBy     []*OrderBy
}

// ParseOptions strips the encoded option markers from args and returns
// the cleaned arg list together with the resulting Option.
func ParseOptions(args ...value.Value) ([]value.Value, *Option) {
	var (
		filteredArgs []value.Value
		opt          = &Option{}
	)
	for _, arg := range args {
		if arg == nil {
			filteredArgs = append(filteredArgs, nil)
			continue
		}
		text, err := arg.ToString()
		if err != nil {
			filteredArgs = append(filteredArgs, arg)
			continue
		}
		var v FuncOption
		if err := json.Unmarshal([]byte(text), &v); err != nil {
			filteredArgs = append(filteredArgs, arg)
			continue
		}
		switch v.Type {
		case OptionDistinct:
			opt.Distinct = true
		case OptionIgnoreNulls:
			opt.IgnoreNulls = true
		case OptionLimit:
			i64 := v.Value.(int64)
			opt.Limit = &i64
		case OptionOrderBy:
			opt.OrderBy = append(opt.OrderBy, v.Value.(*OrderBy))
		default:
			filteredArgs = append(filteredArgs, arg)
			continue
		}
	}
	return filteredArgs, opt
}
