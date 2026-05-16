package bit

import (
	"math/bits"

	"github.com/goccy/googlesqlite/internal/value"
)

// BIT_COUNT counts the set bits of an INT64 or BYTES input. Mirrors
// the GoogleSQL semantic: each byte of a BYTES input contributes its
// own popcount, summed; an INT64 value is interpreted as its 64-bit
// pattern.
func BIT_COUNT(v value.Value) (value.Value, error) {
	switch v.(type) {
	case value.BytesValue:
		b, err := v.ToBytes()
		if err != nil {
			return nil, err
		}
		var sum int64
		for _, vv := range b {
			sum += int64(bits.OnesCount8(vv))
		}
		return value.IntValue(sum), nil
	default:
		vv, err := v.ToInt64()
		if err != nil {
			return nil, err
		}
		return value.IntValue(bits.OnesCount64(uint64(vv))), nil
	}
}
