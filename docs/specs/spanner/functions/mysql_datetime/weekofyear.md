---
name: WEEKOFYEAR
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Synonym of WEEK using ISO-8601 numbering. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_weekofyear in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#weekofyear
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#weekofyear
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/weekofyear.yaml
---

# WEEKOFYEAR

## Summary

Returns the ISO 8601 week-of-year (1–53) of a date. Equivalent to `WEEK(value, 3)`.

## Signatures

- `WEEKOFYEAR(value)`

## Return type

`INT64`.

## Behavior

- Week starts on Monday; week 1 is the first week containing at least 4 days of the new year.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT WEEKOFYEAR(DATE "2024-03-15");   -- 11
SELECT WEEKOFYEAR(DATE "2024-12-31");   -- 1   (belongs to ISO year 2025)
```

## Edge cases

- The ISO year of a date may differ from the calendar year near year boundaries.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#weekofyear>.
