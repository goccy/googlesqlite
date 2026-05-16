---
name: KLL_QUANTILES.EXTRACT_POINT_INT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_point_int64
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_point_int64
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_extract_point_int64.yaml
---

# KLL_QUANTILES.EXTRACT_POINT_INT64

## Summary
Scalar function that takes a finalized KLL sketch over `INT64` values and a
target rank `phi` in `[0, 1]`, and returns the approximate quantile at that
single point as an `INT64`.

## Signatures
- `KLL_QUANTILES.EXTRACT_POINT_INT64(sketch, phi)`

## Behavior
- `sketch` is a `BYTES` value produced by `KLL_QUANTILES.INIT_INT64` or by
  `KLL_QUANTILES.MERGE_PARTIAL` over `INT64`-typed inputs; the underlying
  sketch element type must be `INT64`. Passing a sketch built over a
  different value type (for example, `KLL_QUANTILES.INIT_FLOAT64`) is an
  error.
- `phi` is a `FLOAT64` in the closed interval `[0, 1]` selecting a single
  quantile rank: `0` is the minimum, `1` is the maximum, `0.5` is the
  approximate median.
- Returns the approximate `INT64` value at rank `phi` from the sketch. Unlike
  `KLL_QUANTILES.EXTRACT_INT64`, which returns an array of bucket
  boundaries, this function returns one scalar value for the requested
  point.
- The returned value carries the same approximation error bound that the
  sketch was initialized with (the precision argument to `INIT_INT64`).
- The function is non-aggregate: it operates on an already-finalized sketch
  passed as a scalar argument.
- Returns `NULL` when `sketch` is `NULL`.
- A `phi` value outside `[0, 1]` is an error; a `NULL` `phi` propagates as
  `NULL`.

## Examples
```sql
-- Build an INT64 KLL sketch, then extract the median (phi = 0.5).
WITH Points AS (
  SELECT x FROM UNNEST([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) AS x
)
SELECT
  KLL_QUANTILES.EXTRACT_POINT_INT64(
    (SELECT KLL_QUANTILES.INIT_INT64(x, 1000) FROM Points),
    0.5
  ) AS approx_median;
-- expected: an INT64 close to 5 or 6 (approximate median of 1..10).
```

```sql
-- The endpoints 0 and 1 yield the sketch's min and max.
WITH Points AS (
  SELECT x FROM UNNEST([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) AS x
), S AS (
  SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS sketch FROM Points
)
SELECT
  KLL_QUANTILES.EXTRACT_POINT_INT64(sketch, 0.0) AS approx_min,
  KLL_QUANTILES.EXTRACT_POINT_INT64(sketch, 1.0) AS approx_max
FROM S;
-- expected: approx_min = 1, approx_max = 10.
```

## Edge cases
- Returns `NULL` when `sketch` or `phi` is `NULL`.
- Errors when `phi` is outside the closed interval `[0, 1]`.
- Errors when `sketch` is not a KLL sketch over `INT64` (for example, a
  `FLOAT64` sketch must be read with `KLL_QUANTILES.EXTRACT_POINT_FLOAT64`
  instead).
- Errors when `sketch` is malformed `BYTES` not produced by a
  `KLL_QUANTILES.*` function.
- The result is approximate: the error is bounded by the precision used at
  sketch initialization, not by the size of the input data.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesextract_point_int64>
