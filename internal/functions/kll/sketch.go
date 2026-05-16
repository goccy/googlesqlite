// Package kll implements KLL quantile sketches for the BigQuery
// `kll_quantiles.*` family.
//
// Implementation strategy: full-data sample store. This is NOT a
// true Karnin-Lang-Liberty compactor — it just keeps every input
// value and computes quantiles exactly at extract time. For an
// in-emulator workload that is the right trade-off:
//
//   - Correctness is exact (no approximation error).
//   - Memory cost is linear in input size; emulator queries don't
//     run on TB-scale data so this is fine.
//   - Binary format is googlesqlite-specific (gob-encoded slice
//     of sorted float64) — interoperability with the real
//     BigQuery sketch format is not provided.
//
// The advantage over emitting per-element BYTES is that downstream
// MERGE / EXTRACT calls can compose: a MERGE concatenates two
// sketches' values; EXTRACT sorts and indexes by rank.
package kll

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"sort"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// sketch is the on-the-wire form. Int64 inputs are widened to
// float64 for storage; the per-extract conversion back to INT64
// rounds to nearest.
type sketch struct {
	Values []float64
}

func (s *sketch) marshal() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshal(data []byte) (*sketch, error) {
	if len(data) == 0 {
		return &sketch{}, nil
	}
	var s sketch
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// quantile returns the value at the given rank (0.0..1.0) using
// the same rank rule as BigQuery's KLL_QUANTILES.EXTRACT_POINT:
// rank * (n-1) interpolation, but since EXTRACT_POINT returns a
// stored sketch element, we round to the nearest stored index.
func (s *sketch) quantile(rank float64) (float64, bool) {
	if len(s.Values) == 0 {
		return 0, false
	}
	sort.Float64s(s.Values)
	idx := int(math.Round(rank * float64(len(s.Values)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.Values) {
		idx = len(s.Values) - 1
	}
	return s.Values[idx], true
}

// quantilesArray returns numQuantiles + 1 evenly-spaced
// quantiles from rank 0 to rank 1, mirroring KLL_QUANTILES.EXTRACT.
func (s *sketch) quantilesArray(numQuantiles int64) ([]float64, error) {
	if numQuantiles < 1 {
		return nil, fmt.Errorf("KLL_QUANTILES: num_quantiles must be >= 1")
	}
	out := make([]float64, 0, numQuantiles+1)
	for i := int64(0); i <= numQuantiles; i++ {
		rank := float64(i) / float64(numQuantiles)
		q, ok := s.quantile(rank)
		if !ok {
			return nil, nil
		}
		out = append(out, q)
	}
	return out, nil
}

// ---------------- aggregator ----------------

type aggregator struct {
	values []float64
}

func (a *aggregator) Step(v value.Value) error {
	if v == nil {
		return nil
	}
	f, err := v.ToFloat64()
	if err != nil {
		return err
	}
	a.values = append(a.values, f)
	return nil
}

func (a *aggregator) Done() (value.Value, error) {
	s := &sketch{Values: a.values}
	b, err := s.marshal()
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

// ---------------- merge aggregator ----------------

type mergeAggregator struct {
	values       []float64
	numQuantiles int64
	asInt        bool
	hasNumPicked bool
}

func (m *mergeAggregator) Step(v value.Value) error {
	if v == nil {
		return nil
	}
	b, err := v.ToBytes()
	if err != nil {
		return err
	}
	s, err := unmarshal(b)
	if err != nil {
		return err
	}
	m.values = append(m.values, s.Values...)
	return nil
}

// Done for merge_partial returns the merged sketch as bytes.
func (m *mergeAggregator) DonePartial() (value.Value, error) {
	s := &sketch{Values: m.values}
	b, err := s.marshal()
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

// Done for merge_int64 / merge_float64 returns ARRAY<quantile>.
func (m *mergeAggregator) DoneArray() (value.Value, error) {
	if len(m.values) == 0 {
		return nil, nil
	}
	s := &sketch{Values: m.values}
	qs, err := s.quantilesArray(m.numQuantiles)
	if err != nil {
		return nil, err
	}
	if qs == nil {
		return nil, nil
	}
	out := make([]value.Value, 0, len(qs))
	for _, q := range qs {
		if m.asInt {
			out = append(out, value.IntValue(int64(math.Round(q))))
		} else {
			out = append(out, value.FloatValue(q))
		}
	}
	return &value.ArrayValue{Values: out}, nil
}

// ---------------- bindings ----------------

// BindInit returns a constructor for the INIT aggregator. Same
// function works for both int64 and float64 because we store as
// float64 internally.
func BindInit() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		a := &aggregator{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return a.Step(args[0])
			},
			func() (value.Value, error) {
				return a.Done()
			},
		)
	}
}

// BindMergeInt64 / BindMergeFloat64 merge sketches and emit an
// ARRAY of `num_quantiles + 1` quantiles. The second argument
// (numQuantiles) is required by the analyzer signature.
func bindMergeArray(asInt bool) func() *helper.Aggregator {
	return func() *helper.Aggregator {
		m := &mergeAggregator{asInt: asInt}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if !m.hasNumPicked && len(args) >= 2 && args[1] != nil {
					n, err := args[1].ToInt64()
					if err != nil {
						return err
					}
					m.numQuantiles = n
					m.hasNumPicked = true
				}
				return m.Step(args[0])
			},
			func() (value.Value, error) {
				return m.DoneArray()
			},
		)
	}
}

