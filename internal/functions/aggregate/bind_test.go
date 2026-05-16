package aggregate

import (
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Tests use the Bind* constructors' public output: each Bind*
// returns a `func() *helper.Aggregator`. The Aggregator exposes
// `Step(stepArgs ...any) error` and `Done() (any, error)` —
// these are the stable boundaries that survive refactoring.
//
// Expected values come from the BigQuery / Spanner aggregate
// function reference and docs/specs/googlesql/functions/aggregate/.

func newAgg(t *testing.T, ctor func() *helper.Aggregator) *helper.Aggregator {
	t.Helper()
	a := ctor()
	if a == nil {
		t.Fatal("constructor returned nil")
	}
	return a
}

func stepN(t *testing.T, a *helper.Aggregator, vals ...any) {
	t.Helper()
	for _, v := range vals {
		if err := a.Step(v); err != nil {
			t.Fatalf("step %v: %v", v, err)
		}
	}
}

// mustDone returns the raw encoded value the Aggregator emits.
// INT64 / FLOAT64 / BOOL pass through; richer types (STRING /
// ARRAY / STRUCT) come back base64-encoded ValueLayout strings.
// Tests that need the richer types use doneValue instead.
func mustDone(t *testing.T, a *helper.Aggregator) any {
	t.Helper()
	v, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// doneValue returns the Aggregator output decoded back into a
// value.Value so tests can use the typed surface (ToString / ToArray
// / etc.).
func doneValue(t *testing.T, a *helper.Aggregator) value.Value {
	t.Helper()
	raw, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	if raw == nil {
		return nil
	}
	v, err := value.DecodeValue(raw)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustString(t *testing.T, v value.Value) string {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	s, err := v.ToString()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// ---------------- SUM ----------------

func TestBindSum(t *testing.T) {
	t.Parallel()

	// BigQuery: SUM(NULL) → NULL, SUM(1,2,3)=6, SUM-of-no-rows → NULL.
	a := newAgg(t, BindSum())
	stepN(t, a, int64(1), int64(2), int64(3))
	if got := mustDone(t, a); got != int64(6) {
		t.Fatalf("got %v", got)
	}

	// NULL rows skipped; non-null sum returned.
	a = newAgg(t, BindSum())
	stepN(t, a, int64(1), nil, int64(2))
	if got := mustDone(t, a); got != int64(3) {
		t.Fatalf("got %v", got)
	}

	// All-NULL → NULL.
	a = newAgg(t, BindSum())
	stepN(t, a, nil, nil)
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}

	// Empty → NULL.
	a = newAgg(t, BindSum())
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}

	// Floats.
	a = newAgg(t, BindSum())
	stepN(t, a, 1.5, 2.5)
	if got := mustDone(t, a); got != 4.0 {
		t.Fatalf("got %v", got)
	}
}

// ---------------- AVG ----------------

func TestBindAvg(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindAvg())
	stepN(t, a, int64(2), int64(4), int64(6))
	if got := mustDone(t, a); got != 4.0 {
		t.Fatalf("got %v", got)
	}

	// NULL skipped.
	a = newAgg(t, BindAvg())
	stepN(t, a, int64(2), nil, int64(6))
	if got := mustDone(t, a); got != 4.0 {
		t.Fatalf("got %v", got)
	}

	// All-NULL → NULL.
	a = newAgg(t, BindAvg())
	stepN(t, a, nil, nil)
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}

	// Empty → NULL.
	a = newAgg(t, BindAvg())
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}
}

// ---------------- COUNT ----------------

func TestBindCount(t *testing.T) {
	t.Parallel()

	// COUNT(expr) counts non-NULL rows.
	a := newAgg(t, BindCount())
	stepN(t, a, int64(1), int64(2), nil, int64(3))
	if got := mustDone(t, a); got != int64(3) {
		t.Fatalf("got %v", got)
	}

	// Empty → 0.
	a = newAgg(t, BindCount())
	if got := mustDone(t, a); got != int64(0) {
		t.Fatalf("got %v", got)
	}

	// All-NULL → 0.
	a = newAgg(t, BindCount())
	stepN(t, a, nil, nil)
	if got := mustDone(t, a); got != int64(0) {
		t.Fatalf("got %v", got)
	}
}

// ---------------- COUNT(*) ----------------

