---
name: IS_IPV6
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindIsIPv6 in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv6
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv6
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/is_ipv6.yaml
---

# IS_IPV6

## Summary

Returns `TRUE` if `addr` is a textually valid IPv6 address.

## Signatures

- `IS_IPV6(addr)`

## Arguments

- `addr`: `STRING`.

## Return type

`BOOL`.

## Behavior

- Accepts `::` zero-compression and embedded IPv4-mapped form (`::ffff:a.b.c.d`).
- Returns `FALSE` for IPv4 addresses.
- Returns `NULL` only if `addr` is `NULL`.

## Examples

```sql
SELECT IS_IPV6("::1");                   -- TRUE
SELECT IS_IPV6("::ffff:192.168.0.1");    -- TRUE
SELECT IS_IPV6("192.168.0.1");           -- FALSE
```

## Edge cases

- Bracketed `[::1]` is **not** accepted; brackets are URL-host syntax, not address syntax.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv6>.
