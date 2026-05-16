package hash

import (
	"crypto/sha512"

	"github.com/goccy/googlesqlite/internal/value"
)

// SHA512 returns the SHA-512 digest of v as BYTES (raw 64 bytes).
func SHA512(v []byte) (value.Value, error) {
	sum := sha512.Sum512(v)
	return value.BytesValue(sum[:]), nil
}

func SHA512Bind(v value.Value) (value.Value, error) {
	b, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	return SHA512(b)
}
