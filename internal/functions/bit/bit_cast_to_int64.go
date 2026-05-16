package bit

import "github.com/goccy/googlesqlite/internal/value"

// BIT_CAST_TO_INT64 reinterprets the 64-bit pattern of v (an INT64 or
// UINT64) as an INT64. Only the low 64 bits of the input matter; values
// outside [INT64_MIN, UINT64_MAX] are rejected by the analyzer before
// reaching this function.
func BIT_CAST_TO_INT64(v value.Value) (value.Value, error) {
	x, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	return value.IntValue(x), nil
}
