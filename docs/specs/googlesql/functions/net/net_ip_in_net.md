---
name: NET.IP_IN_NET
dialect: googlesql
category: functions/net
status: implemented
notes: |
  Either the catalog signature uses bytes/string variants we have not aligned with our runtime, or the UDF is missing.  will sweep the net.* family alongside the dialect plumbing.
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netip_in_net
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_ip_in_net.yaml
---

# NET.IP_IN_NET

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

## `NET.IP_IN_NET`

```
NET.IP_IN_NET(address, subnet)
```

**Description**

Takes an IP address and a subnet CIDR as STRING and returns true if the IP
address is contained in the subnet.

This function supports the following formats for `address` and `subnet`:

+ IPv4: Dotted-quad format. For example, `10.1.2.3`.
+ IPv6: Colon-separated format. For example,
  `1234:5678:90ab:cdef:1234:5678:90ab:cdef`. For more examples, see the
  [IP Version 6 Addressing Architecture][net-link-to-ipv6-rfc].
+ CIDR (IPv4): Dotted-quad format. For example, `10.1.2.0/24`
+ CIDR (IPv6): Colon-separated format. For example, `1:2::/48`.

If this function receives a `NULL` input, it returns `NULL`. If the input is
considered invalid, an `OUT_OF_RANGE` error occurs.

**Return Data Type**

BOOL

[net-link-to-ipv6-rfc]: http://www.ietf.org/rfc/rfc2373.txt

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
