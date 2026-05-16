---
name: MAKEDATE
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindMakeDate in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#makedate
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#makedate
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/makedate.yaml
---

# MAKEDATE

## Summary

Constructs a `DATE` from a year and a day-of-year ordinal.

## Signatures

- `MAKEDATE(year, day_of_year)`

## Arguments

- `year`: `INT64` 4-digit year.
- `day_of_year`: `INT64` in `[1, 366]`.

## Return type

`DATE`.

## Behavior

- `day_of_year` past the end of the year rolls into the next year (e.g. `MAKEDATE(2024, 367)` returns `2025-01-01`).
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT MAKEDATE(2024, 75);    -- DATE "2024-03-15"
SELECT MAKEDATE(2024, 1);     -- DATE "2024-01-01"
```

## Edge cases

- `day_of_year <= 0` is rejected.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#makedate>.
