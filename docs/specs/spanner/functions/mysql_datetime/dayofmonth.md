---
name: DAYOFMONTH
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Synonym of DAY; returns the 1-based day-of-month. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_dayofmonth in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofmonth
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofmonth
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/dayofmonth.yaml
---

# DAYOFMONTH

## Summary

Returns the day of the month (1–31). Equivalent to `DAY`.

## Signatures

- `DAYOFMONTH(value)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(DAY FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT DAYOFMONTH(DATE "2024-03-15");   -- 15
```

## Edge cases

- Synonym of `DAY`. Both are provided for MySQL portability.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofmonth>.
