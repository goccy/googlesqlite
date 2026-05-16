---
name: YEAR
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Extracts the four-digit year. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_year in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#year
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#year
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/year.yaml
---

# YEAR

## Summary

Returns the year (e.g. 2024) of a `DATE`, `DATETIME`, or `TIMESTAMP`.

## Signatures

- `YEAR(value)`

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(YEAR FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT YEAR(DATE "2024-03-15");   -- 2024
```

## Edge cases

- For ISO year (which can differ near year boundaries), use `EXTRACT(ISOYEAR FROM value)`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#year>.
