---
name: FROM_DAYS
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindFromDays in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#from_days
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#from_days
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/from_days.yaml
---

# FROM_DAYS

## Summary

Returns the `DATE` whose day number (counted from `0000-01-01` in the proleptic Gregorian calendar) is `n`.

## Signatures

- `FROM_DAYS(n)`

## Arguments

- `n`: `INT64` day number; the inverse of `TO_DAYS`.

## Return type

`DATE`.

## Behavior

- `FROM_DAYS(TO_DAYS(d)) = d` for valid dates.
- Values smaller than the day number for `0001-01-01` (`366`) are not meaningful and may be rejected; consult the upstream reference for the exact lower bound.
- Returns `NULL` if `n` is `NULL`.

## Examples

```sql
SELECT FROM_DAYS(739330);   -- DATE "2024-03-15"
```

## Edge cases

- `FROM_DAYS(0)` is documented as returning `0000-00-00`, but `DATE` does not represent such values; callers should clamp.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#from_days>.
