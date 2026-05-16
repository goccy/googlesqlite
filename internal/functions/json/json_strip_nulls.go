package json

import (
	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// JSON_STRIP_NULLS removes object members whose value is JSON null,
// recursively. Arrays and scalars pass through unchanged. If the
// entire input value is null, returns SQL NULL.
//
// Reference: BigQuery / GoogleSQL JSON_STRIP_NULLS.
func JSON_STRIP_NULLS(v string) (value.Value, error) {
	var node any
	if err := json.Unmarshal([]byte(v), &node); err != nil {
		return nil, err
	}
	stripped, drop := stripNulls(node)
	if drop {
		return nil, nil
	}
	out, err := json.Marshal(stripped)
	if err != nil {
		return nil, err
	}
	return value.JsonValue(string(out)), nil
}

// stripNulls walks a decoded JSON value, recursively removes nil
// entries from maps, and reports whether the (possibly transformed)
// value should itself be dropped. Drop is signalled only for the
// top-level untyped nil; arrays and scalars are kept verbatim because
// JSON_STRIP_NULLS preserves array elements and scalar leaves.
func stripNulls(node any) (any, bool) {
	switch v := node.(type) {
	case nil:
		return nil, true
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, child := range v {
			cleaned, drop := stripNulls(child)
			if drop {
				continue
			}
			out[k] = cleaned
		}
		return out, false
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			cleaned, drop := stripNulls(child)
			if drop {
				out[i] = nil
				continue
			}
			out[i] = cleaned
		}
		return out, false
	default:
		return v, false
	}
}

// BindJsonStripNulls reads only the first argument; the analyzer may
// append trailing option arguments, so it must not enforce arity.
func BindJsonStripNulls(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return JSON_STRIP_NULLS(s)
}
