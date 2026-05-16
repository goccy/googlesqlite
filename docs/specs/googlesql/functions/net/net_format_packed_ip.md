---
name: NET.FORMAT_PACKED_IP
dialect: googlesql
category: functions/net
status: implemented
notes: |
  Either the catalog signature uses bytes/string variants we have not aligned with our runtime, or the UDF is missing.  will sweep the net.* family alongside the dialect plumbing.
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netformat_packed_ip-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_format_packed_ip.yaml
---

# NET.FORMAT_PACKED_IP

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

## `NET.FORMAT_PACKED_IP` (DEPRECATED) 
<a id="net_format_packed_ip"></a>

```
NET.FORMAT_PACKED_IP(bytes_value)
```

**Description**

This function is deprecated. It's the same as [`NET.IP_TO_STRING`][net-link-to-ip-to-string].

**Return Data Type**

STRING

[net-link-to-ip-to-string]: #netip_to_string

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
