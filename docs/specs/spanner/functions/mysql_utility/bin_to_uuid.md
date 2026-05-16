---
name: BIN_TO_UUID
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindBinToUUID in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#bin_to_uuid
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#bin_to_uuid
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/bin_to_uuid.yaml
---

# BIN_TO_UUID

## Summary

Converts a 16-byte binary UUID to its canonical `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` string form.

## Signatures

- `BIN_TO_UUID(bin)`
- `BIN_TO_UUID(bin, swap_flag)`

## Arguments

- `bin`: `BYTES` of length 16.
- `swap_flag`: optional `BOOL`. If `TRUE`, the time-low and time-high fields are swapped (matching MySQLs "v1-time-ordered" form).

## Return type

`STRING`.

## Behavior

- Returns `NULL` if any argument is `NULL`.
- Length-16 enforcement applies; shorter or longer inputs raise an error.

## Examples

```sql
SELECT BIN_TO_UUID(UUID_TO_BIN(GENERATE_UUID()));   -- canonical UUID form
```

## Edge cases

- The canonical UUID format uses lowercase hex characters.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#bin_to_uuid>.
