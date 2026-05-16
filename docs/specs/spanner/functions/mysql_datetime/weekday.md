---
name: WEEKDAY
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Returns the MySQL-convention day-of-week (0=Monday ... 6=Sunday). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_weekday in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#weekday
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#weekday
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/weekday.yaml
---

# WEEKDAY

## Summary

Returns the weekday of a date with **Monday = 0** through **Sunday = 6**.

## Signatures

- `WEEKDAY(value)`

## Return type

`INT64` in `[0, 6]`.

## Behavior

- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT WEEKDAY(DATE "2024-03-15");   -- 4 (Friday)
SELECT WEEKDAY(DATE "2024-03-17");   -- 6 (Sunday)
```

## Edge cases

- For Sunday-first numbering (1=Sunday … 7=Saturday), use `DAYOFWEEK`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#weekday>.
