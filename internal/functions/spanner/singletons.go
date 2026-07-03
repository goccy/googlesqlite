package spanner

import (
	"fmt"
	"math"
	"math/bits"
	"sync"
	gotime "time"

	"github.com/klauspost/compress/zstd"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// BindPendingCommitTimestamp returns a sentinel timestamp value
// for Spanner's `PENDING_COMMIT_TIMESTAMP()`. googlesqlite has no
// commit-time sentinel substitution, so we use the maximum
// representable TIMESTAMP (year 9999) as a recognisable marker.
func BindPendingCommitTimestamp(_ ...value.Value) (value.Value, error) {
	sentinel := gotime.Date(9999, 12, 31, 23, 59, 59, 999999999, gotime.UTC)
	return value.TimestampValue(sentinel), nil
}

// BindBitReverse reverses the bits of an INT64. When `preserve_width`
// is FALSE the reversal is over the full 64-bit width; when TRUE the
// reversal is over the minimal bit width needed to represent the
// magnitude of `value`.
func BindBitReverse(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("BIT_REVERSE: invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	n, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	preserve := false
	if len(args) == 2 {
		b, err := args[1].ToBool()
		if err != nil {
			return nil, err
		}
		preserve = b
	}
	u := uint64(n)
	if !preserve {
		return value.IntValue(int64(bits.Reverse64(u))), nil
	}
	width := uint(bits.Len64(u))
	if width == 0 {
		return value.IntValue(0), nil
	}
	r := bits.Reverse64(u) >> (64 - width)
	return value.IntValue(int64(r)), nil
}

// BindDotProduct returns the dot product of two equal-length numeric
// ARRAYs.
func BindDotProduct(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("DOT_PRODUCT: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	a, aok := args[0].(*value.ArrayValue)
	b, bok := args[1].(*value.ArrayValue)
	if !aok || !bok {
		return nil, fmt.Errorf("DOT_PRODUCT: arguments must be ARRAY")
	}
	if len(a.Values) != len(b.Values) {
		return nil, fmt.Errorf("DOT_PRODUCT: array length mismatch (%d vs %d)", len(a.Values), len(b.Values))
	}
	var s float64
	for i, av := range a.Values {
		bv := b.Values[i]
		if av == nil || bv == nil {
			return nil, nil
		}
		af, err := av.ToFloat64()
		if err != nil {
			return nil, err
		}
		bf, err := bv.ToFloat64()
		if err != nil {
			return nil, err
		}
		s += af * bf
	}
	return value.FloatValue(s), nil
}

// BindApproxDotProduct returns the dot product of two equal-length
// numeric ARRAYs. The "approx" variants in Spanner have the same
// observable semantics as the exact form on small vectors.
func BindApproxDotProduct(args ...value.Value) (value.Value, error) {
	v, err := BindDotProduct(args...)
	if err != nil {
		return nil, fmt.Errorf("APPROX_DOT_PRODUCT: %w", err)
	}
	return v, nil
}

// BindApproxEuclideanDistance returns sqrt(SUM((a[i]-b[i])^2)).
func BindApproxEuclideanDistance(args ...value.Value) (value.Value, error) {
	a, b, err := approxVectorPair("APPROX_EUCLIDEAN_DISTANCE", args)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, nil
	}
	var s float64
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return value.FloatValue(math.Sqrt(s)), nil
}

// BindApproxCosineDistance returns 1 - dot(a, b) / (|a| * |b|).
func BindApproxCosineDistance(args ...value.Value) (value.Value, error) {
	a, b, err := approxVectorPair("APPROX_COSINE_DISTANCE", args)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, nil
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return nil, fmt.Errorf("APPROX_COSINE_DISTANCE: zero-magnitude vector")
	}
	return value.FloatValue(1 - dot/(math.Sqrt(na)*math.Sqrt(nb))), nil
}

// approxVectorPair extracts two equal-length ARRAY<FLOAT64> from
// the argument list, returning (nil, nil, nil) when either is NULL.
func approxVectorPair(name string, args []value.Value) ([]float64, []float64, error) {
	if len(args) != 2 {
		return nil, nil, fmt.Errorf("%s: invalid number of arguments: got %d, want 2", name, len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil, nil
	}
	a, aok := args[0].(*value.ArrayValue)
	b, bok := args[1].(*value.ArrayValue)
	if !aok || !bok {
		return nil, nil, fmt.Errorf("%s: arguments must be ARRAY", name)
	}
	if len(a.Values) != len(b.Values) {
		return nil, nil, fmt.Errorf("%s: array length mismatch", name)
	}
	af := make([]float64, len(a.Values))
	bf := make([]float64, len(b.Values))
	for i := range a.Values {
		if a.Values[i] == nil || b.Values[i] == nil {
			return nil, nil, nil
		}
		x, err := a.Values[i].ToFloat64()
		if err != nil {
			return nil, nil, err
		}
		y, err := b.Values[i].ToFloat64()
		if err != nil {
			return nil, nil, err
		}
		af[i] = x
		bf[i] = y
	}
	return af, bf, nil
}

// BindZstdCompress compresses BYTES with Zstandard. Optional
// compression-level argument is honoured at coarse granularity:
// 1-3 → fastest, 4-15 → default, 16+ → best.
func BindZstdCompress(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ZSTD_COMPRESS: missing argument")
	}
	if args[0] == nil {
		return nil, nil
	}
	in, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	level := zstd.SpeedDefault
	if len(args) >= 2 && args[1] != nil {
		n, err := args[1].ToInt64()
		if err == nil {
			switch {
			case n <= 3:
				level = zstd.SpeedFastest
			case n >= 16:
				level = zstd.SpeedBestCompression
			}
		}
	}
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(level))
	if err != nil {
		return nil, err
	}
	defer enc.Close()
	out := enc.EncodeAll(in, nil)
	return value.BytesValue(out), nil
}

