package math

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func SUB(a, b value.Value) (value.Value, error) {
	return a.Sub(b)
}
