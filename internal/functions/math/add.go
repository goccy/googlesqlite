package math

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func ADD(a, b value.Value) (value.Value, error) {
	return a.Add(b)
}
