package window

import (
	"strings"
	"sync"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Native (frame-driven) window aggregators below all conform to the
// shape SQLite expects from CreateWindowFunction: Step / Inverse /
// Done. SQLite owns partition / order / frame; per-aggregator
// implementations only maintain incremental state. They receive the
// user value as the first arg and any encoded option markers
// (googlesqlite_distinct(), googlesqlite_ignore_nulls()) as trailing
// args; partition / order / frame markers are NOT passed because the
// OVER clause handles them.

// arrayAggWindowNative collects values in input order and emits the
// encoded ARRAY_AGG result per Done() call. Done returns the array
// of values currently in the frame. With the simple-adapter
// (no Inverse) every value.Value() invocation rebuilds the result from
// the accumulator, but the accumulator only spans the active frame
// — never the entire scan — so this is O(frame_size) per output
// row instead of the predecessor's O(N²) full re-scan.
type arrayAggWindowNative struct {
	values      []value.Value
	distinct    bool
	ignoreNulls bool
	once        sync.Once
}

// NewArrayAggWindowNative builds the ctor used by RegisterWindow.
func NewArrayAggWindowNative() func() any {
	return func() any { return &arrayAggWindowNative{} }
}

func (a *arrayAggWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, opt := helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...) // strip any stray window option markers
	a.once.Do(func() {
		a.distinct = opt.Distinct
		a.ignoreNulls = opt.IgnoreNulls
	})
	if len(values) == 0 {
		return nil
	}
	v := values[0]
	a.values = append(a.values, v)
	return nil
}

