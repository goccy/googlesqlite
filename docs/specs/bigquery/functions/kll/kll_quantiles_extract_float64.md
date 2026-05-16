---
name: KLL_QUANTILES.EXTRACT_FLOAT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_double
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_double
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_extract_float64.yaml
---

# KLL_QUANTILES.EXTRACT_FLOAT64

## Summary

Scalar function that extracts approximate quantiles from a `BYTES`
KLL sketch initialised on `FLOAT64` data.

## Signatures

- `KLL_QUANTILES.EXTRACT_FLOAT64(sketch, num_quantiles)`

## Behavior

- Equivalent to `KLL_QUANTILES.EXTRACT_INT64` but accepts sketches
  whose underlying data type is `FLOAT64`.
- Returns an `ARRAY<FLOAT64>` whose elements are the exact min, the
  approximate quantile boundaries (`num_quantiles - 1` of them), and
  the exact max — `num_quantiles + 1` elements total.
- Orders values using the GoogleSQL floating-point sort order
  (`NaN` precedes `-inf`).

## Examples

```sql
SELECT KLL_QUANTILES.EXTRACT_FLOAT64(kll_sketch, 4) AS quartiles
FROM (SELECT KLL_QUANTILES.INIT_FLOAT64(x, 1000) AS kll_sketch
      FROM UNNEST([1.0, 2.0, 3.0, 4.0, 5.0]) AS x);
-- expected: [1.0, 2.0, 3.0, 4.0, 5.0]
```

## Edge cases

- Returns an error when the sketch's underlying type is not
  `FLOAT64`.
- Returns an error when the input is not a valid KLL sketch.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_double>.
