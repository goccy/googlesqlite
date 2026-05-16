package bit

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

// BIT_CAST_TO_INT32 reinterprets the low 32 bits of v as a signed
// INT32. Values that fall outside the INT32 / UINT32 range raise an
// error so the contract matches the BigQuery semantics.
func BIT_CAST_TO_INT32(v value.Value) (value.Value, error) {
	x, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	if x < -1<<31 || x > 1<<32-1 {
		return nil, fmt.Errorf("BIT_CAST_TO_INT32: %d out of [INT32_MIN, UINT32_MAX]", x)
	}
	return value.IntValue(int64(int32(uint32(x)))), nil
}
