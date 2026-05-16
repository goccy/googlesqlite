---
name: NET.SAFE_IP_FROM_STRING
dialect: googlesql
category: functions/net
status: implemented
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netsafe_ip_from_string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_safe_ip_from_string.yaml
---

# NET.SAFE_IP_FROM_STRING

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

## `NET.SAFE_IP_FROM_STRING`

```
NET.SAFE_IP_FROM_STRING(addr_str)
```

**Description**

Similar to [`NET.IP_FROM_STRING`][net-link-to-ip-from-string], but returns `NULL`
instead of throwing an error if the input is invalid.

**Return Data Type**

BYTES

**Example**

```googlesql
SELECT
  addr_str,
  FORMAT("%T", NET.SAFE_IP_FROM_STRING(addr_str)) AS safe_ip_from_string
FROM UNNEST([
  '48.49.50.51',
  '::1',
  '3031:3233:3435:3637:3839:4041:4243:4445',
  '::ffff:192.0.2.128',
  '48.49.50.51/32',
  '48.49.50',
  '::wxyz'
]) AS addr_str;

/*---------------------------------------------------------------------------------------------------------------+
 | addr_str                                | safe_ip_from_string                                                 |
 +---------------------------------------------------------------------------------------------------------------+
 | 48.49.50.51                             | b"0123"                                                             |
 | ::1                                     | b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01" |
 | 3031:3233:3435:3637:3839:4041:4243:4445 | b"0123456789@ABCDE"                                                 |
 | ::ffff:192.0.2.128                      | b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\xff\xc0\x00\x02\x80" |
 | 48.49.50.51/32                          | NULL                                                                |
 | 48.49.50                                | NULL                                                                |
 | ::wxyz                                  | NULL                                                                |
 +---------------------------------------------------------------------------------------------------------------*/
```

[net-link-to-ip-from-string]: #netip_from_string

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
