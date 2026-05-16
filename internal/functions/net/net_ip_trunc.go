package net

import (
	"fmt"
	"net"

	"github.com/goccy/googlesqlite/internal/value"
)

func NET_IP_TRUNC(v []byte, length int64) (value.Value, error) {
	if len(v) != 4 && len(v) != 16 {
		return nil, fmt.Errorf("NET.IP_TRUNC: length of the first argument must be either 4 or 16")
	}
	if length < 0 || int(length) > len(v)*8 {
		return nil, fmt.Errorf("NET.IP_TRUNC: length must be in the range from 0 to %d", len(v)*8)
	}
	ip := net.IP(v)
	mask := net.CIDRMask(int(length), len(v)*8)
	return value.BytesValue(ip.Mask(mask)), nil
}

func BindNetIpTrunc(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("NET.IP_TRUNC: invalid number of arguments: got %d, want 2", len(args))
	}
	v, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	length, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	return NET_IP_TRUNC(v, length)
}
