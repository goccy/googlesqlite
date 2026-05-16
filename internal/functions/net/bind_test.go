// Unit tests for the Bind* surface of the net package.
// Expected outputs follow the BigQuery NET functions reference
// (docs/third_party/googlesql-docs/net_functions.md).
package net_test

import (
	"bytes"
	"testing"

	netfn "github.com/goccy/googlesqlite/internal/functions/net"
	"github.com/goccy/googlesqlite/internal/value"
)

// ------------------------------------------------------------------
// Arity matrix
// ------------------------------------------------------------------

func TestNetBind_Arity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		bind func(...value.Value) (value.Value, error)
	}{
		{"NET.HOST", netfn.BindNetHost},
		{"NET.IP_FROM_STRING", netfn.BindNetIpFromString},
		{"NET.SAFE_IP_FROM_STRING", netfn.BindNetSafeIpFromString},
		{"NET.IP_TO_STRING", netfn.BindNetIpToString},
		{"NET.IPV4_FROM_INT64", netfn.BindNetIpv4FromInt64},
		{"NET.IPV4_TO_INT64", netfn.BindNetIpv4ToInt64},
		{"NET.PUBLIC_SUFFIX", netfn.BindNetPublicSuffix},
		{"NET.REG_DOMAIN", netfn.BindNetRegDomain},
	}
	for _, c := range cases {
		if _, err := c.bind(); err == nil {
			t.Errorf("%s(): expected arity error on zero args", c.name)
		}
		if _, err := c.bind(value.StringValue("x"), value.StringValue("y")); err == nil {
			t.Errorf("%s(...): expected arity error on too-many args", c.name)
		}
	}
	if _, err := netfn.BindNetIpNetMask(value.IntValue(4)); err == nil {
		t.Errorf("NET.IP_NET_MASK(4): expected arity error on <2 args")
	}
	if _, err := netfn.BindNetIpTrunc(value.BytesValue([]byte{0, 0, 0, 0})); err == nil {
		t.Errorf("NET.IP_TRUNC: expected arity error on <2 args")
	}
}

// ------------------------------------------------------------------
// NET.HOST
// ------------------------------------------------------------------

func TestNetHost(t *testing.T) {
	t.Parallel()

	// Per BQ docs: NET.HOST("https://www.example.com/foo?bar") -> "www.example.com"
	got, err := netfn.BindNetHost(value.StringValue("https://www.example.com/foo?bar"))
	if err != nil {
		t.Fatalf("NET.HOST: %v", err)
	}
	if got != value.StringValue("www.example.com") {
		t.Errorf("HOST = %v; want www.example.com", got)
	}
	// Bare hostname (no scheme).
	got, err = netfn.BindNetHost(value.StringValue("foo.example.com"))
	if err != nil {
		t.Fatalf("HOST bare: %v", err)
	}
	if got != value.StringValue("foo.example.com") {
		t.Errorf("HOST bare = %v; want foo.example.com", got)
	}
	// IPv6 host gets brackets back.
	got, err = netfn.BindNetHost(value.StringValue("http://[2001:db8::1]/"))
	if err != nil {
		t.Fatalf("HOST v6: %v", err)
	}
	if got != value.StringValue("[2001:db8::1]") {
		t.Errorf("HOST v6 = %v; want [2001:db8::1]", got)
	}
	// NULL -> NULL.
	got, err = netfn.BindNetHost(value.Value(nil))
	if err != nil {
		t.Fatalf("HOST nil: %v", err)
	}
	if got != nil {
		t.Errorf("HOST nil = %v; want nil", got)
	}
}

// ------------------------------------------------------------------
// NET.IP_FROM_STRING / SAFE / IP_TO_STRING
// ------------------------------------------------------------------

func TestNetIpFromString(t *testing.T) {
	t.Parallel()

	got, err := netfn.BindNetIpFromString(value.StringValue("192.0.2.1"))
	if err != nil {
		t.Fatalf("IP_FROM_STRING: %v", err)
	}
	bv, ok := got.(value.BytesValue)
	if !ok {
		t.Fatalf("got %T; want BytesValue", got)
	}
	if !bytes.Equal(bv, []byte{192, 0, 2, 1}) {
		t.Errorf("bytes = %v; want [192 0 2 1]", []byte(bv))
	}
	// Invalid -> error.
	if _, err := netfn.BindNetIpFromString(value.StringValue("not-an-ip")); err == nil {
		t.Errorf("expected error for invalid input")
	}
	// NULL -> NULL.
	got, err = netfn.BindNetIpFromString(value.Value(nil))
	if err != nil {
		t.Fatalf("nil: %v", err)
	}
	if got != nil {
		t.Errorf("nil = %v; want nil", got)
	}
}

