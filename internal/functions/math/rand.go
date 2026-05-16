package math

import (
	"math/rand"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func RAND() (value.Value, error) {
	return value.FloatValue(rand.Float64()), nil
}

var BindRand = helper.ScalarN(func(_ ...value.Value) (value.Value, error) {
	return RAND()
})