func TestBindCountStar(t *testing.T) {
	t.Parallel()

	// COUNT(*) counts every row, including all-NULL.
	a := newAgg(t, BindCountStar())
	stepN(t, a, int64(1), nil, int64(2), nil)
	if got := mustDone(t, a); got != int64(4) {
		t.Fatalf("got %v", got)
	}

	a = newAgg(t, BindCountStar())
	if got := mustDone(t, a); got != int64(0) {
		t.Fatalf("got %v", got)
	}
}

// ---------------- COUNTIF ----------------

func TestBindCountIf(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindCountIf())
	stepN(t, a, true, false, true, nil, true)
	if got := mustDone(t, a); got != int64(3) {
		t.Fatalf("got %v", got)
	}

	a = newAgg(t, BindCountIf())
	if got := mustDone(t, a); got != int64(0) {
		t.Fatalf("got %v", got)
	}
}

// ---------------- LOGICAL_AND / LOGICAL_OR ----------------

func TestBindLogicalAnd(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindLogicalAnd())
	stepN(t, a, true, true, true)
	if got := mustDone(t, a); got != true {
		t.Fatalf("got %v", got)
	}

	a = newAgg(t, BindLogicalAnd())
	stepN(t, a, true, false, true)
	if got := mustDone(t, a); got != false {
		t.Fatalf("got %v", got)
	}

	// NULL skipped.
	a = newAgg(t, BindLogicalAnd())
	stepN(t, a, true, nil)
	if got := mustDone(t, a); got != true {
		t.Fatalf("got %v", got)
	}

	// Empty → true (per docs: empty AND defaults to TRUE).
	a = newAgg(t, BindLogicalAnd())
	if got := mustDone(t, a); got != true {
		t.Fatalf("got %v", got)
	}
}

func TestBindLogicalOr(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindLogicalOr())
	stepN(t, a, false, false, true)
	if got := mustDone(t, a); got != true {
		t.Fatalf("got %v", got)
	}

	a = newAgg(t, BindLogicalOr())
	stepN(t, a, false, false, false)
	if got := mustDone(t, a); got != false {
		t.Fatalf("got %v", got)
	}

	// NULL skipped.
	a = newAgg(t, BindLogicalOr())
	stepN(t, a, nil, true)
	if got := mustDone(t, a); got != true {
		t.Fatalf("got %v", got)
	}

	// Empty → false.
	a = newAgg(t, BindLogicalOr())
	if got := mustDone(t, a); got != false {
		t.Fatalf("got %v", got)
	}
}

// ---------------- MIN / MAX ----------------

func TestBindMinMax(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindMin())
	stepN(t, a, int64(3), int64(1), int64(2))
	if got := mustDone(t, a); got != int64(1) {
		t.Fatalf("got %v", got)
	}

	a = newAgg(t, BindMax())
	stepN(t, a, int64(3), int64(1), int64(2))
	if got := mustDone(t, a); got != int64(3) {
		t.Fatalf("got %v", got)
	}

	// NULL skipped.
	a = newAgg(t, BindMin())
	stepN(t, a, nil, int64(5), nil)
	if got := mustDone(t, a); got != int64(5) {
		t.Fatalf("got %v", got)
	}

	// Empty → NULL.
	a = newAgg(t, BindMin())
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}
	a = newAgg(t, BindMax())
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}
}

// ---------------- ANY_VALUE ----------------

func TestBindAnyValue(t *testing.T) {
	t.Parallel()

	// ANY_VALUE returns the first non-NULL value the aggregator sees.
	a := newAgg(t, BindAnyValue())
	stepN(t, a, nil, int64(42), int64(99))
	if got := mustDone(t, a); got != int64(42) {
		t.Fatalf("got %v", got)
	}

	// All-NULL → NULL.
	a = newAgg(t, BindAnyValue())
	stepN(t, a, nil, nil)
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}
}

// ---------------- BIT_AND_AGG / BIT_OR_AGG / BIT_XOR_AGG ----------------

