package string

import (
	"encoding/hex"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func FROM_HEX(v string) (value.Value, error) {
	if len(v)%2 != 0 {
		v = "0" + v
	}
	b, err := hex.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

var BindFromHex = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return FROM_HEX(v)
})
