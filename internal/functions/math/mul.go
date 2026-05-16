package math

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func MUL(a, b value.Value) (value.Value, error) {
	return a.Mul(b)
}
