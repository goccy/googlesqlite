---
name: NET.IP_TO_STRING
dialect: googlesql
category: functions/net
status: implemented
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netip_to_string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_ip_to_string.yaml
---

# NET.IP_TO_STRING

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

## `NET.IP_TO_STRING`

```
NET.IP_TO_STRING(addr_bin)
```

**Description**
Converts an IPv4 or IPv6 address from binary (BYTES) format in network byte
order to text (STRING) format.

If the input is 4 bytes, this function returns an IPv4 address as a STRING. If
the input is 16 bytes, it returns an IPv6 address as a STRING.

If this function receives a `NULL` input, it returns `NULL`. If the input has
a length different from 4 or 16, an `OUT_OF_RANGE` error occurs.

**Return Data Type**

STRING

**Example**

```googlesql
SELECT FORMAT("%T", x) AS addr_bin, NET.IP_TO_STRING(x) AS ip_to_string
FROM UNNEST([
  b"0123",
  b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01",
  b"0123456789@ABCDE",
  b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\xff\xc0\x00\x02\x80"
]) AS x;

/*---------------------------------------------------------------------------------------------------------------+
 | addr_bin                                                            | ip_to_string                            |
 +---------------------------------------------------------------------------------------------------------------+
 | b"0123"                                                             | 48.49.50.51                             |
 | b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01" | ::1                                     |
 | b"0123456789@ABCDE"                                                 | 3031:3233:3435:3637:3839:4041:4243:4445 |
 | b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\xff\xc0\x00\x02\x80" | ::ffff:192.0.2.128                      |
 +---------------------------------------------------------------------------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
