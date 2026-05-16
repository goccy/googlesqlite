package bit

import "github.com/goccy/googlesqlite/internal/value"

// BIT_CAST_TO_UINT64 reinterprets the 64-bit pattern of v as a UINT64.
// The analyzer presents UINT64 values to this layer as int64 with the
// raw bit pattern, so the body is a no-op pass-through.
func BIT_CAST_TO_UINT64(v value.Value) (value.Value, error) {
	x, err := v.ToInt64()
	if err != nil {
		return nil, err
	}
	return value.IntValue(x), nil
}
