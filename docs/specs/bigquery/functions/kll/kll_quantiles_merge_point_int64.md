---
name: KLL_QUANTILES.MERGE_POINT_INT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_point_int64
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_point_int64
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_merge_point_int64.yaml
---

# KLL_QUANTILES.MERGE_POINT_INT64

## Summary
Aggregate function that merges `INT64`-typed KLL quantiles sketches and
returns a single approximate quantile value from the merged sketch, where
the quantile point is selected by `phi` (a fraction between 0 and 1).

## Signatures
- `KLL_QUANTILES.MERGE_POINT_INT64(sketch, phi)`

## Behavior
- `sketch` is a `BYTES` value that must be a KLL quantiles sketch
  initialized over `INT64` data (for example, by `KLL_QUANTILES.INIT_INT64`).
- `phi` is a `FLOAT64` between 0 and 1 specifying the quantile to extract,
  expressed as a fraction of the input row count: the function returns a
  value `v` such that approximately `phi * n` inputs are less than or equal
  to `v` and `(1 - phi) * n` inputs are greater than or equal to `v`.
- It is an aggregate function: it merges every `sketch` value across the
  group and then evaluates the quantile on the merged sketch.
- When the merged sketches were initialized with different precisions, the
  result precision is downgraded to the lowest precision involved in the
  merge, except when the aggregations are small enough to still capture the
  input exactly, in which case the contributing sketch's precision is
  preserved.
- Returns the approximate quantile as `INT64`.

## Examples
```sql
-- Initialize two KLL sketches over INT64 inputs (5 rows each), merge
-- them, and return the value at the 90th percentile (phi = 0.9).
SELECT KLL_QUANTILES.MERGE_POINT_INT64(kll_sketch, .9) AS quantile
FROM (SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch
      FROM (SELECT 1 AS x UNION ALL
            SELECT 2 AS x UNION ALL
            SELECT 3 AS x UNION ALL
            SELECT 4 AS x UNION ALL
            SELECT 5)
      UNION ALL
      SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch
      FROM (SELECT 6 AS x UNION ALL
            SELECT 7 AS x UNION ALL
            SELECT 8 AS x UNION ALL
            SELECT 9 AS x UNION ALL
            SELECT 10 AS x));
-- expected: quantile = 9
```

## Edge cases
- Returns an error if the underlying type of one or more input sketches is
  not compatible with `INT64` (for example, a sketch produced by
  `KLL_QUANTILES.INIT_DOUBLE`).
- Returns an error if the input is not a valid KLL quantiles sketch.
- `phi` must lie in the closed interval `[0, 1]`; values outside that
  range are not valid quantile fractions.
- Because KLL sketches are approximate, the returned value is an estimate
  of the true quantile; when sketches with different precisions are merged
  the merged precision falls back to the lowest of the inputs (except when
  the input was small enough to still be captured exactly).

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_point_int64>.
