---
name: NET.IPV4_FROM_INT64
dialect: googlesql
category: functions/net
status: implemented
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netipv4_from_int64
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_ipv4_from_int64.yaml
---

# NET.IPV4_FROM_INT64

## Summary

(TBD — refine from the upstream reference below.)

## Signatures

(TBD)

## Behavior

(TBD)

## Examples

(TBD)

## Edge cases

(TBD)

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/net_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `NET.IPV4_FROM_INT64`

```
NET.IPV4_FROM_INT64(integer_value)
```

**Description**

Converts an IPv4 address from integer format to binary (BYTES) format in network
byte order. In the integer input, the least significant bit of the IP address is
stored in the least significant bit of the integer, regardless of host or client
architecture. For example, `1` means `0.0.0.1`, and `0x1FF` means `0.0.1.255`.

This function checks that either all the most significant 32 bits are 0, or all
the most significant 33 bits are 1 (sign-extended from a 32-bit integer).
In other words, the input should be in the range `[-0x80000000, 0xFFFFFFFF]`;
otherwise, this function throws an error.

This function doesn't support IPv6.

**Return Data Type**

BYTES

**Example**

```googlesql
SELECT x, x_hex, FORMAT("%T", NET.IPV4_FROM_INT64(x)) AS ipv4_from_int64
FROM (
  SELECT CAST(x_hex AS INT64) x, x_hex
  FROM UNNEST(["0x0", "0xABCDEF", "0xFFFFFFFF", "-0x1", "-0x2"]) AS x_hex
);

/*-----------------------------------------------+
 | x          | x_hex      | ipv4_from_int64     |
 +-----------------------------------------------+
 | 0          | 0x0        | b"\x00\x00\x00\x00" |
 | 11259375   | 0xABCDEF   | b"\x00\xab\xcd\xef" |
 | 4294967295 | 0xFFFFFFFF | b"\xff\xff\xff\xff" |
 | -1         | -0x1       | b"\xff\xff\xff\xff" |
 | -2         | -0x2       | b"\xff\xff\xff\xfe" |
 +-----------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
