package numeric

import (
	"fmt"
	"math/big"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func PARSE_NUMERIC(numeric string) (value.Value, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(numeric); !ok {
		return nil, fmt.Errorf("unexpected numeric literal: %s", numeric)
	}
	return &value.NumericValue{Rat: r}, nil
}

var BindParseNumeric = helper.Scalar1(func(a value.Value) (value.Value, error) {
	numeric, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return PARSE_NUMERIC(numeric)
})
