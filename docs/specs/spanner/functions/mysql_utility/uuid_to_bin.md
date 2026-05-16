---
name: UUID_TO_BIN
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindUUIDToBin in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#uuid_to_bin
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#uuid_to_bin
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/uuid_to_bin.yaml
---

# UUID_TO_BIN

## Summary

Converts a UUID `STRING` to its 16-byte binary form. Inverse of `BIN_TO_UUID`.

## Signatures

- `UUID_TO_BIN(s)`
- `UUID_TO_BIN(s, swap_flag)`

## Arguments

- `s`: `STRING` UUID in canonical form.
- `swap_flag`: optional `BOOL`. If `TRUE`, swaps time-low and time-high fields (matching MySQLs time-ordered form).

## Return type

`BYTES` of length 16.

## Behavior

- Invalid UUID input raises an error.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT BYTE_LENGTH(UUID_TO_BIN("00000000-0000-0000-0000-000000000000"));   -- 16
SELECT BIN_TO_UUID(UUID_TO_BIN("123e4567-e89b-12d3-a456-426614174000"));
-- "123e4567-e89b-12d3-a456-426614174000"
```

## Edge cases

- The `swap_flag` form is intended for v1 (time-based) UUIDs; for v4 (random) it just rearranges bytes without semantic meaning.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#uuid_to_bin>.
