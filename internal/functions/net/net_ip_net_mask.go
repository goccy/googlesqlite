package net

import (
	"fmt"
	"net"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func NET_IP_NET_MASK(output, prefix int64) (value.Value, error) {
	prefixInt, err := helper.SafeInt(prefix)
	if err != nil {
		return nil, err
	}
	outputInt, err := helper.SafeInt(output * 8)
	if err != nil {
		return nil, err
	}
	result := net.CIDRMask(prefixInt, outputInt)
	if output != 4 && output != 16 {
		return nil, fmt.Errorf("NET.IP_NET_MASK: the first argument must be either 4 or 16")
	}
	if prefix < 0 || prefix > output*8 {
		return nil, fmt.Errorf("NET.IP_NET_MASK: the second argument must be in the range from 0 to %d", output*8)
	}
	return value.BytesValue(result), nil
}

var BindNetIpNetMask = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	output, err := a.ToInt64()
	if err != nil {
		return nil, err
	}
	prefix, err := b.ToInt64()
	if err != nil {
		return nil, err
	}
	return NET_IP_NET_MASK(output, prefix)
})
