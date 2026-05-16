package net

import (
	"fmt"
	"net/netip"

	"github.com/goccy/googlesqlite/internal/value"
)

func NET_IP_TO_STRING(v []byte) (value.Value, error) {
	ip, ok := netip.AddrFromSlice(v)
	if !ok {
		return nil, fmt.Errorf("NET.IP_TO_STRING: invalid byte array")
	}
	return value.StringValue(ip.String()), nil
}

func BindNetIpToString(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.IP_TO_STRING: invalid number of arguments: got %d, want 1", len(args))
	}
	v, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	return NET_IP_TO_STRING(v)
}
