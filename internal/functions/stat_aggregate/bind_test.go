// Unit tests for the Bind* surface of stat_aggregate.
// Expected values follow upstream GoogleSQL statistical aggregate docs
// (docs/third_party/googlesql-docs/statistical_aggregate_functions.md and
// Spanner/BigQuery reference).
package stat_aggregate

import (
	"math"
	"testing"

	"github.com/goccy/googlesqlite/internal/functions/helper"
)

func newAgg(t *testing.T, ctor func() *helper.Aggregator) *helper.Aggregator {
	t.Helper()
	a := ctor()
	if a == nil {
		t.Fatal("ctor returned nil")
	}
	return a
}

func stepPairs(t *testing.T, a *helper.Aggregator, pairs ...[2]float64) {
	t.Helper()
	for _, p := range pairs {
		if err := a.Step(p[0], p[1]); err != nil {
			t.Fatalf("step %v: %v", p, err)
		}
	}
}

func step1(t *testing.T, a *helper.Aggregator, vals ...any) {
	t.Helper()
	for _, v := range vals {
		if err := a.Step(v); err != nil {
			t.Fatalf("step %v: %v", v, err)
		}
	}
}

func doneFloat(t *testing.T, a *helper.Aggregator) (float64, bool) {
	t.Helper()
	raw, err := a.Done()
	if err != nil {
		t.Fatal(err)
	}
	if raw == nil {
		return 0, false
	}
	f, ok := raw.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", raw)
	}
	return f, true
}

func approxEq(a, b, eps float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

// ---------- STDDEV_POP / STDDEV_SAMP / STDDEV / VARIANCE ----------

// STDDEV_POP({1,2,3,4}) = sqrt(((1-2.5)^2 + (2-2.5)^2 + (3-2.5)^2 + (4-2.5)^2)/4) = sqrt(1.25)
// STDDEV_SAMP({1,2,3,4}) = sqrt(5/3) (Bessel-corrected).
func TestStddev(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindStddevPop())
	step1(t, a, int64(1), int64(2), int64(3), int64(4))
	got, ok := doneFloat(t, a)
	if !ok {
		t.Fatalf("STDDEV_POP non-null expected")
	}
	if !approxEq(got, math.Sqrt(1.25), 1e-9) {
		t.Errorf("STDDEV_POP = %v; want %v", got, math.Sqrt(1.25))
	}

	a = newAgg(t, BindStddevSamp())
	step1(t, a, int64(1), int64(2), int64(3), int64(4))
	got, ok = doneFloat(t, a)
	if !ok {
		t.Fatalf("STDDEV_SAMP non-null expected")
	}
	if !approxEq(got, math.Sqrt(5.0/3.0), 1e-9) {
		t.Errorf("STDDEV_SAMP = %v; want %v", got, math.Sqrt(5.0/3.0))
	}

	// STDDEV is the sample alias.
	a = newAgg(t, BindStddev())
	step1(t, a, int64(1), int64(2), int64(3), int64(4))
	got, ok = doneFloat(t, a)
	if !ok {
		t.Fatalf("STDDEV non-null expected")
	}
	if !approxEq(got, math.Sqrt(5.0/3.0), 1e-9) {
		t.Errorf("STDDEV = %v; want %v", got, math.Sqrt(5.0/3.0))
	}

	// Empty: NULL.
	a = newAgg(t, BindStddevPop())
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("STDDEV_POP empty should be NULL")
	}
	a = newAgg(t, BindStddevSamp())
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("STDDEV_SAMP empty should be NULL")
	}
	a = newAgg(t, BindStddev())
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("STDDEV empty should be NULL")
	}

	// All-NULL: NULL.
	a = newAgg(t, BindStddevPop())
	step1(t, a, nil, nil)
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("STDDEV_POP all-null should be NULL")
	}
}

