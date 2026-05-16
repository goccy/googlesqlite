---
name: TO_SECONDS
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindToSeconds in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#to_seconds
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#to_seconds
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/to_seconds.yaml
---

# TO_SECONDS

## Summary

Returns the number of seconds since year zero (`0000-01-01 00:00:00`) for `value`.

## Signatures

- `TO_SECONDS(value)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.

## Return type

`INT64`.

## Behavior

- Equivalent to `TO_DAYS(value) * 86400 + TIME_TO_SEC(TIME(value))`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT TO_SECONDS(DATE "2024-03-15");   -- 63884131200
```

## Edge cases

- Sub-second components are truncated.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#to_seconds>.
