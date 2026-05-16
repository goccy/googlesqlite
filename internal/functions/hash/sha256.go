package hash

import (
	"crypto/sha256"

	"github.com/goccy/googlesqlite/internal/value"
)

// SHA256 returns the SHA-256 digest of v as BYTES (raw 32 bytes).
func SHA256(v []byte) (value.Value, error) {
	sum := sha256.Sum256(v)
	return value.BytesValue(sum[:]), nil
}

func SHA256Bind(v value.Value) (value.Value, error) {
	b, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	return SHA256(b)
}
