---
name: KLL_QUANTILES.MERGE_POINT_FLOAT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_point_double
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_point_double
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_merge_point_float64.yaml
---

# KLL_QUANTILES.MERGE_POINT_FLOAT64

## Summary

Aggregate function that merges multiple `FLOAT64`-initialised KLL
sketches and returns a single approximate quantile at fraction
`phi`.

## Signatures

- `KLL_QUANTILES.MERGE_POINT_FLOAT64(sketch, phi)`

## Behavior

- Equivalent to `KLL_QUANTILES.MERGE_POINT_INT64` but accepts
  sketches whose underlying data type is `FLOAT64`.
- Orders values using the GoogleSQL floating-point sort order
  (`NaN` precedes `-inf`).
- `phi` must be a `FLOAT64` between 0 and 1.
- Returns `FLOAT64` — the value `v` such that approximately
  `phi * total` inputs are `<= v`.

## Examples

```sql
SELECT KLL_QUANTILES.MERGE_POINT_FLOAT64(kll_sketch, 0.5) AS median
FROM (SELECT KLL_QUANTILES.INIT_FLOAT64(x, 1000) AS kll_sketch
      FROM UNNEST([1.0, 2.0, 3.0, 4.0, 5.0]) AS x
      UNION ALL
      SELECT KLL_QUANTILES.INIT_FLOAT64(x, 1000) AS kll_sketch
      FROM UNNEST([6.0, 7.0, 8.0, 9.0, 10.0]) AS x);
-- expected approximate median around 5.5
```

## Edge cases

- Returns an error when the input sketch's underlying type is not
  `FLOAT64`.
- Returns an error when an input is not a valid KLL sketch.
- `phi` outside `[0, 1]` raises an error.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_point_double>.
