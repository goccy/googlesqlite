---
name: IS_IPV4_MAPPED
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindIsIPv4Mapped in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4_mapped
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4_mapped
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/is_ipv4_mapped.yaml
---

# IS_IPV4_MAPPED

## Summary

Returns `TRUE` if `bin` is a 16-byte IPv6 packed address whose first 80 bits are zero and the next 16 bits are `0xffff` — i.e. the IPv4-mapped IPv6 form `::ffff:a.b.c.d`.

## Signatures

- `IS_IPV4_MAPPED(bin)`

## Return type

`BOOL`.

## Behavior

- Returns `FALSE` for inputs that are not exactly 16 bytes.
- Returns `NULL` only if `bin` is `NULL`.

## Examples

```sql
SELECT IS_IPV4_MAPPED(INET6_ATON("::ffff:192.168.0.1"));  -- TRUE
SELECT IS_IPV4_MAPPED(INET6_ATON("192.168.0.1"));         -- FALSE  (4 bytes)
```

## Edge cases

- Distinguishes IPv4-mapped from IPv4-compatible (see `IS_IPV4_COMPAT`).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4_mapped>.
