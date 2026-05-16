package helper

import (
	"fmt"
	"math"
)

// SafeInt converts a 64-bit integer to int. It returns an error when v
// falls outside the 32-bit range, which is the minimum width the Go
// specification guarantees for int. Callers use the result as a slice
// index, length, count, or a date/time component; none of those can
// legitimately exceed that range, so an out-of-range value is a genuine
// error rather than a silently truncated conversion.
func SafeInt(v int64) (int, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("integer value %d is out of the supported range [%d, %d]", v, math.MinInt32, math.MaxInt32)
	}
	return int(v), nil
}

// SafeInt32 converts a 64-bit integer to int32, returning an error when
// v overflows the INT32 range.
func SafeInt32(v int64) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("integer value %d is out of the INT32 range [%d, %d]", v, math.MinInt32, math.MaxInt32)
	}
	return int32(v), nil
}

// SafeByte converts a 64-bit integer to byte, returning an error when v
// falls outside the [0, 255] range.
func SafeByte(v int64) (byte, error) {
	if v < 0 || v > math.MaxUint8 {
		return 0, fmt.Errorf("integer value %d is out of the BYTE range [0, 255]", v)
	}
	return byte(v), nil
}

// SafeUint32 converts a 64-bit integer to uint32, returning an error
// when v falls outside the [0, 4294967295] range.
func SafeUint32(v int64) (uint32, error) {
	if v < 0 || v > math.MaxUint32 {
		return 0, fmt.Errorf("integer value %d is out of the UINT32 range [0, 4294967295]", v)
	}
	return uint32(v), nil
}
