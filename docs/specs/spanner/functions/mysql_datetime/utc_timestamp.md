---
name: UTC_TIMESTAMP
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindUtcTimestamp in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#utc_timestamp
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#utc_timestamp
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/utc_timestamp.yaml
---

# UTC_TIMESTAMP

## Summary

Returns the current UTC `DATETIME` at the moment of evaluation.

## Signatures

- `UTC_TIMESTAMP()`

## Return type

`DATETIME` in UTC.

## Behavior

- Always evaluates against UTC, regardless of session time zone.

## Examples

```sql
SELECT UTC_TIMESTAMP();   -- DATETIME in UTC, e.g. "2024-03-15T14:40:00"
```

## Edge cases

- Stable within a single statement (unlike `SYSDATE`).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#utc_timestamp>.
