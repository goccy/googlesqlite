---
name: LOCALTIME
dialect: spanner
category: functions/mysql_timestamp
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses delegate to GoogleSQL CURRENT_TIMESTAMP in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#localtime
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#localtime
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_timestamp/localtime.yaml
---

# LOCALTIME

## Summary

Returns the current `DATETIME` in the session time zone. Synonym of `LOCALTIMESTAMP` and `NOW`.

## Signatures

- `LOCALTIME()`
- `LOCALTIME(fsp)`  -- where fsp is fractional-seconds precision

## Return type

`DATETIME`.

## Behavior

- Stable within a single statement.
- The optional `fsp` argument selects sub-second precision (0–6).

## Examples

```sql
SELECT LOCALTIME();      -- DATETIME of the moment in session zone
SELECT LOCALTIME(3);     -- DATETIME with millisecond precision
```

## Edge cases

- For deterministic, time-zone-explicit code, prefer `CURRENT_DATETIME`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/timestamp_functions#localtime>.
