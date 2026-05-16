package net

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func NET_IP_FROM_STRING(v string) (value.Value, error) {
	ip, err := parseIP(v)
	if err != nil {
		return nil, fmt.Errorf("NET.IP_FROM_STRING: invalid ip address %v", v)
	}
	return value.BytesValue(ip), nil
}

var BindNetIpFromString = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return NET_IP_FROM_STRING(v)
})
