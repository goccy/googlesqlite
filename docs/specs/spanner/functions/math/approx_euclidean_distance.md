---
name: APPROX_EUCLIDEAN_DISTANCE
dialect: spanner
category: functions/math
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_euclidean_distance
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_euclidean_distance
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/math/approx_euclidean_distance.yaml
---

# APPROX_EUCLIDEAN_DISTANCE

## Summary

Approximate Euclidean (L2) distance between two equal-length numeric arrays.

## Signatures

- `APPROX_EUCLIDEAN_DISTANCE(a, b)`

## Return type

`FLOAT64` (non-negative).

## Behavior

- Computes `SQRT(SUM((a[i] - b[i])^2))` in lower-precision arithmetic.
- Returns `NULL` if either argument is `NULL` or contains a `NULL` element.
- Mismatched lengths raise an error.

## Examples

```sql
SELECT APPROX_EUCLIDEAN_DISTANCE([0.0, 0.0], [3.0, 4.0]);   -- ~5.0
```

## Edge cases

- For exact arithmetic, use `EUCLIDEAN_DISTANCE`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_euclidean_distance>.
