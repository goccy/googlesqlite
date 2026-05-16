---
name: MONTH
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Extracts the month component (1-12). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_month in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#month
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#month
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/month.yaml
---

# MONTH

## Summary

Returns the month number (1–12) of a `DATE`, `DATETIME`, or `TIMESTAMP`.

## Signatures

- `MONTH(value)`

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(MONTH FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT MONTH(DATE "2024-03-15");   -- 3
```

## Edge cases

- For the English name, use `MONTHNAME`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#month>.