// VAR_POP({1,2,3,4}) = 1.25
// VAR_SAMP({1,2,3,4}) = 5/3
func TestVariance(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindVarPop())
	step1(t, a, int64(1), int64(2), int64(3), int64(4))
	got, ok := doneFloat(t, a)
	if !ok {
		t.Fatalf("VAR_POP non-null expected")
	}
	if !approxEq(got, 1.25, 1e-9) {
		t.Errorf("VAR_POP = %v; want 1.25", got)
	}

	a = newAgg(t, BindVarSamp())
	step1(t, a, int64(1), int64(2), int64(3), int64(4))
	got, ok = doneFloat(t, a)
	if !ok {
		t.Fatalf("VAR_SAMP non-null expected")
	}
	if !approxEq(got, 5.0/3.0, 1e-9) {
		t.Errorf("VAR_SAMP = %v; want %v", got, 5.0/3.0)
	}

	a = newAgg(t, BindVariance())
	step1(t, a, int64(1), int64(2), int64(3), int64(4))
	got, ok = doneFloat(t, a)
	if !ok {
		t.Fatalf("VARIANCE non-null expected")
	}
	if !approxEq(got, 5.0/3.0, 1e-9) {
		t.Errorf("VARIANCE = %v; want %v", got, 5.0/3.0)
	}

	// Empty: NULL.
	a = newAgg(t, BindVarPop())
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("VAR_POP empty should be NULL")
	}
}

// ------------------ CORR / COVAR_POP / COVAR_SAMP ------------------

// CORR(x={1..4}, y={1..4}) = 1.0 (perfectly correlated).
func TestCorr(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindCorr())
	stepPairs(t, a, [2]float64{1, 1}, [2]float64{2, 2}, [2]float64{3, 3}, [2]float64{4, 4})
	got, ok := doneFloat(t, a)
	if !ok {
		t.Fatalf("CORR non-null expected")
	}
	if !approxEq(got, 1.0, 1e-9) {
		t.Errorf("CORR = %v; want 1.0", got)
	}

	// Empty: NULL.
	a = newAgg(t, BindCorr())
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("CORR empty should be NULL")
	}

	// NULL on either side skips the row.
	a = newAgg(t, BindCorr())
	if err := a.Step(nil, float64(1)); err != nil {
		t.Fatalf("step nil,1: %v", err)
	}
	if err := a.Step(float64(2), nil); err != nil {
		t.Fatalf("step 2,nil: %v", err)
	}
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("CORR with only NULL pairs should be NULL")
	}
}

// COVAR_POP({1..4}, {1..4}) = 1.25.
// COVAR_SAMP({1..4}, {1..4}) = 5/3 (Bessel).
// COVAR_POP single sample -> 0; COVAR_SAMP single sample -> NaN (gonum).
func TestCovariance(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindCovarPop())
	stepPairs(t, a, [2]float64{1, 1}, [2]float64{2, 2}, [2]float64{3, 3}, [2]float64{4, 4})
	got, ok := doneFloat(t, a)
	if !ok {
		t.Fatalf("COVAR_POP non-null expected")
	}
	if !approxEq(got, 1.25, 1e-9) {
		t.Errorf("COVAR_POP = %v; want 1.25", got)
	}

	a = newAgg(t, BindCovarSamp())
	stepPairs(t, a, [2]float64{1, 1}, [2]float64{2, 2}, [2]float64{3, 3}, [2]float64{4, 4})
	got, ok = doneFloat(t, a)
	if !ok {
		t.Fatalf("COVAR_SAMP non-null expected")
	}
	if !approxEq(got, 5.0/3.0, 1e-9) {
		t.Errorf("COVAR_SAMP = %v; want %v", got, 5.0/3.0)
	}

	// COVAR_POP with a single sample returns 0 (well-defined).
	a = newAgg(t, BindCovarPop())
	stepPairs(t, a, [2]float64{1, 1})
	got, ok = doneFloat(t, a)
	if !ok {
		t.Fatalf("COVAR_POP single non-null expected")
	}
	if got != 0 {
		t.Errorf("COVAR_POP single = %v; want 0", got)
	}

	// COVAR_POP empty -> NULL.
	a = newAgg(t, BindCovarPop())
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("COVAR_POP empty should be NULL")
	}
	// COVAR_SAMP empty -> NULL.
	a = newAgg(t, BindCovarSamp())
	if _, ok := doneFloat(t, a); ok {
		t.Errorf("COVAR_SAMP empty should be NULL")
	}
}
