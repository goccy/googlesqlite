---
name: RADIANS
dialect: spanner
category: functions/mysql_numeric
status: implemented
notes: |
  Converts degrees to radians. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_radians in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#radians
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#radians
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_numeric/radians.yaml
---

# RADIANS

## Summary

Converts degrees to radians. `RADIANS(x) = x * PI() / 180`.

## Signatures

- `RADIANS(x)`

## Arguments

- `x`: `FLOAT64` (degrees).

## Return type

`FLOAT64` (radians).

## Behavior

- Returns `NULL` if `x` is `NULL`.

## Examples

```sql
SELECT RADIANS(180);   -- 3.141592653589793
SELECT RADIANS(0);     -- 0.0
```

## Edge cases

- Identical to a manual `x * 3.14159... / 180` but uses the maximally precise `PI()` constant.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#radians>.
