package string

import (
	"encoding/hex"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TO_HEX(v []byte) (value.Value, error) {
	return value.StringValue(hex.EncodeToString(v)), nil
}

var BindToHex = helper.Scalar1(func(a value.Value) (value.Value, error) {
	b, err := a.ToBytes()
	if err != nil {
		return nil, err
	}
	return TO_HEX(b)
})
