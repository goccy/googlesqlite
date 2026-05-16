// Unit tests for the Bind* surface of the hll package.
// Expected outputs follow upstream GoogleSQL HLL++ reference
// (docs/third_party/googlesql-docs/hll_functions.md): EXTRACT/MERGE return
// the estimated cardinality of the sketches produced by INIT.
package hll

import (
	"math/big"
	"testing"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func newAgg(t *testing.T, ctor func() *helper.Aggregator) *helper.Aggregator {
	t.Helper()
	a := ctor()
	if a == nil {
		t.Fatal("ctor returned nil")
	}
	return a
}

func TestHllInitDone_Empty(t *testing.T) {
	t.Parallel()

	// HLL_COUNT.INIT over an empty input returns NULL (no Step calls
	// means the inner sketch was never allocated).
	a := newAgg(t, BindHllCountInit())
	raw, err := a.Done()
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if raw != nil {
		t.Errorf("empty Done = %v; want nil", raw)
	}
}

func TestHllInit_ProducesSketch(t *testing.T) {
	t.Parallel()

	// Run HLL_COUNT.INIT(x) over five distinct strings — Done must
	// return a non-nil byte sketch.
	a := newAgg(t, BindHllCountInit())
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		if err := a.Step(s); err != nil {
			t.Fatalf("Step(%q): %v", s, err)
		}
	}
	raw, err := a.Done()
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if raw == nil {
		t.Fatalf("expected non-nil sketch")
	}
	sketch, ok := raw.([]byte)
	if !ok {
		// EncodeValue may return string for BytesValue; try DecodeValue.
		v, err := value.DecodeValue(raw)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		bv, err := v.ToBytes()
		if err != nil {
			t.Fatalf("ToBytes: %v", err)
		}
		sketch = bv
	}
	if len(sketch) == 0 {
		t.Fatalf("empty sketch returned")
	}

	// EXTRACT returns the estimated cardinality of the sketch.
	got, err := BindHllCountExtract(value.BytesValue(sketch))
	if err != nil {
		t.Fatalf("EXTRACT: %v", err)
	}
	n, err := got.ToInt64()
	if err != nil {
		t.Fatalf("ToInt64: %v", err)
	}
	if n < 1 {
		t.Errorf("EXTRACT cardinality = %d; expected >= 1", n)
	}

	// EXTRACT(NULL) -> NULL.
	got, err = BindHllCountExtract(nil)
	if err != nil {
		t.Fatalf("EXTRACT NULL: %v", err)
	}
	if got != nil {
		t.Errorf("EXTRACT(NULL) = %v; want nil", got)
	}
}

// HLL_COUNT.MERGE over INIT sketches yields an int64 cardinality.
// HLL_COUNT.MERGE_PARTIAL returns the merged sketch as bytes.
func TestHllMergeAndMergePartial(t *testing.T) {
	t.Parallel()

	// mkSketchEncoded runs HLL_COUNT.INIT over the given strings and
	// returns the over-the-wire (base64-encoded ValueLayout) form
	// suitable to hand back into a downstream Aggregator.Step.
	mkSketchEncoded := func(vals ...string) any {
		a := newAgg(t, BindHllCountInit())
		for _, s := range vals {
			if err := a.Step(s); err != nil {
				t.Fatalf("step: %v", err)
			}
		}
		raw, err := a.Done()
		if err != nil {
			t.Fatalf("done: %v", err)
		}
		if raw == nil {
			t.Fatalf("nil sketch")
		}
		return raw
	}

	s1 := mkSketchEncoded("a", "b", "c")
	s2 := mkSketchEncoded("c", "d", "e")

	// MERGE: feed both sketches, expect cardinality >= 3 (union has at
	// least the size of either side; HLL is approximate so we don't
	// assert == 5).
	m := newAgg(t, BindHllCountMerge())
	if err := m.Step(s1); err != nil {
		t.Fatalf("merge step1: %v", err)
	}
	if err := m.Step(s2); err != nil {
		t.Fatalf("merge step2: %v", err)
	}
	// NULL step is skipped.
	if err := m.Step(nil); err != nil {
		t.Fatalf("merge step nil: %v", err)
	}
	raw, err := m.Done()
	if err != nil {
		t.Fatalf("merge done: %v", err)
	}
	n, ok := raw.(int64)
	if !ok {
		t.Fatalf("MERGE Done = %T; want int64", raw)
	}
	if n < 3 {
		t.Errorf("MERGE cardinality = %d; expected >= 3", n)
	}

	// MERGE of zero rows -> 0 (per helper Done).
	m = newAgg(t, BindHllCountMerge())
	raw, err = m.Done()
	if err != nil {
		t.Fatalf("merge empty: %v", err)
	}
	if n, ok := raw.(int64); !ok || n != 0 {
		t.Errorf("empty MERGE = %v; want int64(0)", raw)
	}

	// MERGE_PARTIAL: returns a non-nil sketch over both inputs.
	mp := newAgg(t, BindHllCountMergePartial())
	if err := mp.Step(s1); err != nil {
		t.Fatalf("partial step1: %v", err)
	}
	if err := mp.Step(s2); err != nil {
		t.Fatalf("partial step2: %v", err)
	}
	if err := mp.Step(nil); err != nil {
		t.Fatalf("partial step nil: %v", err)
	}
	raw, err = mp.Done()
	if err != nil {
		t.Fatalf("partial done: %v", err)
	}
	if raw == nil {
		t.Fatalf("expected non-nil merged sketch")
	}

	// MERGE_PARTIAL with zero rows -> NULL.
	mp = newAgg(t, BindHllCountMergePartial())
	raw, err = mp.Done()
	if err != nil {
		t.Fatalf("partial empty done: %v", err)
	}
	if raw != nil {
		t.Errorf("partial empty = %v; want nil", raw)
	}
}

