---
name: NET.IP_NET_MASK
dialect: googlesql
category: functions/net
status: implemented
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netip_net_mask
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_ip_net_mask.yaml
---

# NET.IP_NET_MASK

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

## `NET.IP_NET_MASK`

```
NET.IP_NET_MASK(num_output_bytes, prefix_length)
```

**Description**

Returns a network mask: a byte sequence with length equal to `num_output_bytes`,
where the first `prefix_length` bits are set to 1 and the other bits are set to
0. `num_output_bytes` and `prefix_length` are INT64.
This function throws an error if `num_output_bytes` isn't 4 (for IPv4) or 16
(for IPv6). It also throws an error if `prefix_length` is negative or greater
than `8 * num_output_bytes`.

**Return Data Type**

BYTES

**Example**

```googlesql
SELECT x, y, FORMAT("%T", NET.IP_NET_MASK(x, y)) AS ip_net_mask
FROM UNNEST([
  STRUCT(4 as x, 0 as y),
  (4, 20),
  (4, 32),
  (16, 0),
  (16, 1),
  (16, 128)
]);

/*--------------------------------------------------------------------------------+
 | x  | y   | ip_net_mask                                                         |
 +--------------------------------------------------------------------------------+
 | 4  | 0   | b"\x00\x00\x00\x00"                                                 |
 | 4  | 20  | b"\xff\xff\xf0\x00"                                                 |
 | 4  | 32  | b"\xff\xff\xff\xff"                                                 |
 | 16 | 0   | b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" |
 | 16 | 1   | b"\x80\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" |
 | 16 | 128 | b"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" |
 +--------------------------------------------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
