package net

import (
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/value"
)

func NET_REG_DOMAIN(v string) (value.Value, error) {
	parsed := parseURL(v)
	if parsed == nil {
		return nil, nil
	}
	host := parsed.Hostname()
	splitHost := strings.Split(host, ".")
	suffix, err := publicSuffix(host)
	if err != nil {
		return nil, fmt.Errorf("NET.REG_DOMAIN: invalid hostname %s", host)
	}
	splitSuffix := strings.Split(suffix, ".")
	if host == "" || suffix == "" || len(splitHost) <= len(splitSuffix) {
		return nil, nil
	}
	return value.StringValue(strings.Join(splitHost[len(splitHost)-len(splitSuffix)-1:], ".")), nil
}

func BindNetRegDomain(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.REG_DOMAIN: invalid number of arguments: got %d, want 1", len(args))
	}
	v, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return NET_REG_DOMAIN(v)
}
