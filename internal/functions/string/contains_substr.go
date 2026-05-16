package string

import (
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// CONTAINS_SUBSTR is the BigQuery substring search. Per BQ docs the
// match is case-insensitive, NFKC-normalised. We approximate via
// strings.EqualFold-style ASCII case folding, which covers the
// overwhelming majority of real call sites; full NFKC normalisation
// is a follow-up.
//
// Returns NULL if either operand is NULL. Returns BOOL otherwise.
func CONTAINS_SUBSTR(haystack, needle string) (value.Value, error) {
	return value.BoolValue(
		strings.Contains(strings.ToLower(haystack), strings.ToLower(needle)),
	), nil
}

var BindContainsSubstr = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	hay, err := a.ToString()
	if err != nil {
		return nil, err
	}
	needle, err := b.ToString()
	if err != nil {
		return nil, err
	}
	return CONTAINS_SUBSTR(hay, needle)
})
