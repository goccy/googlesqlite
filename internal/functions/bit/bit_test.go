package bit

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

func toInt64(t *testing.T, v value.Value) int64 {
	t.Helper()
	if v == nil {
		t.Fatalf("nil value, expected non-nil")
	}
	n, err := v.ToInt64()
	if err != nil {
		t.Fatalf("ToInt64: %v", err)
	}
	return n
}

// TestBitAnd covers two AND patterns; values come from the
// GoogleSQL bitwise-AND examples (bigquery docs operators page).
func TestBitAnd(t *testing.T) {
	cases := []struct{ a, b, want int64 }{
		{0xF0, 0x0F, 0x00},
		{0xFF, 0x0F, 0x0F},
		{-1, 0x10, 0x10},
	}
	for _, tc := range cases {
		got, err := BIT_AND(value.IntValue(tc.a), value.IntValue(tc.b))
		if err != nil {
			t.Fatalf("BIT_AND(%d,%d): %v", tc.a, tc.b, err)
		}
		if v := toInt64(t, got); v != tc.want {
			t.Errorf("BIT_AND(%d,%d) = %d, want %d", tc.a, tc.b, v, tc.want)
		}
	}
}

func TestBitOr(t *testing.T) {
	cases := []struct{ a, b, want int64 }{
		{0xF0, 0x0F, 0xFF},
		{0, 0, 0},
		{0x01, 0x02, 0x03},
	}
	for _, tc := range cases {
		got, err := BIT_OR(value.IntValue(tc.a), value.IntValue(tc.b))
		if err != nil {
			t.Fatalf("BIT_OR: %v", err)
		}
		if v := toInt64(t, got); v != tc.want {
			t.Errorf("BIT_OR(%d,%d) = %d, want %d", tc.a, tc.b, v, tc.want)
		}
	}
}

func TestBitXor(t *testing.T) {
	cases := []struct{ a, b, want int64 }{
		{0xF0, 0xFF, 0x0F},
		{0x05, 0x05, 0x00},
		{0x00, 0x01, 0x01},
	}
	for _, tc := range cases {
		got, err := BIT_XOR(value.IntValue(tc.a), value.IntValue(tc.b))
		if err != nil {
			t.Fatalf("BIT_XOR: %v", err)
		}
		if v := toInt64(t, got); v != tc.want {
			t.Errorf("BIT_XOR(%d,%d) = %d, want %d", tc.a, tc.b, v, tc.want)
		}
	}
}

func TestBitNot(t *testing.T) {
	got, err := BIT_NOT(value.IntValue(0))
	if err != nil {
		t.Fatalf("BIT_NOT: %v", err)
	}
	if v := toInt64(t, got); v != ^int64(0) {
		t.Errorf("BIT_NOT(0) = %d, want %d", v, ^int64(0))
	}

	got, err = BIT_NOT(value.IntValue(-1))
	if err != nil {
		t.Fatalf("BIT_NOT: %v", err)
	}
	if v := toInt64(t, got); v != 0 {
		t.Errorf("BIT_NOT(-1) = %d, want 0", v)
	}
}

func TestBitLeftShift(t *testing.T) {
	got, err := BIT_LEFT_SHIFT(value.IntValue(1), value.IntValue(4))
	if err != nil {
		t.Fatalf("BIT_LEFT_SHIFT: %v", err)
	}
	if v := toInt64(t, got); v != 16 {
		t.Errorf("1 << 4 = %d, want 16", v)
	}
	got, _ = BIT_LEFT_SHIFT(value.IntValue(0), value.IntValue(10))
	if v := toInt64(t, got); v != 0 {
		t.Errorf("0 << 10 = %d, want 0", v)
	}
}

func TestBitRightShift(t *testing.T) {
	got, err := BIT_RIGHT_SHIFT(value.IntValue(16), value.IntValue(4))
	if err != nil {
		t.Fatalf("BIT_RIGHT_SHIFT: %v", err)
	}
	if v := toInt64(t, got); v != 1 {
		t.Errorf("16 >> 4 = %d, want 1", v)
	}
}

// TestBitCastToInt32 mirrors the BigQuery BIT_CAST_TO_INT32 docs:
// values within the UINT32 range round-trip; -1 is rejected only when
// it falls outside the documented bounds.
func TestBitCastToInt32(t *testing.T) {
	got, err := BIT_CAST_TO_INT32(value.IntValue(0xFFFFFFFF))
	if err != nil {
		t.Fatalf("BIT_CAST_TO_INT32: %v", err)
	}
	if v := toInt64(t, got); v != -1 {
		t.Errorf("BIT_CAST_TO_INT32(0xFFFFFFFF) = %d, want -1", v)
	}

	got, err = BIT_CAST_TO_INT32(value.IntValue(1))
	if err != nil {
		t.Fatalf("BIT_CAST_TO_INT32: %v", err)
	}
	if v := toInt64(t, got); v != 1 {
		t.Errorf("BIT_CAST_TO_INT32(1) = %d, want 1", v)
	}

	if _, err := BIT_CAST_TO_INT32(value.IntValue(int64(1) << 33)); err == nil {
		t.Errorf("BIT_CAST_TO_INT32 should reject out-of-range")
	}
}

func TestBitCastToInt64(t *testing.T) {
	got, err := BIT_CAST_TO_INT64(value.IntValue(42))
	if err != nil {
		t.Fatalf("BIT_CAST_TO_INT64: %v", err)
	}
	if v := toInt64(t, got); v != 42 {
		t.Errorf("BIT_CAST_TO_INT64(42) = %d, want 42", v)
	}
}

