---
name: APPROX_COSINE_DISTANCE
dialect: spanner
category: functions/math
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_cosine_distance
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_cosine_distance
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/math/approx_cosine_distance.yaml
---

# APPROX_COSINE_DISTANCE

## Summary

Approximate cosine distance between two equal-length numeric arrays. Uses lower-precision arithmetic for speed and is intended for vector-search workloads with tolerance for small numerical error.

## Signatures

- `APPROX_COSINE_DISTANCE(a, b)`

## Arguments

- `a`, `b`: `ARRAY<FLOAT32>` or `ARRAY<FLOAT64>` of equal length.

## Return type

`FLOAT64` in `[0, 2]`.

## Behavior

- `cosine_distance = 1 - cosine_similarity`.
- Returns `NULL` if either argument is `NULL` or contains a `NULL` element.
- Returns an error if either array has zero norm.

## Examples

```sql
SELECT APPROX_COSINE_DISTANCE([1.0, 0.0], [1.0, 0.0]);   -- ~0.0
SELECT APPROX_COSINE_DISTANCE([1.0, 0.0], [0.0, 1.0]);   -- ~1.0
```

## Edge cases

- For exact arithmetic, use `COSINE_DISTANCE` instead.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_cosine_distance>.
