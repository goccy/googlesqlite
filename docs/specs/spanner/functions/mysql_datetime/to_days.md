---
name: TO_DAYS
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindToDays in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#to_days
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#to_days
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/to_days.yaml
---

# TO_DAYS

## Summary

Returns the day number of `value`, counted from year zero (`0000-01-01`) in the proleptic Gregorian calendar.

## Signatures

- `TO_DAYS(value)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.

## Return type

`INT64`.

## Behavior

- Inverse of `FROM_DAYS`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT TO_DAYS(DATE "2024-03-15");   -- 739330
SELECT FROM_DAYS(TO_DAYS(DATE "2024-03-15"));   -- DATE "2024-03-15"
```

## Edge cases

- Dates earlier than `0001-01-01` (day number `366`) are not meaningful for typical date arithmetic.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#to_days>.
