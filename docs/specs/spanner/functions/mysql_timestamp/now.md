---
name: NOW
dialect: spanner
category: functions/mysql_timestamp
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses delegate to GoogleSQL CURRENT_TIMESTAMP in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#now
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#now
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_timestamp/now.yaml
---

# NOW

## Summary

Returns the current `DATETIME` in the session time zone. The 1-argument form selects sub-second precision.

## Signatures

- `NOW()`
- `NOW(fsp)`

## Return type

`DATETIME`.

## Behavior

- Stable within a single statement.
- Different from `SYSDATE`: `NOW()` is memoized per statement, `SYSDATE` re-reads the clock on every call.

## Examples

```sql
SELECT NOW();        -- DATETIME of the moment
SELECT NOW(6);       -- with microsecond precision
```

## Edge cases

- For UTC, use `UTC_TIMESTAMP`. For GoogleSQL-native, use `CURRENT_TIMESTAMP`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#now>.
