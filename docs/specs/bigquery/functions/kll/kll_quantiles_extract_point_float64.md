---
name: KLL_QUANTILES.EXTRACT_POINT_FLOAT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_point_double
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_point_double
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_extract_point_float64.yaml
---

# KLL_QUANTILES.EXTRACT_POINT_FLOAT64

## Summary

Scalar function that extracts a single approximate quantile at
fraction `phi` from a `FLOAT64`-initialised KLL sketch.

## Signatures

- `KLL_QUANTILES.EXTRACT_POINT_FLOAT64(sketch, phi)`

## Behavior

- Equivalent to `KLL_QUANTILES.EXTRACT_POINT_INT64` but accepts
  sketches whose underlying data type is `FLOAT64`.
- `phi` must be a `FLOAT64` between 0 and 1.
- Returns the value `v` such that approximately `phi * n` inputs are
  `<= v` and `(1 - phi) * n` inputs are `>= v`.
- Returns `FLOAT64`.

## Examples

```sql
SELECT KLL_QUANTILES.EXTRACT_POINT_FLOAT64(kll_sketch, 0.5) AS median
FROM (SELECT KLL_QUANTILES.INIT_FLOAT64(x, 1000) AS kll_sketch
      FROM UNNEST([1.0, 2.0, 3.0, 4.0, 5.0]) AS x);
-- expected: 3.0
```

## Edge cases

- Returns an error when the sketch's underlying type is not
  `FLOAT64`.
- Returns an error when the input is not a valid KLL sketch.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_point_double>.
