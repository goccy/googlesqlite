---
name: PERIOD_DIFF
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindPeriodDiff in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#period_diff
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#period_diff
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/period_diff.yaml
---

# PERIOD_DIFF

## Summary

Returns the number of months between two `YYYYMM`-encoded periods.

## Signatures

- `PERIOD_DIFF(p1, p2)`

## Arguments

- `p1`, `p2`: `INT64` encoded as `YYYYMM` (or `YYMM`, see `PERIOD_ADD`).

## Return type

`INT64`. Sign matches `p1 - p2` in months.

## Behavior

- Two-digit years follow MySQLs pivot at `70`.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT PERIOD_DIFF(202403, 202308);   -- 7
SELECT PERIOD_DIFF(202308, 202403);   -- -7
```

## Edge cases

- Inputs with month component `0` or `> 12` are rejected.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#period_diff>.
