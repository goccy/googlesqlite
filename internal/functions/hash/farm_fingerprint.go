package hash

import (
	"github.com/dgryski/go-farm"
	"github.com/goccy/googlesqlite/internal/value"
)

// FARM_FINGERPRINT returns the FarmHash 64-bit fingerprint of v as an
// INT64. The fingerprint is stable across runs.
func FARM_FINGERPRINT(v []byte) (value.Value, error) {
	return value.IntValue(farm.Fingerprint64(v)), nil
}

func FarmFingerprintBind(v value.Value) (value.Value, error) {
	b, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	return FARM_FINGERPRINT(b)
}