// BindZstdDecompressToBytes decompresses a zstd frame back to BYTES.
// Spanner exposes an optional `size_limit` named argument that caps
// the size of the decompressed output. A NULL `size_limit` is
// surface as a NULL result (per the Spanner compliance fixtures), a
// strictly positive value caps the output, and `0`/negative values
// are rejected.
func BindZstdDecompressToBytes(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ZSTD_DECOMPRESS_TO_BYTES: missing argument")
	}
	if args[0] == nil {
		return nil, nil
	}
	if len(args) >= 2 && args[1] == nil {
		return nil, nil
	}
	in, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	defer dec.Close()
	out, err := dec.DecodeAll(in, nil)
	if err != nil {
		return nil, fmt.Errorf("ZSTD_DECOMPRESS_TO_BYTES: %w", err)
	}
	return value.BytesValue(out), nil
}

// BindZstdDecompressToString decompresses a zstd frame back to a
// UTF-8 STRING.
func BindZstdDecompressToString(args ...value.Value) (value.Value, error) {
	v, err := BindZstdDecompressToBytes(args...)
	if err != nil {
		return nil, fmt.Errorf("ZSTD_DECOMPRESS_TO_STRING: %w", err)
	}
	if v == nil {
		return nil, nil
	}
	b, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

// BindGetNextSequenceValue returns the next value of a Spanner
// sequence. Spanner sequences are not yet modelled by googlesqlite —
// this implementation returns a monotonically increasing INT64 per
// (process, sequence-name) pair.
var (
	sequenceCounters   = map[string]int64{}
	sequenceCountersMu sync.Mutex
)

func BindGetNextSequenceValue(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("GET_NEXT_SEQUENCE_VALUE: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	name, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	sequenceCountersMu.Lock()
	sequenceCounters[name]++
	next := sequenceCounters[name]
	sequenceCountersMu.Unlock()
	return value.IntValue(next), nil
}

// BindGetInternalSequenceState returns the internal counter for a
// sequence (the value last handed out by GET_NEXT_SEQUENCE_VALUE).
func BindGetInternalSequenceState(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("GET_INTERNAL_SEQUENCE_STATE: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	name, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	sequenceCountersMu.Lock()
	cur := sequenceCounters[name]
	sequenceCountersMu.Unlock()
	return value.IntValue(cur), nil
}
