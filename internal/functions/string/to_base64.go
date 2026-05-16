package string

import (
	"encoding/base64"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TO_BASE64(v []byte) (value.Value, error) {
	return value.StringValue(base64.StdEncoding.EncodeToString(v)), nil
}

var BindToBase64 = helper.Scalar1(func(a value.Value) (value.Value, error) {
	b, err := a.ToBytes()
	if err != nil {
		return nil, err
	}
	return TO_BASE64(b)
})
