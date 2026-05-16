package text_analysis

import (
	"fmt"
	"math"
	"sort"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/functions/window"
	"github.com/goccy/googlesqlite/internal/value"
)

// TF_IDF emits per-row TF-IDF scores for each term in the row's
// tokenised document, evaluated across the analytic window. The
// implementation follows the BigQuery formula:
//
//	TF(t, d)   = count(t in d) / |d|
//	IDF(t, N)  = log((1 + N) / (1 + df(t))) + 1
//	TF-IDF     = TF * IDF
//
// where N is the number of documents in the window and df(t) is
// the number of documents containing term t. The returned ARRAY
// is sorted by score descending, with ties broken by term to
// keep the test surface deterministic.
//
// max_distinct_tokens caps the global dictionary; terms ranked
// outside the cap (by descending df) are merged into the
// `__unknown__` bucket. frequency_threshold drops terms whose
// in-document occurrence is below the threshold from the result
// of the current row only (other rows still see them).
type TF_IDF struct {
	maxDistinctTokens int64
	freqThreshold     int64
}

func (f *TF_IDF) Step(arr value.Value, _ int64, _ int64, opt *window.WindowFuncStatus, agg *window.WindowFuncAggregatedStatus) error {
	return agg.Step(arr, opt)
}

func (f *TF_IDF) Done(agg *window.WindowFuncAggregatedStatus) (value.Value, error) {
	var result value.Value
	if err := agg.Done(func(values []value.Value, start, end int) error {
		if len(values) == 0 {
			result = &value.ArrayValue{}
			return nil
		}
		// 1. Compute per-document term counts and overall df.
		type docCount struct {
			counts map[string]int64
			total  int64
		}
		docs := make([]docCount, 0, len(values))
		df := map[string]int64{}
		for _, v := range values {
			dc := docCount{counts: map[string]int64{}}
			if v != nil {
				arr, err := v.ToArray()
				if err != nil {
					return err
				}
				for _, tok := range arr.Values {
					if tok == nil {
						continue
					}
					s, err := tok.ToString()
					if err != nil {
						return err
					}
					dc.counts[s]++
					dc.total++
				}
			}
			docs = append(docs, dc)
			for term := range dc.counts {
				df[term]++
			}
		}
		// 2. Decide which terms survive the max_distinct_tokens
		// cap. Rank by df descending, term ascending as tiebreak.
		type termDF struct {
			term string
			df   int64
		}
		ranked := make([]termDF, 0, len(df))
		for t, c := range df {
			ranked = append(ranked, termDF{t, c})
		}
		sort.Slice(ranked, func(i, j int) bool {
			if ranked[i].df != ranked[j].df {
				return ranked[i].df > ranked[j].df
			}
			return ranked[i].term < ranked[j].term
		})
		allowed := map[string]struct{}{}
		cap := f.maxDistinctTokens
		if cap <= 0 || cap > int64(len(ranked)) {
			cap = int64(len(ranked))
		}
		for i := int64(0); i < cap; i++ {
			allowed[ranked[i].term] = struct{}{}
		}
		// 3. Pick the current row's doc; merge filtered terms
		// into the unknown bucket then compute scores.
		rowIdx := int(agg.RowID - 1)
		if rowIdx < 0 || rowIdx >= len(docs) {
			result = &value.ArrayValue{}
			return nil
		}
		_ = start
		_ = end
		dc := docs[rowIdx]
		nDocs := int64(len(docs))
		mergedCounts := map[string]int64{}
		unknownCount := int64(0)
		for term, count := range dc.counts {
			if _, ok := allowed[term]; ok {
				mergedCounts[term] += count
				continue
			}
			unknownCount += count
		}
		if unknownCount > 0 {
			mergedCounts["__unknown__"] = unknownCount
		}
		// Recompute df for filtered (term -> #docs that contain it
		// among the allowed set; unknown groups everyone else).
		mergedDF := map[string]int64{}
		for term := range mergedCounts {
			if term == "__unknown__" {
				continue
			}
			mergedDF[term] = df[term]
		}
		if unknownCount > 0 {
			unknownDF := int64(0)
			for _, d := range docs {
				has := false
				for term := range d.counts {
					if _, ok := allowed[term]; !ok {
						has = true
						break
					}
				}
				if has {
					unknownDF++
				}
			}
			mergedDF["__unknown__"] = unknownDF
		}
		// 4. Emit scores in deterministic order.
		out := &value.ArrayValue{Values: make([]value.Value, 0, len(mergedCounts))}
		type scored struct {
			term  string
			score float64
		}
		rows := make([]scored, 0, len(mergedCounts))
		for term, count := range mergedCounts {
			if count < f.freqThreshold {
				continue
			}
			tf := float64(count) / float64(dc.total)
			idf := math.Log(float64(1+nDocs)/float64(1+mergedDF[term])) + 1
			rows = append(rows, scored{term, tf * idf})
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].score != rows[j].score {
				return rows[i].score > rows[j].score
			}
			return rows[i].term < rows[j].term
		})
		for _, r := range rows {
			out.Values = append(out.Values, &value.StructValue{
				Keys:   []string{"term", "score"},
				Values: []value.Value{value.StringValue(r.term), value.FloatValue(r.score)},
				M: map[string]value.Value{
					"term":  value.StringValue(r.term),
					"score": value.FloatValue(r.score),
				},
			})
		}
		result = out
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// BindTfIdf wires the per-row tokens (ARRAY<STRING>) plus optional
// INT64 max_distinct_tokens / frequency_threshold args.
func BindTfIdf() func() *window.WindowAggregator {
	return func() *window.WindowAggregator {
		fn := &TF_IDF{maxDistinctTokens: 32000, freqThreshold: 0}
		return window.NewWindowAggregator(
			func(args []value.Value, opt *window.WindowFuncStatus, agg *window.WindowFuncAggregatedStatus) error {
				if len(args) == 0 {
					return fmt.Errorf("TF_IDF: missing tokens argument")
				}
				if len(args) >= 2 && args[1] != nil {
					n, err := args[1].ToInt64()
					if err != nil {
						return err
					}
					fn.maxDistinctTokens = n
				}
				if len(args) >= 3 && args[2] != nil {
					n, err := args[2].ToInt64()
					if err != nil {
						return err
					}
					fn.freqThreshold = n
				}
				if helper.ExistsNull(args[:1]) {
					return agg.Step(nil, opt)
				}
				return fn.Step(args[0], fn.maxDistinctTokens, fn.freqThreshold, opt, agg)
			},
			func(agg *window.WindowFuncAggregatedStatus) (value.Value, error) {
				return fn.Done(agg)
			},
		)
	}
}