// HLL_COUNT.INIT supports an optional second precision argument.
func TestHllInit_WithPrecision(t *testing.T) {
	t.Parallel()

	a := newAgg(t, BindHllCountInit())
	if err := a.Step("a", int64(12)); err != nil {
		t.Fatalf("step with precision: %v", err)
	}
	if err := a.Step("b", int64(12)); err != nil {
		t.Fatalf("step2 with precision: %v", err)
	}
	raw, err := a.Done()
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if raw == nil {
		t.Errorf("expected non-nil sketch with precision arg")
	}

	// NULL precision is skipped (treated as no-op).
	a = newAgg(t, BindHllCountInit())
	if err := a.Step("a", nil); err != nil {
		t.Fatalf("step nil precision: %v", err)
	}
	raw, err = a.Done()
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if raw != nil {
		t.Errorf("nil precision step should not initialize sketch; got %v", raw)
	}
}

// EXTRACT over an invalid byte sketch surfaces the underlying error.
func TestHllExtract_Invalid(t *testing.T) {
	t.Parallel()
	if _, err := BindHllCountExtract(value.BytesValue([]byte{0xff, 0xff, 0xff, 0xff})); err == nil {
		t.Errorf("expected error for garbage sketch")
	}
}

// HLL_COUNT.INIT accepts INT64 / NUMERIC / STRING / BYTES inputs.
// The Step switch dispatches on the value type, so exercising each
// branch keeps coverage of the hashing paths.
func TestHllInit_AllInputTypes(t *testing.T) {
	t.Parallel()

	// Integer input.
	a := newAgg(t, BindHllCountInit())
	if err := a.Step(int64(7)); err != nil {
		t.Fatalf("int Step: %v", err)
	}
	raw, err := a.Done()
	if err != nil {
		t.Fatalf("int Done: %v", err)
	}
	if raw == nil {
		t.Errorf("int Done = nil; want sketch")
	}

	// Numeric input — pass the encoded form so the aggregator decodes
	// it to a *NumericValue and the numeric switch arm runs.
	enc, err := value.EncodeValue(&value.NumericValue{Rat: big.NewRat(1, 2)})
	if err != nil {
		t.Fatalf("encode numeric: %v", err)
	}
	a = newAgg(t, BindHllCountInit())
	if err := a.Step(enc); err != nil {
		t.Fatalf("numeric Step: %v", err)
	}
	raw, err = a.Done()
	if err != nil {
		t.Fatalf("numeric Done: %v", err)
	}
	if raw == nil {
		t.Errorf("numeric Done = nil; want sketch")
	}

	// Bytes input — encode a BytesValue.
	enc, err = value.EncodeValue(value.BytesValue([]byte("hello")))
	if err != nil {
		t.Fatalf("encode bytes: %v", err)
	}
	a = newAgg(t, BindHllCountInit())
	if err := a.Step(enc); err != nil {
		t.Fatalf("bytes Step: %v", err)
	}
	raw, err = a.Done()
	if err != nil {
		t.Fatalf("bytes Done: %v", err)
	}
	if raw == nil {
		t.Errorf("bytes Done = nil; want sketch")
	}
}
