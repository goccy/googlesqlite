---
name: INET_ATON
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindInetAton in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet_aton
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet_aton
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/inet_aton.yaml
---

# INET_ATON

## Summary

Returns the integer (network-byte-order) representation of an IPv4 dotted-quad address string.

## Signatures

- `INET_ATON(addr)`

## Arguments

- `addr`: `STRING` of the form `"a.b.c.d"`.

## Return type

`INT64`.

## Behavior

- Each octet must be in `[0, 255]`.
- Shorthand forms (`"127"`, `"127.1"`) are **not** accepted by Spanners `INET_ATON` — full dotted-quad is required.
- Returns `NULL` for invalid input.

## Examples

```sql
SELECT INET_ATON("192.168.0.1");   -- 3232235521
SELECT INET_ATON("0.0.0.0");       -- 0
SELECT INET_ATON("not-an-ip");     -- NULL
```

## Edge cases

- For IPv6, use `INET6_ATON`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet_aton>.