func TestNetSafeIpFromString(t *testing.T) {
	t.Parallel()

	// Valid: same as IP_FROM_STRING.
	got, err := netfn.BindNetSafeIpFromString(value.StringValue("::1"))
	if err != nil {
		t.Fatalf("SAFE_IP_FROM_STRING ipv6: %v", err)
	}
	if _, ok := got.(value.BytesValue); !ok {
		t.Fatalf("got %T; want BytesValue", got)
	}
	// Invalid -> nil (no error).
	got, err = netfn.BindNetSafeIpFromString(value.StringValue("not-an-ip"))
	if err != nil {
		t.Fatalf("SAFE invalid: %v", err)
	}
	if got != nil {
		t.Errorf("safe invalid = %v; want nil", got)
	}
	// NULL -> NULL.
	got, err = netfn.BindNetSafeIpFromString(value.Value(nil))
	if err != nil {
		t.Fatalf("safe nil: %v", err)
	}
	if got != nil {
		t.Errorf("safe nil = %v; want nil", got)
	}
}

func TestNetIpToString(t *testing.T) {
	t.Parallel()

	// IPv4 four bytes -> dotted quad.
	got, err := netfn.BindNetIpToString(value.BytesValue{192, 0, 2, 1})
	if err != nil {
		t.Fatalf("IP_TO_STRING v4: %v", err)
	}
	if got != value.StringValue("192.0.2.1") {
		t.Errorf("v4 = %v; want 192.0.2.1", got)
	}
	// IPv6 16 bytes -> short form.
	v6 := value.BytesValue{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	got, err = netfn.BindNetIpToString(v6)
	if err != nil {
		t.Fatalf("IP_TO_STRING v6: %v", err)
	}
	if got != value.StringValue("2001:db8::1") {
		t.Errorf("v6 = %v; want 2001:db8::1", got)
	}
	// Wrong length -> error.
	if _, err := netfn.BindNetIpToString(value.BytesValue{1, 2, 3}); err == nil {
		t.Errorf("expected error for 3-byte input")
	}
}

// ------------------------------------------------------------------
// NET.IPV4_FROM_INT64 / TO_INT64
// ------------------------------------------------------------------

func TestNetIpv4Int64Roundtrip(t *testing.T) {
	t.Parallel()

	// 1.2.3.4 -> 0x01020304 = 16909060.
	got, err := netfn.BindNetIpv4FromInt64(value.IntValue(16909060))
	if err != nil {
		t.Fatalf("FROM_INT64: %v", err)
	}
	bv, ok := got.(value.BytesValue)
	if !ok {
		t.Fatalf("got %T; want BytesValue", got)
	}
	if !bytes.Equal(bv, []byte{1, 2, 3, 4}) {
		t.Errorf("bytes = %v; want [1 2 3 4]", []byte(bv))
	}
	got, err = netfn.BindNetIpv4ToInt64(bv)
	if err != nil {
		t.Fatalf("TO_INT64: %v", err)
	}
	if got != value.IntValue(16909060) {
		t.Errorf("roundtrip = %v; want 16909060", got)
	}
	// TO_INT64 with non-4-byte slice -> error.
	if _, err := netfn.BindNetIpv4ToInt64(value.BytesValue{1, 2, 3}); err == nil {
		t.Errorf("expected error on 3-byte input")
	}
}

// ------------------------------------------------------------------
// NET.IP_NET_MASK
// ------------------------------------------------------------------

func TestNetIpNetMask(t *testing.T) {
	t.Parallel()

	// IP_NET_MASK(4, 24) -> 255.255.255.0
	got, err := netfn.BindNetIpNetMask(value.IntValue(4), value.IntValue(24))
	if err != nil {
		t.Fatalf("IP_NET_MASK(4, 24): %v", err)
	}
	bv := got.(value.BytesValue)
	if !bytes.Equal(bv, []byte{255, 255, 255, 0}) {
		t.Errorf("got %v; want [255 255 255 0]", []byte(bv))
	}
	// IP_NET_MASK(16, 0) -> all-zero 16 bytes.
	got, err = netfn.BindNetIpNetMask(value.IntValue(16), value.IntValue(0))
	if err != nil {
		t.Fatalf("IP_NET_MASK(16, 0): %v", err)
	}
	bv = got.(value.BytesValue)
	if !bytes.Equal(bv, make([]byte, 16)) {
		t.Errorf("got %v; want 16 zeros", []byte(bv))
	}
	// IP_NET_MASK(5, 24) -> error.
	if _, err := netfn.BindNetIpNetMask(value.IntValue(5), value.IntValue(24)); err == nil {
		t.Errorf("expected error for invalid output length")
	}
	// IP_NET_MASK(4, 99) -> error (prefix out of range).
	if _, err := netfn.BindNetIpNetMask(value.IntValue(4), value.IntValue(99)); err == nil {
		t.Errorf("expected error for prefix out of range")
	}
	// NULL inputs -> NULL.
	got, err = netfn.BindNetIpNetMask(nil, value.IntValue(0))
	if err != nil {
		t.Fatalf("NULL output: %v", err)
	}
	if got != nil {
		t.Errorf("NULL output = %v; want nil", got)
	}
}

// ------------------------------------------------------------------
// NET.IP_TRUNC
// ------------------------------------------------------------------

func TestNetIpTrunc(t *testing.T) {
	t.Parallel()

	// IP_TRUNC(1.2.3.4 bytes, 24) -> 1.2.3.0
	got, err := netfn.BindNetIpTrunc(value.BytesValue{1, 2, 3, 4}, value.IntValue(24))
	if err != nil {
		t.Fatalf("IP_TRUNC: %v", err)
	}
	bv := got.(value.BytesValue)
	if !bytes.Equal(bv, []byte{1, 2, 3, 0}) {
		t.Errorf("got %v; want [1 2 3 0]", []byte(bv))
	}
	// Invalid input length.
	if _, err := netfn.BindNetIpTrunc(value.BytesValue{1, 2, 3}, value.IntValue(0)); err == nil {
		t.Errorf("expected error on non-4/16 input")
	}
	// Length out of range.
	if _, err := netfn.BindNetIpTrunc(value.BytesValue{1, 2, 3, 4}, value.IntValue(99)); err == nil {
		t.Errorf("expected error on length > bits")
	}
}

// ------------------------------------------------------------------
// NET.PUBLIC_SUFFIX / NET.REG_DOMAIN
// ------------------------------------------------------------------

func TestNetPublicSuffixAndRegDomain(t *testing.T) {
	t.Parallel()

	// Per BQ NET docs:
	//  NET.PUBLIC_SUFFIX("http://www.example.com")  -> "com"
	//  NET.REG_DOMAIN("http://www.example.com")     -> "example.com"
	got, err := netfn.BindNetPublicSuffix(value.StringValue("http://www.example.com"))
	if err != nil {
		t.Fatalf("PUBLIC_SUFFIX: %v", err)
	}
	if got != value.StringValue("com") {
		t.Errorf("PUBLIC_SUFFIX = %v; want com", got)
	}
	got, err = netfn.BindNetRegDomain(value.StringValue("http://www.example.com"))
	if err != nil {
		t.Fatalf("REG_DOMAIN: %v", err)
	}
	if got != value.StringValue("example.com") {
		t.Errorf("REG_DOMAIN = %v; want example.com", got)
	}

	// Multi-label suffix from upstream examples.
	got, err = netfn.BindNetPublicSuffix(value.StringValue("http://foo.bar.co.uk"))
	if err != nil {
		t.Fatalf("PUBLIC_SUFFIX uk: %v", err)
	}
	if got != value.StringValue("co.uk") {
		t.Errorf("PUBLIC_SUFFIX uk = %v; want co.uk", got)
	}
	got, err = netfn.BindNetRegDomain(value.StringValue("http://foo.bar.co.uk"))
	if err != nil {
		t.Fatalf("REG_DOMAIN uk: %v", err)
	}
	if got != value.StringValue("bar.co.uk") {
		t.Errorf("REG_DOMAIN uk = %v; want bar.co.uk", got)
	}

	// Unparseable / no host -> nil on both helpers.
	got, err = netfn.BindNetPublicSuffix(value.StringValue("\x7f"))
	if err != nil {
		t.Fatalf("PUBLIC_SUFFIX garbage: %v", err)
	}
	if got != nil {
		t.Errorf("PUBLIC_SUFFIX garbage = %v; want nil", got)
	}
	got, err = netfn.BindNetRegDomain(value.StringValue("\x7f"))
	if err != nil {
		t.Fatalf("REG_DOMAIN garbage: %v", err)
	}
	if got != nil {
		t.Errorf("REG_DOMAIN garbage = %v; want nil", got)
	}
}
