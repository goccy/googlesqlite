package hll

import (
	"github.com/DataDog/go-hll"

	"github.com/goccy/googlesqlite/internal/value"
)

func init() {
	_ = hll.Defaults(hll.Settings{
		Log2m:             15,
		Regwidth:          8,
		ExplicitThreshold: hll.AutoExplicitThreshold,
		SparseEnabled:     true,
	})
}

// HLL_COUNT_EXTRACT is registered as a scalar function (not an
// aggregate) — it accepts a serialized sketch and returns its
// cardinality.
func HLL_COUNT_EXTRACT(sketch []byte) (value.Value, error) {
	h, err := hll.FromBytes(sketch)
	if err != nil {
		return nil, err
	}
	return value.IntValue(h.Cardinality()), nil
}
