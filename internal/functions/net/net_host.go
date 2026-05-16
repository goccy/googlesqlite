package net

import (
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func NET_HOST(v string) (value.Value, error) {
	parsed := parseURL(v)
	if parsed == nil {
		return nil, nil
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		return nil, nil
	}
	if strings.HasPrefix(parsed.Host, "[") {
		return value.StringValue("[" + hostname + "]"), nil
	}
	return value.StringValue(hostname), nil
}

var BindNetHost = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return NET_HOST(v)
})
