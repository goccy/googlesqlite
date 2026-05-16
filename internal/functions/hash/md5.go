//nolint:gosec
package hash

import (
	"crypto/md5"

	"github.com/goccy/googlesqlite/internal/value"
)

// MD5 returns the MD5 digest of v as BYTES. The 16-byte result is
// the raw digest, not a hex encoding; pair with TO_HEX for the
// printable form.
func MD5(v []byte) (value.Value, error) {
	sum := md5.Sum(v)
	return value.BytesValue(sum[:]), nil
}

func MD5Bind(v value.Value) (value.Value, error) {
	b, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	return MD5(b)
}
