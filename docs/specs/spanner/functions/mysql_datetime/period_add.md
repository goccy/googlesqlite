---
name: PERIOD_ADD
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindPeriodAdd in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#period_add
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#period_add
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/period_add.yaml
---

# PERIOD_ADD

## Summary

Adds a number of months to a `YYYYMM`-encoded period and returns the resulting period.

## Signatures

- `PERIOD_ADD(period, months)`

## Arguments

- `period`: `INT64` encoded as `YYYYMM` or `YYMM` (4-digit form is required for full year disambiguation).
- `months`: `INT64`.

## Return type

`INT64` in `YYYYMM` form.

## Behavior

- Two-digit years follow MySQL conversion: `00`–`69` map to `2000`–`2069`; `70`–`99` map to `1970`–`1999`.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT PERIOD_ADD(202403, 5);   -- 202408
SELECT PERIOD_ADD(202403, -5);  -- 202310
```

## Edge cases

- Periods such as `2024 13` (invalid month) are rejected.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#period_add>.
