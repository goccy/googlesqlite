---
name: DEGREES
dialect: spanner
category: functions/mysql_numeric
status: implemented
notes: |
  Converts radians to degrees. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_degrees in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#degrees
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#degrees
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_numeric/degrees.yaml
---

# DEGREES

## Summary

Converts radians to degrees. `DEGREES(x) = x * 180 / PI()`.

## Signatures

- `DEGREES(x)`

## Arguments

- `x`: `FLOAT64` (radians).

## Return type

`FLOAT64` (degrees).

## Behavior

- Returns `NULL` if `x` is `NULL`.

## Examples

```sql
SELECT DEGREES(ACOS(-1));   -- 180.0
SELECT DEGREES(0);          -- 0.0
```

## Edge cases

- `DEGREES(NaN)` returns `NaN`; `DEGREES(±Inf)` returns `±Inf`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#degrees>.
