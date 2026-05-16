---
name: MICROSECOND
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Extracts the microsecond-of-second component. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_microsecond in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#microsecond
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#microsecond
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/microsecond.yaml
---

# MICROSECOND

## Summary

Returns the microseconds component (0–999999) of a `TIME`, `DATETIME`, or `TIMESTAMP`.

## Signatures

- `MICROSECOND(value)`

## Return type

`INT64`.

## Behavior

- Equivalent to `EXTRACT(MICROSECOND FROM value)`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT MICROSECOND(TIMESTAMP "2024-03-15 14:30:00.123456+00");   -- 123456
```

## Edge cases

- Inputs without sub-second precision return `0`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#microsecond>.