func BindMergeInt64() func() *helper.Aggregator   { return bindMergeArray(true) }
func BindMergeFloat64() func() *helper.Aggregator { return bindMergeArray(false) }

// BindMergePartial returns the merged sketch as BYTES.
func BindMergePartial() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		m := &mergeAggregator{}
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				return m.Step(args[0])
			},
			func() (value.Value, error) {
				return m.DonePartial()
			},
		)
	}
}

// BindExtract returns an ARRAY<value> of `num_quantiles + 1` evenly
// spaced quantiles. Element type is derived from the registered
// name (int64 / float64 variant).
func BindExtractInt64(args ...value.Value) (value.Value, error) {
	return extractArray(args, true)
}

func BindExtractFloat64(args ...value.Value) (value.Value, error) {
	return extractArray(args, false)
}

func extractArray(args []value.Value, asInt bool) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("KLL_QUANTILES.EXTRACT: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	s, err := unmarshal(b)
	if err != nil {
		return nil, err
	}
	n, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	qs, err := s.quantilesArray(n)
	if err != nil {
		return nil, err
	}
	if qs == nil {
		return nil, nil
	}
	out := make([]value.Value, 0, len(qs))
	for _, q := range qs {
		if asInt {
			out = append(out, value.IntValue(int64(math.Round(q))))
		} else {
			out = append(out, value.FloatValue(q))
		}
	}
	return &value.ArrayValue{Values: out}, nil
}

// BindExtractPoint returns a single quantile value at the given rank.
func BindExtractPointInt64(args ...value.Value) (value.Value, error) {
	return extractPoint(args, true)
}

func BindExtractPointFloat64(args ...value.Value) (value.Value, error) {
	return extractPoint(args, false)
}

func extractPoint(args []value.Value, asInt bool) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("KLL_QUANTILES.EXTRACT_POINT: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	s, err := unmarshal(b)
	if err != nil {
		return nil, err
	}
	rank, err := args[1].ToFloat64()
	if err != nil {
		return nil, err
	}
	q, ok := s.quantile(rank)
	if !ok {
		return nil, nil
	}
	if asInt {
		return value.IntValue(int64(math.Round(q))), nil
	}
	return value.FloatValue(q), nil
}

// MergePoint combines merge + extract_point so callers can do
// merge-then-quantile in one aggregate.
type mergePointAggregator struct {
	mergeAggregator
	rank  float64
	asInt bool
}

func bindMergePoint(asInt bool) func() *helper.Aggregator {
	return func() *helper.Aggregator {
		mp := &mergePointAggregator{asInt: asInt}
		first := true
		return helper.NewAggregator(
			func(args []value.Value, opt *helper.Option) error {
				if first && len(args) >= 2 && args[1] != nil {
					f, err := args[1].ToFloat64()
					if err != nil {
						return err
					}
					mp.rank = f
					first = false
				}
				return mp.Step(args[0])
			},
			func() (value.Value, error) {
				s := &sketch{Values: mp.values}
				q, ok := s.quantile(mp.rank)
				if !ok {
					return nil, nil
				}
				if mp.asInt {
					return value.IntValue(int64(math.Round(q))), nil
				}
				return value.FloatValue(q), nil
			},
		)
	}
}

func BindMergePointInt64() func() *helper.Aggregator {
	return bindMergePoint(true)
}

func BindMergePointFloat64() func() *helper.Aggregator {
	return bindMergePoint(false)
}
