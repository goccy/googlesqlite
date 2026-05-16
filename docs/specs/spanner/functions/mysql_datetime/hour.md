---
name: HOUR
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Extracts the hour-of-day (0-23). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_hour in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#hour
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#hour
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/hour.yaml
---

# HOUR

## Summary

Returns the hour component (0–23) of a `TIME`, `DATETIME`, or `TIMESTAMP`. For elapsed-time `TIME` values, the result may exceed 23.

## Signatures

- `HOUR(value)`

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(HOUR FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT HOUR(TIMESTAMP "2024-03-15 14:30:00+00");   -- 14
SELECT HOUR(TIME "08:15:00");                      -- 8
```

## Edge cases

- For `TIMESTAMP`, the hour is reported in the session time zone (or UTC if no zone is set).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#hour>.
