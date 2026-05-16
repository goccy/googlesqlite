---
name: QUARTER
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Returns the calendar quarter (1-4). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_quarter in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#quarter
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#quarter
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/quarter.yaml
---

# QUARTER

## Summary

Returns the calendar quarter (1–4) of a `DATE`, `DATETIME`, or `TIMESTAMP`.

## Signatures

- `QUARTER(value)`

## Return type

`INT64` in `[1, 4]`.

## Behavior

- Equivalent to `EXTRACT(QUARTER FROM value)`.
- Q1 = Jan–Mar, Q2 = Apr–Jun, Q3 = Jul–Sep, Q4 = Oct–Dec.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT QUARTER(DATE "2024-03-15");   -- 1
SELECT QUARTER(DATE "2024-08-01");   -- 3
```

## Edge cases

- Fiscal-year quarter calculations require date arithmetic on the input first.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#quarter>.
