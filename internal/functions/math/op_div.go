package math

import (
	"github.com/goccy/googlesqlite/internal/value"
)

func OP_DIV(a, b value.Value) (value.Value, error) {
	return a.Div(b)
}
