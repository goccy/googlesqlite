package bit

import "github.com/goccy/googlesqlite/internal/value"

// BIT_AND returns the bitwise AND of two INT64 values.
func BIT_AND(a, b value.Value) (value.Value, error) {
	va, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	vb, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return value.IntValue(va & vb), nil
}
