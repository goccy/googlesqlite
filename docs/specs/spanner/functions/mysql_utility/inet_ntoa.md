---
name: INET_NTOA
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindInetNtoa in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet_ntoa
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet_ntoa
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/inet_ntoa.yaml
---

# INET_NTOA

## Summary

Returns the dotted-quad string representation of an IPv4 address given as a packed integer.

## Signatures

- `INET_NTOA(n)`

## Arguments

- `n`: `INT64` in `[0, 4294967295]`.

## Return type

`STRING` of the form `"a.b.c.d"`.

## Behavior

- Out-of-range `n` returns `NULL`.
- `INET_NTOA(INET_ATON(s)) = s` for canonical inputs.

## Examples

```sql
SELECT INET_NTOA(3232235521);   -- "192.168.0.1"
SELECT INET_NTOA(0);            -- "0.0.0.0"
```

## Edge cases

- For IPv6, use `INET6_NTOA`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet_ntoa>.
