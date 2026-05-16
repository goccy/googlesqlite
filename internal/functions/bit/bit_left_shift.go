package bit

import "github.com/goccy/googlesqlite/internal/value"

// BIT_LEFT_SHIFT shifts a left by b bits.
func BIT_LEFT_SHIFT(a, b value.Value) (value.Value, error) {
	va, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	vb, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return value.IntValue(va << vb), nil
}
