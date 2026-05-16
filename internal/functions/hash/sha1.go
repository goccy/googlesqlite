//nolint:gosec
package hash

import (
	"crypto/sha1"

	"github.com/goccy/googlesqlite/internal/value"
)

// SHA1 returns the SHA-1 digest of v as BYTES (raw 20 bytes).
func SHA1(v []byte) (value.Value, error) {
	sum := sha1.Sum(v)
	return value.BytesValue(sum[:]), nil
}

func SHA1Bind(v value.Value) (value.Value, error) {
	b, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	return SHA1(b)
}
