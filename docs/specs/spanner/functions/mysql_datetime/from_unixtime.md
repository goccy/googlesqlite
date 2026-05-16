---
name: FROM_UNIXTIME
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindFromUnixtime in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#from_unixtime
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#from_unixtime
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/from_unixtime.yaml
---

# FROM_UNIXTIME

## Summary

Returns the `TIMESTAMP` (or formatted string) corresponding to a Unix epoch timestamp in seconds.

## Signatures

- `FROM_UNIXTIME(epoch_seconds)`
- `FROM_UNIXTIME(epoch_seconds, format)`

## Arguments

- `epoch_seconds`: `INT64` or `FLOAT64`. Fractional seconds are honored.
- `format`: optional MySQL-style format string; when supplied, the result is the formatted string instead of a `TIMESTAMP`.

## Return type

`TIMESTAMP` (1-arg form) or `STRING` (2-arg form).

## Behavior

- Negative `epoch_seconds` produce timestamps before the epoch.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT FROM_UNIXTIME(1710518400);                       -- TIMESTAMP "2024-03-15 14:40:00+00"
SELECT FROM_UNIXTIME(1710518400, "%Y-%m-%d %H:%i:%s");  -- "2024-03-15 14:40:00"
```

## Edge cases

- The session time zone affects the formatted string in the 2-arg form.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#from_unixtime>.
