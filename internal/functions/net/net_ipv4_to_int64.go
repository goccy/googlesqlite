package net

import (
	"encoding/binary"
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func NET_IPV4_TO_INT64(v []byte) (value.Value, error) {
	if len(v) != 4 {
		return nil, fmt.Errorf("NET.IPV4_TO_INT64: length of bytes array must be 4")
	}
	return value.IntValue(binary.BigEndian.Uint32(v)), nil
}

func BindNetIpv4ToInt64(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.IPV4_TO_INT64: invalid number of arguments: got %d, want 1", len(args))
	}
	v, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	return NET_IPV4_TO_INT64(v)
}
