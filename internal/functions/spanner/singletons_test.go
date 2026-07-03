package spanner

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestBindPendingCommitTimestamp(t *testing.T) {
	t.Parallel()

	v, err := BindPendingCommitTimestamp()
	if err != nil {
		t.Fatal(err)
	}
	tv, ok := v.(value.TimestampValue)
	if !ok {
		t.Fatalf("expected TimestampValue, got %T", v)
	}
	// Sentinel is far in the future.
	if time.Time(tv).Year() != 9999 {
		t.Fatalf("got year %d", time.Time(tv).Year())
	}
}

func TestBindBitReverse(t *testing.T) {
	t.Parallel()

	// Spec reference: spanner docs `BIT_REVERSE(1, false)` reverses 64 bits.
	// 1 with 64-bit reversal is `1 << 63` = -9223372036854775808.
	got, err := BindBitReverse(value.IntValue(1), value.BoolValue(false))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != -1<<63 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// With preserve_width on 1, width=1 -> result is 1.
	got, err = BindBitReverse(value.IntValue(1), value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// BIT_REVERSE(0, true) = 0.
	got, err = BindBitReverse(value.IntValue(0), value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// Single-arg form defaults to 64-bit reversal.
	got, err = BindBitReverse(value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != -1<<63 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if v, _ := BindBitReverse(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindBitReverse(); err == nil {
		t.Fatal("expected arg count error")
	}
	if _, err := BindBitReverse(value.IntValue(1), value.IntValue(2), value.IntValue(3)); err == nil {
		t.Fatal("expected arg count error")
	}
}

func arr(vs ...float64) *value.ArrayValue {
	out := make([]value.Value, len(vs))
	for i, v := range vs {
		out[i] = value.FloatValue(v)
	}
	return &value.ArrayValue{Values: out}
}

func TestBindDotProduct(t *testing.T) {
	t.Parallel()

	// dot([1,2,3], [4,5,6]) = 32.
	got, err := BindDotProduct(arr(1, 2, 3), arr(4, 5, 6))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 32 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// Mismatched lengths -> error.
	if _, err := BindDotProduct(arr(1, 2), arr(1, 2, 3)); err == nil {
		t.Fatal("expected length-mismatch error")
	}

	// Non-array -> error.
	if _, err := BindDotProduct(value.IntValue(1), arr(1)); err == nil {
		t.Fatal("expected non-array error")
	}

	if _, err := BindDotProduct(arr(1)); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindDotProduct(nil, arr(1)); v != nil {
		t.Fatal("expected null")
	}

	// NULL element in array -> null result.
	withNull := &value.ArrayValue{Values: []value.Value{value.FloatValue(1), nil}}
	if v, _ := BindDotProduct(withNull, arr(1, 2)); v != nil {
		t.Fatal("expected null result for element-null")
	}
}

func TestBindApproxDistances(t *testing.T) {
	t.Parallel()

	// approx dot product wraps BindDotProduct.
	got, err := BindApproxDotProduct(arr(1, 0), arr(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// Euclidean distance between [1,0] and [0,1] = sqrt(2).
	got, err = BindApproxEuclideanDistance(arr(1, 0), arr(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(mustFloat64(t, got)-math.Sqrt(2)) > 1e-9 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// Cosine distance between identical vectors = 0.
	got, err = BindApproxCosineDistance(arr(1, 2), arr(1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(mustFloat64(t, got)) > 1e-9 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// Cosine distance for orthogonal vectors = 1.
	got, err = BindApproxCosineDistance(arr(1, 0), arr(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(mustFloat64(t, got)-1) > 1e-9 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// Zero-magnitude vector -> error.
	if _, err := BindApproxCosineDistance(arr(0, 0), arr(1, 1)); err == nil {
		t.Fatal("expected zero-magnitude error")
	}

	if _, err := BindApproxEuclideanDistance(arr(1)); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindApproxEuclideanDistance(nil, arr(1)); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindApproxCosineDistance(nil, arr(1)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindApproxCosineDistance(arr(1), arr(1, 2)); err == nil {
		t.Fatal("expected length-mismatch error")
	}
}

func TestBindZstdRoundTrip(t *testing.T) {
	t.Parallel()

	payload := []byte("hello world hello world hello world")
	for _, level := range []int64{1, 10, 20} {
		cv, err := BindZstdCompress(value.BytesValue(payload), value.IntValue(level))
		if err != nil {
			t.Fatalf("compress lvl %d: %v", level, err)
		}
		dv, err := BindZstdDecompressToBytes(cv)
		if err != nil {
			t.Fatalf("decompress lvl %d: %v", level, err)
		}
		if string(mustBytes(t, dv)) != string(payload) {
			t.Fatalf("lvl %d: round-trip mismatch", level)
		}
		sv, err := BindZstdDecompressToString(cv)
		if err != nil {
			t.Fatal(err)
		}
		if mustString(t, sv) != string(payload) {
			t.Fatal("string round-trip mismatch")
		}
	}

	// NULL propagation through both functions.
	if v, _ := BindZstdCompress(nil); v != nil {
		t.Fatal("expected null compress")
	}
	if v, _ := BindZstdDecompressToBytes(nil); v != nil {
		t.Fatal("expected null decompress")
	}
	if v, _ := BindZstdDecompressToString(nil); v != nil {
		t.Fatal("expected null decompress->string")
	}
	if v, _ := BindZstdDecompressToBytes(value.BytesValue("zstd-frame"), nil); v != nil {
		t.Fatal("expected null size-limit -> null result")
	}

	// Missing args -> error.
	if _, err := BindZstdCompress(); err == nil {
		t.Fatal("expected arg count error")
	}
	if _, err := BindZstdDecompressToBytes(); err == nil {
		t.Fatal("expected arg count error")
	}

	// Invalid frame -> error from decompress.
	if _, err := BindZstdDecompressToBytes(value.BytesValue("not-a-zstd-frame")); err == nil {
		t.Fatal("expected decompress error")
	}
}

func TestBindGetNextSequenceValue(t *testing.T) {
	t.Parallel()

	// Use a unique name so the test does not interact with other tests.
	name := value.StringValue("test-seq-1")
	first, err := BindGetNextSequenceValue(name)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BindGetNextSequenceValue(name)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, second) != mustInt64(t, first)+1 {
		t.Fatal("sequence did not increment")
	}

	state, err := BindGetInternalSequenceState(name)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, state) != mustInt64(t, second) {
		t.Fatal("internal state should match last value")
	}

	if v, _ := BindGetNextSequenceValue(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindGetInternalSequenceState(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindGetNextSequenceValue(); err == nil {
		t.Fatal("expected arg count error")
	}
	if _, err := BindGetInternalSequenceState(); err == nil {
		t.Fatal("expected arg count error")
	}
}

// TestBindSequenceFunctionsConcurrent locks in the sequenceCounters
// synchronisation: GET_NEXT_SEQUENCE_VALUE and
// GET_INTERNAL_SEQUENCE_STATE are invoked as SQL function callbacks
// from concurrently running statements (parallel spec runs hit this),
// so the shared counter map must tolerate concurrent readers and
// writers. Run under -race this test reproduced the original crash
// deterministically before the mutex was added. It also checks the
// counter never skips or repeats: N goroutines × M increments on one
// sequence must end exactly at N*M.
func TestBindSequenceFunctionsConcurrent(t *testing.T) {
	t.Parallel()

	const (
		goroutines = 8
		increments = 200
	)
	name := value.StringValue("test-seq-concurrent")

	// The counter map is package-level and survives across -count=N
	// reruns in one process, so assert the delta, not the absolute.
	before, err := BindGetInternalSequenceState(name)
	if err != nil {
		t.Fatal(err)
	}
	start := mustInt64(t, before)

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < increments; i++ {
				if _, err := BindGetNextSequenceValue(name); err != nil {
					t.Error(err)
					return
				}
				if _, err := BindGetInternalSequenceState(name); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()

	state, err := BindGetInternalSequenceState(name)
	if err != nil {
		t.Fatal(err)
	}
	if got := mustInt64(t, state) - start; got != goroutines*increments {
		t.Fatalf("counter lost updates: got %d, want %d", got, goroutines*increments)
	}
}
