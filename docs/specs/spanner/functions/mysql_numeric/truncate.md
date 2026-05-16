---
name: TRUNCATE
dialect: spanner
category: functions/mysql_numeric
status: implemented
notes: |
  Truncates a FLOAT64 toward zero to the requested decimal places. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_truncate in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#truncate
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#truncate
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_numeric/truncate.yaml
---

# TRUNCATE

## Summary

Truncates `x` to `d` decimal places (toward zero).

## Signatures

- `TRUNCATE(x, d)`

## Arguments

- `x`: `FLOAT64` or `NUMERIC`.
- `d`: `INT64`. Negative values truncate digits to the left of the decimal point.

## Return type

Same numeric kind as `x`.

## Behavior

- Differs from `ROUND` in that no half-away-from-zero rounding is applied.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT TRUNCATE(1.999, 1);     -- 1.9
SELECT TRUNCATE(-1.999, 1);    -- -1.9
SELECT TRUNCATE(1234.5, -2);   -- 1200.0
```

## Edge cases

- For GoogleSQL-native code, prefer `TRUNC(x, d)`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#truncate>.