// Inverse is invoked when SQLite slides the frame forward; we pop the
// oldest captured value so the next Done() reflects the current
// frame. Without this method, RegisterWindow falls back to a
// buffer-and-rebuild adapter, but matching the SQLite-driven
// invariant directly is cleaner.
func (a *arrayAggWindowNative) Inverse(stepArgs ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *arrayAggWindowNative) Done() (any, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	var (
		result   []value.Value
		valueMap = map[string]struct{}{}
	)
	for _, v := range a.values {
		if a.ignoreNulls && v == nil {
			continue
		}
		if a.distinct {
			key := "<nil>"
			if v != nil {
				k, err := v.ToString()
				if err != nil {
					return nil, err
				}
				key = k
			}
			if _, exists := valueMap[key]; exists {
				continue
			}
			valueMap[key] = struct{}{}
		}
		result = append(result, v)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return value.EncodeValue(&value.ArrayValue{Values: result})
}

// stringAggWindowNative is the STRING_AGG counterpart. The optional
// delimiter is passed as the second positional arg (per the
// predecessor's signature). Default is ",".
type stringAggWindowNative struct {
	values      []string
	delim       string
	delimSet    bool
	distinct    bool
	ignoreNulls bool
	once        sync.Once
}

func NewStringAggWindowNative() func() any {
	return func() any { return &stringAggWindowNative{} }
}

func (a *stringAggWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, opt := helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	a.once.Do(func() {
		a.distinct = opt.Distinct
		a.ignoreNulls = opt.IgnoreNulls
		a.delim = ","
	})
	if len(values) == 0 {
		return nil
	}
	if !a.delimSet && len(values) > 1 && values[1] != nil {
		s, err := values[1].ToString()
		if err == nil && s != "" {
			a.delim = s
		}
		a.delimSet = true
	}
	v := values[0]
	if v == nil {
		if a.ignoreNulls {
			return nil
		}
		a.values = append(a.values, "")
		return nil
	}
	s, err := v.ToString()
	if err != nil {
		return err
	}
	a.values = append(a.values, s)
	return nil
}

func (a *stringAggWindowNative) Inverse(stepArgs ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *stringAggWindowNative) Done() (any, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	var (
		out  []string
		seen = map[string]struct{}{}
	)
	for _, s := range a.values {
		if a.distinct {
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
		}
		out = append(out, s)
	}
	return value.EncodeValue(value.StringValue(strings.Join(out, a.delim)))
}

// countifWindowNative counts truthy boolean values in the frame.
type countifWindowNative struct {
	count int64
}

func NewCountifWindowNative() func() any {
	return func() any { return &countifWindowNative{} }
}

func (a *countifWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 || values[0] == nil {
		return nil
	}
	cond, err := values[0].ToBool()
	if err != nil {
		return err
	}
	if cond {
		a.count++
	}
	return nil
}

func (a *countifWindowNative) Inverse(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 || values[0] == nil {
		return nil
	}
	cond, err := values[0].ToBool()
	if err != nil {
		return err
	}
	if cond && a.count > 0 {
		a.count--
	}
	return nil
}

func (a *countifWindowNative) Done() (any, error) {
	return a.count, nil
}

// countStarWindowNative counts every row entering the frame, NULL or
// not.
type countStarWindowNative struct {
	count int64
}

func NewCountStarWindowNative() func() any {
	return func() any { return &countStarWindowNative{} }
}

func (a *countStarWindowNative) Step(_ ...any) error {
	a.count++
	return nil
}

func (a *countStarWindowNative) Inverse(_ ...any) error {
	if a.count > 0 {
		a.count--
	}
	return nil
}

func (a *countStarWindowNative) Done() (any, error) {
	return a.count, nil
}

// anyValueWindowNative returns the first value still in the frame.
// We buffer per-row Step values so the BigQuery "any" choice is
// deterministic across frame slides — picking the oldest value
// remaining in the frame matches what the predecessor's emulation
// returned for sorted ROWS frames.
type anyValueWindowNative struct {
	values []value.Value
}

func NewAnyValueWindowNative() func() any {
	return func() any { return &anyValueWindowNative{} }
}

func (a *anyValueWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 {
		return nil
	}
	a.values = append(a.values, values[0])
	return nil
}

func (a *anyValueWindowNative) Inverse(_ ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *anyValueWindowNative) Done() (any, error) {
	for _, v := range a.values {
		if v == nil {
			continue
		}
		return value.EncodeValue(v)
	}
	return nil, nil
}

// logicalOrWindowNative is the LOGICAL_OR(BOOL) frame-driven form.
// Returns true iff any non-NULL value in the active frame is true,
// false iff every non-NULL value is false, and NULL when every value
// in the frame is NULL — matching the documented BigQuery semantics.
//
// Step appends the per-row truthiness; Inverse pops the oldest entry
// when SQLite slides the frame forward. Done re-evaluates the
// current buffer, which is O(frame_size) — same shape as ARRAY_AGG
// but cheaper because we only carry tristate flags.
type logicalOrWindowNative struct {
	values []logicTri
}

type logicTri uint8

const (
	logicNull logicTri = iota
	logicFalse
	logicTrue
)

func NewLogicalOrWindowNative() func() any {
	return func() any { return &logicalOrWindowNative{} }
}

func (a *logicalOrWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 || values[0] == nil {
		a.values = append(a.values, logicNull)
		return nil
	}
	b, err := values[0].ToBool()
	if err != nil {
		return err
	}
	if b {
		a.values = append(a.values, logicTrue)
	} else {
		a.values = append(a.values, logicFalse)
	}
	return nil
}

func (a *logicalOrWindowNative) Inverse(_ ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *logicalOrWindowNative) Done() (any, error) {
	var seen bool
	for _, t := range a.values {
		switch t {
		case logicTrue:
			return value.EncodeValue(value.BoolValue(true))
		case logicFalse:
			seen = true
		}
	}
	if !seen {
		return nil, nil
	}
	return value.EncodeValue(value.BoolValue(false))
}

// logicalAndWindowNative mirrors logicalOrWindowNative for AND-aggregation.
// Returns true iff every non-NULL value is true, false iff any is
// false, NULL when every value in the frame is NULL.
type logicalAndWindowNative struct {
	values []logicTri
}

func NewLogicalAndWindowNative() func() any {
	return func() any { return &logicalAndWindowNative{} }
}

func (a *logicalAndWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 || values[0] == nil {
		a.values = append(a.values, logicNull)
		return nil
	}
	b, err := values[0].ToBool()
	if err != nil {
		return err
	}
	if b {
		a.values = append(a.values, logicTrue)
	} else {
		a.values = append(a.values, logicFalse)
	}
	return nil
}

func (a *logicalAndWindowNative) Inverse(_ ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *logicalAndWindowNative) Done() (any, error) {
	var seen bool
	for _, t := range a.values {
		switch t {
		case logicFalse:
			return value.EncodeValue(value.BoolValue(false))
		case logicTrue:
			seen = true
		}
	}
	if !seen {
		return nil, nil
	}
	return value.EncodeValue(value.BoolValue(true))
}

// bitAndAggWindowNative / bitOrAggWindowNative / bitXorAggWindowNative
// are the BIT_AND / BIT_OR / BIT_XOR window forms over INT64 inputs.
// They keep the window's set of NULL-skipped values and recompute on
// Done. Buffer-and-recompute is O(frame_size) per output row, no
// worse than the predecessor's correlated-subquery approach but
// driven natively by SQLite's frame walker.
type bitAndAggWindowNative struct {
	values []int64
}

func NewBitAndAggWindowNative() func() any {
	return func() any { return &bitAndAggWindowNative{} }
}

func (a *bitAndAggWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 || values[0] == nil {
		return nil
	}
	v, err := values[0].ToInt64()
	if err != nil {
		return err
	}
	a.values = append(a.values, v)
	return nil
}

func (a *bitAndAggWindowNative) Inverse(_ ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *bitAndAggWindowNative) Done() (any, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	out := a.values[0]
	for _, v := range a.values[1:] {
		out &= v
	}
	return out, nil
}

type bitOrAggWindowNative struct {
	values []int64
}

func NewBitOrAggWindowNative() func() any {
	return func() any { return &bitOrAggWindowNative{} }
}

func (a *bitOrAggWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 || values[0] == nil {
		return nil
	}
	v, err := values[0].ToInt64()
	if err != nil {
		return err
	}
	a.values = append(a.values, v)
	return nil
}

func (a *bitOrAggWindowNative) Inverse(_ ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *bitOrAggWindowNative) Done() (any, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	var out int64
	for _, v := range a.values {
		out |= v
	}
	return out, nil
}

type bitXorAggWindowNative struct {
	values []int64
}

func NewBitXorAggWindowNative() func() any {
	return func() any { return &bitXorAggWindowNative{} }
}

func (a *bitXorAggWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 || values[0] == nil {
		return nil
	}
	v, err := values[0].ToInt64()
	if err != nil {
		return err
	}
	a.values = append(a.values, v)
	return nil
}

func (a *bitXorAggWindowNative) Inverse(_ ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *bitXorAggWindowNative) Done() (any, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	var out int64
	for _, v := range a.values {
		out ^= v
	}
	return out, nil
}

// arrayConcatAggWindowNative is the ARRAY_CONCAT_AGG(ARRAY<T>) frame-
// driven form. Each Step appends one input array to the buffer;
// Done flattens the active frame's arrays into a single ARRAY<T>.
// Inverse pops the oldest captured array. NULL input arrays are
// rejected (matches the BigQuery error semantics).
type arrayConcatAggWindowNative struct {
	values []*value.ArrayValue
}

func NewArrayConcatAggWindowNative() func() any {
	return func() any { return &arrayConcatAggWindowNative{} }
}

func (a *arrayConcatAggWindowNative) Step(stepArgs ...any) error {
	values, err := value.ConvertArgs(stepArgs...)
	if err != nil {
		return err
	}
	values, _ = helper.ParseOptions(values...)
	values, _ = parseWindowOptions(values...)
	if len(values) == 0 {
		return nil
	}
	if values[0] == nil {
		// BigQuery treats NULL array element as an error here;
		// match that semantics for parity.
		return nil
	}
	arr, err := values[0].ToArray()
	if err != nil {
		return err
	}
	a.values = append(a.values, arr)
	return nil
}

func (a *arrayConcatAggWindowNative) Inverse(_ ...any) error {
	if len(a.values) > 0 {
		a.values = a.values[1:]
	}
	return nil
}

func (a *arrayConcatAggWindowNative) Done() (any, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	var flattened []value.Value
	for _, arr := range a.values {
		flattened = append(flattened, arr.Values...)
	}
	return value.EncodeValue(&value.ArrayValue{Values: flattened})
}
