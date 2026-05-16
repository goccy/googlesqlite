---
name: NET.IP_TRUNC
dialect: googlesql
category: functions/net
status: implemented
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netip_trunc
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_ip_trunc.yaml
---

# NET.IP_TRUNC

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

## `NET.IP_TRUNC`

```
NET.IP_TRUNC(addr_bin, prefix_length)
```

**Description**
Takes `addr_bin`, an IPv4 or IPv6 address in binary (BYTES) format in network
byte order, and returns a subnet address in the same format. The result has the
same length as `addr_bin`, where the first `prefix_length` bits are equal to
those in `addr_bin` and the remaining bits are 0.

This function throws an error if `LENGTH(addr_bin)` isn't 4 or 16, or if
`prefix_len` is negative or greater than `LENGTH(addr_bin) * 8`.

**Return Data Type**

BYTES

**Example**

```googlesql
SELECT
  FORMAT("%T", x) as addr_bin, prefix_length,
  FORMAT("%T", NET.IP_TRUNC(x, prefix_length)) AS ip_trunc
FROM UNNEST([
  STRUCT(b"\xAA\xBB\xCC\xDD" as x, 0 as prefix_length),
  (b"\xAA\xBB\xCC\xDD", 11), (b"\xAA\xBB\xCC\xDD", 12),
  (b"\xAA\xBB\xCC\xDD", 24), (b"\xAA\xBB\xCC\xDD", 32),
  (b'0123456789@ABCDE', 80)
]);

/*-----------------------------------------------------------------------------+
 | addr_bin            | prefix_length | ip_trunc                              |
 +-----------------------------------------------------------------------------+
 | b"\xaa\xbb\xcc\xdd" | 0             | b"\x00\x00\x00\x00"                   |
 | b"\xaa\xbb\xcc\xdd" | 11            | b"\xaa\xa0\x00\x00"                   |
 | b"\xaa\xbb\xcc\xdd" | 12            | b"\xaa\xb0\x00\x00"                   |
 | b"\xaa\xbb\xcc\xdd" | 24            | b"\xaa\xbb\xcc\x00"                   |
 | b"\xaa\xbb\xcc\xdd" | 32            | b"\xaa\xbb\xcc\xdd"                   |
 | b"0123456789@ABCDE" | 80            | b"0123456789\x00\x00\x00\x00\x00\x00" |
 +-----------------------------------------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
