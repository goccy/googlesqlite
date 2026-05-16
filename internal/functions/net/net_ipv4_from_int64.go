package net

import (
	"encoding/binary"
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func NET_IPV4_FROM_INT64(v int64) (value.Value, error) {
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, uint32(v))
	return value.BytesValue(ip), nil
}

func BindNetIpv4FromInt64(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.IPV4_FROM_INT64: invalid number of arguments: got %d, want 1", len(args))
	}
	v, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	return NET_IPV4_FROM_INT64(v)
}