func TestBitCastToUint32(t *testing.T) {
	got, err := BIT_CAST_TO_UINT32(value.IntValue(-1))
	if err != nil {
		t.Fatalf("BIT_CAST_TO_UINT32: %v", err)
	}
	if v := toInt64(t, got); v != 0xFFFFFFFF {
		t.Errorf("BIT_CAST_TO_UINT32(-1) = %d, want 0xFFFFFFFF", v)
	}
	if _, err := BIT_CAST_TO_UINT32(value.IntValue(int64(1) << 33)); err == nil {
		t.Errorf("BIT_CAST_TO_UINT32 should reject out-of-range")
	}
}

func TestBitCastToUint64(t *testing.T) {
	got, err := BIT_CAST_TO_UINT64(value.IntValue(123))
	if err != nil {
		t.Fatalf("BIT_CAST_TO_UINT64: %v", err)
	}
	if v := toInt64(t, got); v != 123 {
		t.Errorf("BIT_CAST_TO_UINT64(123) = %d, want 123", v)
	}
}

// TestBitCountInt64: GoogleSQL BIT_COUNT(INT64) returns the number of
// set bits in the two's-complement representation.
func TestBitCountInt64(t *testing.T) {
	cases := []struct {
		in   int64
		want int64
	}{
		{0, 0},
		{0x0F, 4},
		{0xFF, 8},
		{-1, 64},
	}
	for _, tc := range cases {
		got, err := BIT_COUNT(value.IntValue(tc.in))
		if err != nil {
			t.Fatalf("BIT_COUNT(%d): %v", tc.in, err)
		}
		if v := toInt64(t, got); v != tc.want {
			t.Errorf("BIT_COUNT(%d) = %d, want %d", tc.in, v, tc.want)
		}
	}
}

// newBadInt builds a value that fails ToInt64 (and ToBytes via
// ArrayValue), letting us cover error-return paths in the bit kernels
// without modifying production code.
func newBadInt() value.Value {
	return &value.ArrayValue{Values: []value.Value{value.IntValue(1)}}
}

func TestBitErrorPaths(t *testing.T) {
	bad := newBadInt()
	if _, err := BIT_AND(bad, value.IntValue(1)); err == nil {
		t.Error("BIT_AND first arg error not propagated")
	}
	if _, err := BIT_AND(value.IntValue(1), bad); err == nil {
		t.Error("BIT_AND second arg error not propagated")
	}
	if _, err := BIT_OR(bad, value.IntValue(1)); err == nil {
		t.Error("BIT_OR first arg error not propagated")
	}
	if _, err := BIT_OR(value.IntValue(1), bad); err == nil {
		t.Error("BIT_OR second arg error not propagated")
	}
	if _, err := BIT_XOR(bad, value.IntValue(1)); err == nil {
		t.Error("BIT_XOR first arg error not propagated")
	}
	if _, err := BIT_XOR(value.IntValue(1), bad); err == nil {
		t.Error("BIT_XOR second arg error not propagated")
	}
	if _, err := BIT_NOT(bad); err == nil {
		t.Error("BIT_NOT error not propagated")
	}
	if _, err := BIT_LEFT_SHIFT(bad, value.IntValue(1)); err == nil {
		t.Error("BIT_LEFT_SHIFT first arg error not propagated")
	}
	if _, err := BIT_LEFT_SHIFT(value.IntValue(1), bad); err == nil {
		t.Error("BIT_LEFT_SHIFT second arg error not propagated")
	}
	if _, err := BIT_RIGHT_SHIFT(bad, value.IntValue(1)); err == nil {
		t.Error("BIT_RIGHT_SHIFT first arg error not propagated")
	}
	if _, err := BIT_RIGHT_SHIFT(value.IntValue(1), bad); err == nil {
		t.Error("BIT_RIGHT_SHIFT second arg error not propagated")
	}
	if _, err := BIT_CAST_TO_INT32(bad); err == nil {
		t.Error("BIT_CAST_TO_INT32 error not propagated")
	}
	if _, err := BIT_CAST_TO_INT64(bad); err == nil {
		t.Error("BIT_CAST_TO_INT64 error not propagated")
	}
	if _, err := BIT_CAST_TO_UINT32(bad); err == nil {
		t.Error("BIT_CAST_TO_UINT32 error not propagated")
	}
	if _, err := BIT_CAST_TO_UINT64(bad); err == nil {
		t.Error("BIT_CAST_TO_UINT64 error not propagated")
	}
	if _, err := BIT_COUNT(bad); err == nil {
		t.Error("BIT_COUNT error not propagated")
	}
}

func TestBitCountBytes(t *testing.T) {
	got, err := BIT_COUNT(value.BytesValue([]byte{0x0F, 0xF0, 0xFF}))
	if err != nil {
		t.Fatalf("BIT_COUNT(bytes): %v", err)
	}
	if v := toInt64(t, got); v != 16 {
		t.Errorf("BIT_COUNT([0x0F,0xF0,0xFF]) = %d, want 16", v)
	}

	got, err = BIT_COUNT(value.BytesValue([]byte{}))
	if err != nil {
		t.Fatalf("BIT_COUNT([]): %v", err)
	}
	if v := toInt64(t, got); v != 0 {
		t.Errorf("BIT_COUNT([]) = %d, want 0", v)
	}
}
