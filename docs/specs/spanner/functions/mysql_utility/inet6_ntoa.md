---
name: INET6_NTOA
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindInet6Ntoa in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet6_ntoa
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet6_ntoa
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/inet6_ntoa.yaml
---

# INET6_NTOA

## Summary

Returns the canonical textual form of a packed IP address.

## Signatures

- `INET6_NTOA(bin)`

## Arguments

- `bin`: `BYTES` of length 4 (IPv4) or 16 (IPv6).

## Return type

`STRING`.

## Behavior

- IPv6 output uses RFC 5952 canonical form (lowercase hex, longest run of zero-groups compressed to `::`).
- Returns `NULL` if `bin` is not 4 or 16 bytes.

## Examples

```sql
SELECT INET6_NTOA(INET6_ATON("::1"));            -- "::1"
SELECT INET6_NTOA(INET6_ATON("192.168.0.1"));    -- "192.168.0.1"
```

## Edge cases

- An IPv4-mapped IPv6 address renders as `"::ffff:a.b.c.d"`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet6_ntoa>.
