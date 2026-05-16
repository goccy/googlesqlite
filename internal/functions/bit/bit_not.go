package bit

import "github.com/goccy/googlesqlite/internal/value"

// BIT_NOT returns the bitwise complement of an INT64 value.
func BIT_NOT(a value.Value) (value.Value, error) {
	v, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	return value.IntValue(^v), nil
}
