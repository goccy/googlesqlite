---
name: MONTHNAME
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindMonthName in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#monthname
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#monthname
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/monthname.yaml
---

# MONTHNAME

## Summary

Returns the English month name of a date (e.g. "March").

## Signatures

- `MONTHNAME(value)`

## Return type

`STRING` — one of `"January"` … `"December"`.

## Behavior

- Always returns the English name regardless of session locale.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT MONTHNAME(DATE "2024-03-15");   -- "March"
```

## Edge cases

- For numeric month, use `MONTH`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#monthname>.
