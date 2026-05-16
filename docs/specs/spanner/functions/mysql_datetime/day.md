---
name: DAY
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Extracts the day-of-month component from a TIMESTAMP-like value. Catalog entry registered in registerSpannerExtensionFunctions (internal/catalog.go); runtime delegate is mysql_day in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#day
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#day
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/day.yaml
---

# DAY

## Summary

Returns the day of the month (1–31) of a `DATE`, `DATETIME`, or `TIMESTAMP`. Synonym of `DAYOFMONTH`.

## Signatures

- `DAY(value)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(DAY FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT DAY(DATE "2024-03-15");   -- 15
```

## Edge cases

- Use `DAYOFWEEK`/`WEEKDAY` for day-of-week; `DAYOFYEAR` for day-of-year.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#day>.
