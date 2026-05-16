package bit

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
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
	// Reinterpret the low 32 bits of x as a signed int32. x may be
	// negative, so mask first to obtain a value in [0, 4294967295]
	// before the checked uint32 conversion.
	masked := x & 0xFFFFFFFF
	u, err := helper.SafeUint32(masked)
	if err != nil {
		return nil, err
	}
	return value.IntValue(int64(int32(u))), nil
}
