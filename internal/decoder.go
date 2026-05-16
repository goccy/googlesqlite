package internal

import "github.com/goccy/googlesqlite/internal/value"

// DecodeValue is a thin shim that forwards to the canonical
// implementation in internal/value. The decode logic lives in the
// leaf value package so per-category function sub-packages and the
// window infrastructure can use it without going through internal.
func DecodeValue(v any) (value.Value, error) {
	return value.DecodeValue(v)
}
