package text_analysis

import (
	"fmt"
	"sort"

	"github.com/goccy/googlesqlite/internal/value"
)

// BAG_OF_WORDS counts per-term occurrences in an already-tokenised
// document. NULL elements are kept and counted as a distinct
// "null" term — they appear in the output with a NULL term field.
// Output is ARRAY<STRUCT<term STRING, count INT64>>; we sort with
// NULL term first then by term lexicographically to give a stable
// result the test harness can compare against.
func BAG_OF_WORDS(tokens *value.ArrayValue) (value.Value, error) {
	if tokens == nil {
		return nil, nil
	}
	nullCount := int64(0)
	counts := map[string]int64{}
	for _, t := range tokens.Values {
		if t == nil {
			nullCount++
			continue
		}
		s, err := t.ToString()
		if err != nil {
			return nil, err
		}
		counts[s]++
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]value.Value, 0, len(counts)+1)
	mkRow := func(term value.Value, c int64) *value.StructValue {
		return &value.StructValue{
			Keys:   []string{"term", "count"},
			Values: []value.Value{term, value.IntValue(c)},
			M: map[string]value.Value{
				"term":  term,
				"count": value.IntValue(c),
			},
		}
	}
	if nullCount > 0 {
		out = append(out, mkRow(nil, nullCount))
	}
	for _, k := range keys {
		out = append(out, mkRow(value.StringValue(k), counts[k]))
	}
	return &value.ArrayValue{Values: out}, nil
}

func BindBagOfWords(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("BAG_OF_WORDS: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	arr, err := args[0].ToArray()
	if err != nil {
		return nil, err
	}
	return BAG_OF_WORDS(arr)
}
