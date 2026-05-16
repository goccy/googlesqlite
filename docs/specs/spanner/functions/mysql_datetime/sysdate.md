---
name: SYSDATE
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindSysDate in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#sysdate
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#sysdate
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/sysdate.yaml
---

# SYSDATE

## Summary

Returns the current server `DATETIME` at the moment of evaluation. Unlike `NOW()`, it is **not** memoized within a single statement: each call may return a slightly different value.

## Signatures

- `SYSDATE()`

## Return type

`DATETIME`.

## Behavior

- Each call re-reads the system clock, so it does not provide a stable transaction-time anchor.
- The value reflects the session time zone.

## Examples

```sql
SELECT SYSDATE();                  -- DATETIME of the moment
SELECT SYSDATE(), SLEEP(2), SYSDATE();  -- two distinct values, ~2s apart
```

## Edge cases

- Prefer `CURRENT_DATETIME` / `CURRENT_TIMESTAMP` for deterministic per-statement values.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#sysdate>.
