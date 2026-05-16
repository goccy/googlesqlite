---
name: DATEDIFF
dialect: spanner
category: functions/mysql_timestamp
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindDateDiff in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#datediff
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#datediff
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_timestamp/datediff.yaml
---

# DATEDIFF

## Summary

Returns the number of days between two date/datetime values: `expr1 - expr2`.

## Signatures

- `DATEDIFF(expr1, expr2)`

## Arguments

- `expr1`, `expr2`: `DATE`, `DATETIME`, or `TIMESTAMP`. Time-of-day components are ignored.

## Return type

`INT64`. Sign matches `expr1 - expr2`.

## Behavior

- Equivalent to `DATE_DIFF(DATE(expr1), DATE(expr2), DAY)`.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT DATEDIFF(DATE "2024-03-20", DATE "2024-03-15");   -- 5
SELECT DATEDIFF(DATE "2024-03-15", DATE "2024-03-20");   -- -5
```

## Edge cases

- Ignores time-of-day even when `TIMESTAMP` is supplied.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#datediff>.
