---
name: DAYNAME
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindDayName in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayname
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayname
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/dayname.yaml
---

# DAYNAME

## Summary

Returns the English weekday name of a date (e.g. "Monday").

## Signatures

- `DAYNAME(value)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.

## Return type

`STRING` — one of `"Sunday"` … `"Saturday"`.

## Behavior

- Always returns the English name regardless of session locale.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT DAYNAME(DATE "2024-03-15");   -- "Friday"
```

## Edge cases

- For numeric weekday codes, use `DAYOFWEEK` (1=Sunday) or `WEEKDAY` (0=Monday).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#dayname>.
