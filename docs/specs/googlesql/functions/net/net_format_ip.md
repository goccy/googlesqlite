---
name: NET.FORMAT_IP
dialect: googlesql
category: functions/net
status: implemented
notes: |
  Either the catalog signature uses bytes/string variants we have not aligned with our runtime, or the UDF is missing.  will sweep the net.* family alongside the dialect plumbing.
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netformat_ip-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_format_ip.yaml
---

# NET.FORMAT_IP

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

## `NET.FORMAT_IP` (DEPRECATED) 
<a id="net_format_ip"></a>

```
NET.FORMAT_IP(integer)
```

**Description**

This function is deprecated. It's the same as
[`NET.IP_TO_STRING`][net-link-to-ip-to-string]`(`[`NET.IPV4_FROM_INT64`][net-link-to-ipv4-from-int64]`(integer))`,
except that this function doesn't allow negative input values.

**Return Data Type**

STRING

[net-link-to-ip-to-string]: #netip_to_string

[net-link-to-ipv4-from-int64]: #netipv4_from_int64

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
