package net

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func NET_PUBLIC_SUFFIX(v string) (value.Value, error) {
	parsed := parseURL(v)
	if parsed == nil {
		return nil, nil
	}
	host := parsed.Hostname()
	suffix, err := publicSuffix(host)
	if err != nil {
		return nil, fmt.Errorf("NET.PUBLIC_SUFFIX: invalid hostname %s", host)
	}
	if suffix == "" {
		return nil, nil
	}
	return value.StringValue(suffix), nil
}

func BindNetPublicSuffix(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.PUBLIC_SUFFIX: invalid number of arguments: got %d, want 1", len(args))
	}
	v, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return NET_PUBLIC_SUFFIX(v)
}
