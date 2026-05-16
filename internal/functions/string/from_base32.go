package string

import (
	"encoding/base32"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func FROM_BASE32(v string) (value.Value, error) {
	b, err := base32.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

var BindFromBase32 = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return FROM_BASE32(v)
})
