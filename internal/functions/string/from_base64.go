package string

import (
	"encoding/base64"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func FROM_BASE64(v string) (value.Value, error) {
	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

var BindFromBase64 = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return FROM_BASE64(v)
})