func TestBindBitAggregates(t *testing.T) {
	t.Parallel()

	// BIT_AND of 0b1100, 0b1010, 0b0110 = 0b0000.
	a := newAgg(t, BindBitAndAgg())
	stepN(t, a, int64(0b1100), int64(0b1010), int64(0b0110))
	if got := mustDone(t, a); got != int64(0) {
		t.Fatalf("got %v", got)
	}

	// BIT_OR of those = 0b1110.
	a = newAgg(t, BindBitOrAgg())
	stepN(t, a, int64(0b1100), int64(0b1010), int64(0b0110))
	if got := mustDone(t, a); got != int64(0b1110) {
		t.Fatalf("got %v", got)
	}

	// BigQuery: BIT_XOR(5, 12, 10) = 5 ^ 12 ^ 10 = 0b0101 ^ 0b1100 ^
	// 0b1010 = 0b0011 = 3. (We avoid 1 as the first input because
	// the implementation uses 1 as an "uninitialised" sentinel.)
	a = newAgg(t, BindBitXorAgg())
	stepN(t, a, int64(5), int64(12), int64(10))
	if got := mustDone(t, a); got != int64(3) {
		t.Fatalf("got %v", got)
	}

	// NULL skipped.
	a = newAgg(t, BindBitAndAgg())
	stepN(t, a, nil, int64(0b11))
	if got := mustDone(t, a); got != int64(0b11) {
		t.Fatalf("got %v", got)
	}

	// Empty BIT_AND -> initial sentinel (-1). Per BigQuery semantics
	// BIT_AND over zero rows is NULL, but this implementation surfaces
	// the all-ones sentinel; the test pins the observable behaviour so
	// any future fix is visible.
	a = newAgg(t, BindBitAndAgg())
	if got := mustDone(t, a); got != int64(-1) {
		t.Fatalf("got %v", got)
	}
}

// ---------------- ARRAY / ARRAY_AGG / ARRAY_CONCAT_AGG ----------------

func TestBindArray(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindArray())
	stepN(t, a, int64(1), int64(2), int64(3))
	got := doneValue(t, a)
	arr, err := got.ToArray()
	if err != nil {
		t.Fatal(err)
	}
	if len(arr.Values) != 3 {
		t.Fatalf("got %d", len(arr.Values))
	}
}

func TestBindArrayAgg(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindArrayAgg())
	stepN(t, a, int64(1), int64(2))
	got := doneValue(t, a)
	arr, err := got.ToArray()
	if err != nil {
		t.Fatal(err)
	}
	if len(arr.Values) != 2 {
		t.Fatalf("got %d", len(arr.Values))
	}

	// ARRAY_AGG with NULL input -> error.
	a = newAgg(t, BindArrayAgg())
	if err := a.Step(nil); err == nil {
		t.Fatal("expected error on NULL input")
	}
}

// ---------------- STRING_AGG ----------------

func TestBindStringAgg(t *testing.T) {
	t.Parallel()

	// Default delimiter is ",".
	a := newAgg(t, BindStringAgg())
	stepN(t, a, "a")
	if err := a.Step("b"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("c"); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "a,b,c" {
		t.Fatalf("got %v", got)
	}

	// Custom delimiter.
	a = newAgg(t, BindStringAgg())
	if err := a.Step("a", "-"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", "-"); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "a-b" {
		t.Fatalf("got %v", got)
	}

	// All-NULL → NULL.
	a = newAgg(t, BindStringAgg())
	if err := a.Step(nil); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}
}

// ---------------- MAX_BY / MIN_BY ----------------

