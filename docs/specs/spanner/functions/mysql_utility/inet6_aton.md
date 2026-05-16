---
name: INET6_ATON
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindInet6Aton in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet6_aton
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet6_aton
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/inet6_aton.yaml
---

# INET6_ATON

## Summary

Returns the 16-byte (IPv6) or 4-byte (IPv4) packed binary form of an IP address string.

## Signatures

- `INET6_ATON(addr)`

## Arguments

- `addr`: `STRING` containing a textual IP address. Both IPv4 dotted-quad and IPv6 colon-hex forms are accepted.

## Return type

`BYTES` — 4 bytes for IPv4, 16 bytes for IPv6.

## Behavior

- Returns `NULL` for invalid input.
- IPv4-mapped IPv6 addresses (e.g. `"::ffff:192.0.2.1"`) return 16 bytes.

## Examples

```sql
SELECT INET6_ATON("192.168.0.1");   -- 4-byte BYTES
SELECT INET6_ATON("::1");           -- 16-byte BYTES with last byte = 0x01
```

## Edge cases

- For pure IPv4 packing into an `INT64`, use `INET_ATON` instead.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#inet6_aton>.
