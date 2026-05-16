---
name: NET.IPV4_TO_INT64
dialect: googlesql
category: functions/net
status: implemented
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netipv4_to_int64
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_ipv4_to_int64.yaml
---

# NET.IPV4_TO_INT64

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

## `NET.IPV4_TO_INT64`

```
NET.IPV4_TO_INT64(addr_bin)
```

**Description**

Converts an IPv4 address from binary (BYTES) format in network byte order to
integer format. In the integer output, the least significant bit of the IP
address is stored in the least significant bit of the integer, regardless of
host or client architecture. For example, `1` means `0.0.0.1`, and `0x1FF` means
`0.0.1.255`. The output is in the range `[0, 0xFFFFFFFF]`.

If the input length isn't 4, this function throws an error.

This function doesn't support IPv6.

**Return Data Type**

INT64

**Example**

```googlesql
SELECT
  FORMAT("%T", x) AS addr_bin,
  FORMAT("0x%X", NET.IPV4_TO_INT64(x)) AS ipv4_to_int64
FROM
UNNEST([b"\x00\x00\x00\x00", b"\x00\xab\xcd\xef", b"\xff\xff\xff\xff"]) AS x;

/*-------------------------------------+
 | addr_bin            | ipv4_to_int64 |
 +-------------------------------------+
 | b"\x00\x00\x00\x00" | 0x0           |
 | b"\x00\xab\xcd\xef" | 0xABCDEF      |
 | b"\xff\xff\xff\xff" | 0xFFFFFFFF    |
 +-------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
