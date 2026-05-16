---
name: NET.PARSE_IP
dialect: googlesql
category: functions/net
status: implemented
notes: |
  Either the catalog signature uses bytes/string variants we have not aligned with our runtime, or the UDF is missing.  will sweep the net.* family alongside the dialect plumbing.
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netparse_ip-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_parse_ip.yaml
---

# NET.PARSE_IP

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

## `NET.PARSE_IP` (DEPRECATED) 
<a id="net_parse_ip"></a>

```
NET.PARSE_IP(addr_str)
```

**Description**

This function is deprecated. It's the same as
[`NET.IPV4_TO_INT64`][net-link-to-ipv4-to-int64]`(`[`NET.IP_FROM_STRING`][net-link-to-ip-from-string]`(addr_str))`,
except that this function truncates the input at the first `'\x00'` character,
if any, while `NET.IP_FROM_STRING` treats `'\x00'` as invalid.

**Return Data Type**

INT64

[net-link-to-ip-to-string]: #netip_to_string

[net-link-to-ipv4-to-int64]: #netipv4_to_int64

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
