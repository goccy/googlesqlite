---
name: KLL_QUANTILES.EXTRACT_INT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_int64
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_int64
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_extract_int64.yaml
---

# KLL_QUANTILES.EXTRACT_INT64

## Summary
Scalar function that extracts a chosen number of approximate quantiles from
an `INT64`-initialized KLL sketch, returning the sketch's minimum value, the
requested approximate quantiles, and the sketch's maximum value.

## Signatures
- `KLL_QUANTILES.EXTRACT_INT64(sketch, num_quantiles)`

## Behavior
- `sketch` is a `BYTES` KLL quantiles sketch initialized on the `INT64` data
  type (typically produced by `KLL_QUANTILES.INIT_INT64` or
  `KLL_QUANTILES.MERGE_PARTIAL`).
- `num_quantiles` is a positive `INT64` specifying how many roughly
  equal-sized subsets the captured input values are partitioned into; its
  maximum value is 100,000.
- Returns an `ARRAY<INT64>` of length `num_quantiles + 1`, ordered as:
  the minimum input value, each approximate quantile in ascending order,
  then the maximum input value.
- For example, with `num_quantiles = 3`, a result of `[0, 34, 67, 100]`
  reports `0` as the minimum, `34` and `67` as the approximate quantiles,
  and `100` as the maximum, describing the segments `0..34`, `34..67`,
  and `67..100`.
- Errors are produced if `sketch` is not a valid KLL quantiles sketch, or
  if its underlying data type is not `INT64`.
- This scalar form is similar in semantics to the aggregate
  `KLL_QUANTILES.MERGE_INT64`, but operates on a single already-built
  sketch rather than aggregating multiple sketches.

## Examples
```sql
WITH Data AS (
  SELECT x FROM UNNEST(GENERATE_ARRAY(1, 100)) AS x
)
SELECT
  KLL_QUANTILES.EXTRACT_INT64(kll_sketch, 2) AS halves,
  KLL_QUANTILES.EXTRACT_INT64(kll_sketch, 3) AS terciles,
  KLL_QUANTILES.EXTRACT_INT64(kll_sketch, 4) AS quartiles,
  KLL_QUANTILES.EXTRACT_INT64(kll_sketch, 6) AS sextiles,
FROM (SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch FROM Data);
-- expected:
--   halves    = [1, 50, 100]
--   terciles  = [1, 34, 67, 100]
--   quartiles = [1, 25, 50, 75, 100]
--   sextiles  = [1, 17, 34, 50, 67, 84, 100]
```

## Edge cases
- Errors if `sketch` is not a valid KLL quantiles sketch.
- Errors if the sketch's underlying type is anything other than `INT64`
  (use `KLL_QUANTILES.EXTRACT_FLOAT64` / `EXTRACT_UINT64` for those types).
- Errors if `num_quantiles` exceeds 100,000 or is not positive.
- Quantile values returned are approximate; the minimum and maximum,
  however, are the exact extremes seen by the sketch's input.
- The output array length is always `num_quantiles + 1`, not
  `num_quantiles`, because both extremes are included alongside the
  internal cut points.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_int64>.
