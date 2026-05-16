---
name: WEEK
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Returns the ISO-8601 week number. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_week in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#week
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#week
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/week.yaml
---

# WEEK

## Summary

Returns the ISO/locale week number of a date. The exact numbering scheme is controlled by an optional `mode` argument.

## Signatures

- `WEEK(value)`
- `WEEK(value, mode)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.
- `mode`: `INT64` in `[0, 7]` selecting the week-numbering scheme. The default mode follows the session setting (typically `0`).

## Return type

`INT64`.

## Behavior

- Modes encode whether the week starts on Sunday or Monday and whether week 1 is the first week containing the first day of the year or contains 4+ days, etc. See the upstream reference for the table.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT WEEK(DATE "2024-03-15");      -- locale-default week number
SELECT WEEK(DATE "2024-03-15", 1);   -- ISO 8601 numbering
```

## Edge cases

- Use `WEEKOFYEAR` for the unconfigurable ISO 8601 form.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#week>.