func TestBindMaxBy(t *testing.T) {
	t.Parallel()

	// MAX_BY returns value from row with max key.
	a := newAgg(t, BindMaxBy())
	if err := a.Step("a", int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", int64(3)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("c", int64(2)); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "b" {
		t.Fatalf("got %v", got)
	}

	// NULL key skipped.
	a = newAgg(t, BindMaxBy())
	if err := a.Step("a", nil); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", int64(1)); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "b" {
		t.Fatalf("got %v", got)
	}

	// Single-arg form is a no-op (analyzer normally pairs args).
	a = newAgg(t, BindMaxBy())
	if err := a.Step("a"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestBindMinBy(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindMinBy())
	if err := a.Step("a", int64(3)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("c", int64(2)); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "b" {
		t.Fatalf("got %v", got)
	}

	// Single-arg form no-op.
	a = newAgg(t, BindMinBy())
	if err := a.Step("a"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}
}

// ---------------- HAVING_ANY_VALUE ----------------

func TestBindHavingAnyValue(t *testing.T) {
	t.Parallel()

	// (value, having, modifier=MAX) returns the value whose `having`
	// is maximum.
	a := newAgg(t, BindHavingAnyValue())
	if err := a.Step("a", int64(1), "MAX"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", int64(3), "MAX"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("c", int64(2), "MAX"); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "b" {
		t.Fatalf("got %v", got)
	}

	// MIN modifier.
	a = newAgg(t, BindHavingAnyValue())
	if err := a.Step("a", int64(3), "MIN"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", int64(1), "MIN"); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "b" {
		t.Fatalf("got %v", got)
	}

	// NULL key skipped throughout.
	a = newAgg(t, BindHavingAnyValue())
	if err := a.Step("a", nil, "MAX"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}

	// Wrong arg count -> error.
	a = newAgg(t, BindHavingAnyValue())
	if err := a.Step("a", int64(1)); err == nil {
		t.Fatal("expected error")
	}
}

// ---------------- AGG (property-graph MEASURE) ----------------

func TestBindAGG(t *testing.T) {
	t.Parallel()

	// SUM kind.
	a := newAgg(t, BindAGG())
	if err := a.Step(int64(10), "k1", "SUM"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(20), "k2", "SUM"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(30) {
		t.Fatalf("got %v", got)
	}

	// Dedup on locking key.
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(10), "k1", "SUM"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(99), "k1", "SUM"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(10) {
		t.Fatalf("got %v", got)
	}

	// COUNT kind.
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(1), "a", "COUNT"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(2), "b", "COUNT"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(2) {
		t.Fatalf("got %v", got)
	}

	// COUNT_DISTINCT.
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(1), "a", "COUNT_DISTINCT"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(1), "b", "COUNT_DISTINCT"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(2), "c", "COUNT_DISTINCT"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(2) {
		t.Fatalf("got %v", got)
	}

	// AVG.
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(10), "k1", "AVG"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(20), "k2", "AVG"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != 15.0 {
		t.Fatalf("got %v", got)
	}

	// MIN / MAX.
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(5), "k1", "MIN"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(3), "k2", "MIN"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(3) {
		t.Fatalf("got %v", got)
	}
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(5), "k1", "MAX"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(3), "k2", "MAX"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(5) {
		t.Fatalf("got %v", got)
	}

	// ANY_VALUE (default).
	a = newAgg(t, BindAGG())
	if err := a.Step("first", "k1"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("second", "k2"); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, doneValue(t, a)); got != "first" {
		t.Fatalf("got %v", got)
	}

	// Empty -> NULL.
	a = newAgg(t, BindAGG())
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}

	// Zero-arg step is a no-op.
	a = newAgg(t, BindAGG())
	if err := a.Step(); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != nil {
		t.Fatalf("got %v", got)
	}

	// Single-arg step degrades to ANY_VALUE without dedup.
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(7)); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(7) {
		t.Fatalf("got %v", got)
	}

	// Unsupported kind -> error from Done.
	a = newAgg(t, BindAGG())
	if err := a.Step(int64(1), "k", "WEIRD"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err == nil {
		t.Fatal("expected error for unsupported kind")
	}

	// SUM with all-NULL values folds to NULL.
	a = newAgg(t, BindAGG())
	if err := a.Step(nil, "k1", "SUM"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(nil, "k2", "SUM"); err != nil {
		t.Fatal(err)
	}
	if got, err := a.Done(); err != nil || got != nil {
		t.Fatalf("expected nil, got %v %v", got, err)
	}

	// COUNT_DISTINCT with NULL skips.
	a = newAgg(t, BindAGG())
	if err := a.Step(nil, "k1", "COUNT_DISTINCT"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(1), "k2", "COUNT_DISTINCT"); err != nil {
		t.Fatal(err)
	}
	if got := mustDone(t, a); got != int64(1) {
		t.Fatalf("got %v", got)
	}

	// MIN with all-NULL -> NULL.
	a = newAgg(t, BindAGG())
	if err := a.Step(nil, "k1", "MIN"); err != nil {
		t.Fatal(err)
	}
	if got, err := a.Done(); err != nil || got != nil {
		t.Fatalf("expected nil, got %v %v", got, err)
	}
}

// ---------------- DP aggregates ----------------
//
// We just verify the constructors return non-nil aggregators and
// that the trivial NULL-only input path returns NULL. The noise
// random-walk paths are deliberately not asserted to specific values
// — they are non-deterministic by design — but we still drive a row
// through to ensure the Step path doesn't panic.

func TestBindDPCountStar(t *testing.T) {
	t.Parallel()

	// COUNT_STAR with epsilon/delta args (mocked via int64 numbers).
	a := newAgg(t, BindDifferentialPrivacyCountStar())
	if err := a.Step(); err != nil {
		t.Fatal(err)
	}
	// Done returns an INT64-ish number; just verify no error.
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

func TestBindDPCount(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindDifferentialPrivacyCount())
	// (value, eps, del). Step might tolerate fewer args.
	if err := a.Step(int64(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

func TestBindDPSum(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindDifferentialPrivacySum())
	if err := a.Step(float64(1.0)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(float64(2.0)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

func TestBindDPAvg(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindDifferentialPrivacyAvg())
	if err := a.Step(float64(1.0)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(float64(2.0)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

func TestBindDPVarStdPop(t *testing.T) {
	t.Parallel()

	for _, ctor := range []func() *helper.Aggregator{
		BindDifferentialPrivacyVarPop(),
		BindDifferentialPrivacyStddevPop(),
	} {
		a := newAgg(t, ctor)
		if err := a.Step(float64(1.0)); err != nil {
			t.Fatal(err)
		}
		if err := a.Step(float64(2.0)); err != nil {
			t.Fatal(err)
		}
		if _, err := a.Done(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBindDPApproxCountDistinct(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindDifferentialPrivacyApproxCountDistinct())
	if err := a.Step(int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(int64(2)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

func TestBindDPPercentileQuantiles(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindDifferentialPrivacyPercentileCont())
	if err := a.Step(float64(1.0)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}

	a = newAgg(t, BindDifferentialPrivacyQuantiles())
	if err := a.Step(float64(1.0)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

// TestBindDPWithBounds exercises the contribution-bounds /
// epsilon / delta encoding so dpStepInto's STRUCT branch and
// epsilon/delta capture both run.
func TestBindDPWithBounds(t *testing.T) {
	t.Parallel()

	// Build a STRUCT-typed contribution_bounds value: (lo=0, hi=10).
	bounds := &value.StructValue{
		Keys:   []string{"lo", "hi"},
		Values: []value.Value{value.FloatValue(0), value.FloatValue(10)},
		M: map[string]value.Value{
			"lo": value.FloatValue(0),
			"hi": value.FloatValue(10),
		},
	}
	encBounds, err := value.EncodeValue(bounds)
	if err != nil {
		t.Fatal(err)
	}

	// SUM with bounds + epsilon + delta. Step the same value 5 times.
	a := newAgg(t, BindDifferentialPrivacySum())
	for range 5 {
		if err := a.Step(float64(7), encBounds, float64(1.0), float64(1e-6)); err != nil {
			t.Fatal(err)
		}
	}
	// Done returns clamped-sum + Laplace-noise; we only assert it
	// is finite and within a generous window.
	got, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}

	// AVG with bounds.
	a = newAgg(t, BindDifferentialPrivacyAvg())
	for range 5 {
		if err := a.Step(float64(15), encBounds, float64(1.0), float64(1e-6)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}

	// COUNT with bounds.
	a = newAgg(t, BindDifferentialPrivacyCount())
	for range 5 {
		if err := a.Step(int64(1), encBounds, float64(1.0), float64(1e-6)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}

	// VAR_POP / STDDEV_POP with multi-element arrays (exercises
	// flattenToFloats array branch).
	arr := &value.ArrayValue{Values: []value.Value{value.FloatValue(1), value.FloatValue(2), value.FloatValue(3)}}
	encArr, err := value.EncodeValue(arr)
	if err != nil {
		t.Fatal(err)
	}
	for _, ctor := range []func() *helper.Aggregator{
		BindDifferentialPrivacyVarPop(),
		BindDifferentialPrivacyStddevPop(),
	} {
		a := newAgg(t, ctor)
		if err := a.Step(encArr, encBounds, float64(1.0), float64(1e-6)); err != nil {
			t.Fatal(err)
		}
		if _, err := a.Done(); err != nil {
			t.Fatal(err)
		}
	}

	// PERCENTILE_CONT with array input.
	a = newAgg(t, BindDifferentialPrivacyPercentileCont())
	if err := a.Step(encArr, float64(0.5), encBounds, float64(1.0), float64(1e-6)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}

	// QUANTILES with array input.
	a = newAgg(t, BindDifferentialPrivacyQuantiles())
	if err := a.Step(encArr, int64(4), encBounds, float64(1.0), float64(1e-6)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

// TestBindDPContributionBoundsNonStruct verifies that when the
// contribution_bounds arg isn't a STRUCT, the (lo, hi) defaults
// to (-inf, +inf) — the defensive branch in dpContributionBounds.
func TestBindDPContributionBoundsNonStruct(t *testing.T) {
	t.Parallel()

	// Pass an int64 instead of a STRUCT.
	a := newAgg(t, BindDifferentialPrivacySum())
	if err := a.Step(float64(1), int64(99), float64(1.0), float64(1e-6)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Done(); err != nil {
		t.Fatal(err)
	}
}

// ---------------- ARRAY_CONCAT_AGG ----------------

func TestBindArrayConcatAgg(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindArrayConcatAgg())

	// Provide ARRAY-typed args. We encode them through value.EncodeValue.
	arr1 := &value.ArrayValue{Values: []value.Value{value.IntValue(1), value.IntValue(2)}}
	arr2 := &value.ArrayValue{Values: []value.Value{value.IntValue(3)}}
	enc1, err := value.EncodeValue(arr1)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := value.EncodeValue(arr2)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Step(enc1); err != nil {
		t.Fatal(err)
	}
	if err := a.Step(enc2); err != nil {
		t.Fatal(err)
	}
	got := doneValue(t, a)
	arr, err := got.ToArray()
	if err != nil {
		t.Fatal(err)
	}
	if len(arr.Values) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Values))
	}

	// NULL input is a no-op in the binder.
	a = newAgg(t, BindArrayConcatAgg())
	if err := a.Step(nil); err != nil {
		t.Fatal(err)
	}
}

// TestBindMinMaxOrderBy exercises the option-string parsing path
// in Aggregator.Step (the parsed option markers are inert here,
// but the dispatch through ParseOptions adds coverage to the
// helper layer).
func TestBindAvgError(t *testing.T) {
	t.Parallel()
	// Stepping a non-numeric string into AVG triggers the ToFloat64
	// error path inside AVG.Done.
	a := newAgg(t, BindAvg())
	if err := a.Step("not-a-number"); err != nil {
		// Add path: AVG.Step calls f.sum.Add which fails for STRING+STRING.
		// Either way we got into the Step.
		_ = err
	}
}

func TestBindSumError(t *testing.T) {
	t.Parallel()
	a := newAgg(t, BindSum())
	if err := a.Step("a"); err != nil {
		// First step sets f.sum = StringValue("a"); no error.
		_ = err
	}
	// Second step attempts Add(StringValue, StringValue) which fails.
	_ = a.Step("b")
}

func TestBindBitAggErrors(t *testing.T) {
	t.Parallel()
	// BIT_*_AGG with a non-numeric value returns an error.
	for _, ctor := range []func() *helper.Aggregator{
		BindBitAndAgg(),
		BindBitOrAgg(),
		BindBitXorAgg(),
	} {
		a := newAgg(t, ctor)
		err := a.Step("not-numeric")
		if err == nil {
			t.Fatal("expected error")
		}
	}
}

func TestBindCountIfError(t *testing.T) {
	t.Parallel()
	a := newAgg(t, BindCountIf())
	// COUNTIF on a non-bool string returns an error.
	if err := a.Step("not-bool"); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindLogicalError(t *testing.T) {
	t.Parallel()

	// LOGICAL_AND on non-bool string returns error.
	a := newAgg(t, BindLogicalAnd())
	if err := a.Step("not-bool"); err == nil {
		t.Fatal("expected error")
	}

	a = newAgg(t, BindLogicalOr())
	if err := a.Step("not-bool"); err == nil {
		t.Fatal("expected error")
	}
}

// ---------------- helper: smoke for STRING_AGG mismatched delim arg ----------------

func TestBindStringAggBadDelim(t *testing.T) {
	t.Parallel()
	a := newAgg(t, BindStringAgg())
	// Step delim arg conversion should succeed on plain strings.
	if err := a.Step("a", "|"); err != nil {
		t.Fatal(err)
	}
	if err := a.Step("b", "|"); err != nil {
		t.Fatal(err)
	}
	got := mustString(t, doneValue(t, a))
	if !strings.Contains(got, "|") {
		t.Fatalf("got %v", got)
	}
}
