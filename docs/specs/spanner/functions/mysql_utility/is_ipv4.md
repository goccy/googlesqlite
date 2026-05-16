---
name: IS_IPV4
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindIsIPv4 in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/is_ipv4.yaml
---

# IS_IPV4

## Summary

Returns `TRUE` if `addr` is a textually valid IPv4 dotted-quad address.

## Signatures

- `IS_IPV4(addr)`

## Arguments

- `addr`: `STRING`.

## Return type

`BOOL`.

## Behavior

- Strict dotted-quad: each octet must be in `[0, 255]` and exactly four octets are required.
- Returns `NULL` only if `addr` is `NULL`.

## Examples

```sql
SELECT IS_IPV4("192.168.0.1");   -- TRUE
SELECT IS_IPV4("::1");           -- FALSE
SELECT IS_IPV4("256.0.0.1");     -- FALSE
```

## Edge cases

- Leading zeros and shorthand forms are not accepted.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_ipv4>.
