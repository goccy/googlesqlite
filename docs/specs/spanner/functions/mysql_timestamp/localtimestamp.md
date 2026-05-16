---
name: LOCALTIMESTAMP
dialect: spanner
category: functions/mysql_timestamp
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses delegate to GoogleSQL CURRENT_TIMESTAMP in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#localtimestamp
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#localtimestamp
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_timestamp/localtimestamp.yaml
---

# LOCALTIMESTAMP

## Summary

Returns the current `DATETIME` in the session time zone. Synonym of `LOCALTIME` and `NOW`.

## Signatures

- `LOCALTIMESTAMP()`
- `LOCALTIMESTAMP(fsp)`

## Return type

`DATETIME`.

## Behavior

- Stable within a single statement.

## Examples

```sql
SELECT LOCALTIMESTAMP();   -- DATETIME of the moment in session zone
```

## Edge cases

- See `LOCALTIME`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#localtimestamp>.
