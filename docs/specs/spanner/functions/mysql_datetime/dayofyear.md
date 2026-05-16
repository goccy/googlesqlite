---
name: DAYOFYEAR
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Returns the 1-based day-of-year. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_dayofyear in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofyear
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofyear
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/dayofyear.yaml
---

# DAYOFYEAR

## Summary

Returns the day of the year (1–366).

## Signatures

- `DAYOFYEAR(value)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.

## Return type

`INT64`.

## Behavior

- January 1 returns `1`; December 31 returns `365` or `366` for leap years.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT DAYOFYEAR(DATE "2024-03-15");   -- 75
SELECT DAYOFYEAR(DATE "2024-12-31");   -- 366  (2024 is a leap year)
```

## Edge cases

- Equivalent to `EXTRACT(DAYOFYEAR FROM value)`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayofyear>.
