---
name: MINUTE
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Extracts the minute-of-hour (0-59). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_minute in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#minute
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#minute
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/minute.yaml
---

# MINUTE

## Summary

Returns the minute component (0–59) of a `TIME`, `DATETIME`, or `TIMESTAMP`.

## Signatures

- `MINUTE(value)`

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(MINUTE FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT MINUTE(TIMESTAMP "2024-03-15 14:30:45+00");   -- 30
```

## Edge cases

- For `TIMESTAMP`, the minute is reported in the session time zone.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#minute>.
