package net

import (
	"encoding/binary"
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func NET_IPV4_FROM_INT64(v int64) (value.Value, error) {
	// GoogleSQL accepts an input in [-0x80000000, 0xFFFFFFFF]: the low
	// 32 bits hold the address and the upper bits must be a sign
	// extension of bit 31.
	if v < -0x80000000 || v > 0xFFFFFFFF {
		return nil, fmt.Errorf("NET.IPV4_FROM_INT64: %d out of range [-0x80000000, 0xFFFFFFFF]", v)
	}
	// Reinterpret the low 32 bits as an unsigned address; masking keeps
	// the value in [0, 0xFFFFFFFF] for the checked conversion.
	u32, err := helper.SafeUint32(v & 0xFFFFFFFF)
	if err != nil {
		return nil, err
	}
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, u32)
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
