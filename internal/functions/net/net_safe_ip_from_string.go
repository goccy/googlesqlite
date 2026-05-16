package net

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func NET_SAFE_IP_FROM_STRING(v string) (value.Value, error) {
	ip, err := parseIP(v)
	if err != nil {
		return nil, nil
	}
	return value.BytesValue(ip), nil
}

var BindNetSafeIpFromString = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return NET_SAFE_IP_FROM_STRING(v)
})
