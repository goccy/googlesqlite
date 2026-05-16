---
name: DAYOFWEEK
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Returns the MySQL-convention day-of-week (1=Sunday ... 7=Saturday). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_dayofweek in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofweek
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofweek
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/dayofweek.yaml
---

# DAYOFWEEK

## Summary

Returns the weekday index of a date with **Sunday = 1** through **Saturday = 7**.

## Signatures

- `DAYOFWEEK(value)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.

## Return type

`INT64` in `[1, 7]`.

## Behavior

- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT DAYOFWEEK(DATE "2024-03-15");   -- 6 (Friday)
SELECT DAYOFWEEK(DATE "2024-03-17");   -- 1 (Sunday)
```

## Edge cases

- For ISO 8601 day numbering (Monday = 1 … Sunday = 7), use `WEEKDAY(value) + 1`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofweek>.
