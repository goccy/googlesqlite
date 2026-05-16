---
name: SECOND
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Extracts the second-of-minute (0-59). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_second in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#second
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#second
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/second.yaml
---

# SECOND

## Summary

Returns the seconds component (0–59) of a `TIME`, `DATETIME`, or `TIMESTAMP`.

## Signatures

- `SECOND(value)`

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(SECOND FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT SECOND(TIMESTAMP "2024-03-15 14:30:45+00");   -- 45
```

## Edge cases

- Sub-second precision is not included; use `MICROSECOND` for that.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#second>.
