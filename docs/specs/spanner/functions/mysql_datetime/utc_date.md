---
name: UTC_DATE
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindUtcDate in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#utc_date
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#utc_date
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/utc_date.yaml
---

# UTC_DATE

## Summary

Returns the current UTC `DATE` at the moment of evaluation.

## Signatures

- `UTC_DATE()`

## Return type

`DATE`.

## Behavior

- Always evaluates against UTC, regardless of session time zone.

## Examples

```sql
SELECT UTC_DATE();   -- DATE in UTC, e.g. "2024-03-15"
```

## Edge cases

- Stable within a single statement.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#utc_date>.
