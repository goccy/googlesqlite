package bit

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// BIT_CAST_TO_UINT32 reinterprets the low 32 bits of v as an unsigned
// UINT32 (carried as INT64 on the SQLite side).
func BIT_CAST_TO_UINT32(v value.Value) (value.Value, error) {
	x, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	if x < -1<<31 || x > 1<<32-1 {
		return nil, fmt.Errorf("BIT_CAST_TO_UINT32: %d out of [INT32_MIN, UINT32_MAX]", x)
	}
	return value.IntValue(int64(uint32(x))), nil
}
