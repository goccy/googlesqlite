---
name: IS_IPV4_COMPAT
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindIsIPv4Compat in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4_compat
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4_compat
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/is_ipv4_compat.yaml
---

# IS_IPV4_COMPAT

## Summary

Returns `TRUE` if `bin` is a 16-byte IPv6 packed address whose first 12 bytes are zero (i.e. the deprecated `"IPv4-compatible"` form `::a.b.c.d`).

## Signatures

- `IS_IPV4_COMPAT(bin)`

## Arguments

- `bin`: `BYTES`. Anything other than 16 bytes returns `FALSE` (or `NULL` if input is `NULL`).

## Return type

`BOOL`.

## Behavior

- The IPv4-compatible IPv6 prefix is `0:0:0:0:0:0:` (96 zero bits).
- The all-zero address (`::`) and the loopback (`::1`) also fall in this range and return `TRUE`.

## Examples

```sql
SELECT IS_IPV4_COMPAT(INET6_ATON("::192.168.0.1"));   -- TRUE
SELECT IS_IPV4_COMPAT(INET6_ATON("::ffff:192.168.0.1")); -- FALSE  (IPv4-mapped, not -compatible)
```

## Edge cases

- The IPv4-compatible form is deprecated by RFC 4291; prefer IPv4-mapped (`::ffff:a.b.c.d`).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4_compat>.
