package structs

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func STRUCT_FIELD(v value.Value, idx int) (value.Value, error) {
	sv, err := v.ToStruct()
	if err != nil {
		return nil, err
	}
	// Struct values passed in from the Go driver (via map[string]interface{})
	// may not include every declared field — e.g. a table schema of
	// STRUCT<fieldA, fieldB> receiving a row that only sets fieldB. The
	// stored value then has fewer entries than the declared field count,
	// and an out-of-range index that the analyzer emitted against the
	// declared schema would panic here. Treat out-of-range reads as NULL,
	// matching BigQuery's behavior for missing struct fields.
	if idx < 0 || idx >= len(sv.Values) {
		return nil, nil
	}
	return sv.Values[idx], nil
}

var BindStructField = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	i64, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return STRUCT_FIELD(a, int(i64))
})
