---
name: KLL_QUANTILES.MERGE_FLOAT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_float64
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_float64
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_merge_float64.yaml
---

# KLL_QUANTILES.MERGE_FLOAT64

## Summary
Aggregate function that merges KLL quantiles sketches initialized on
`FLOAT64` data and returns an `ARRAY<FLOAT64>` containing the exact minimum,
the requested approximate quantiles, and the exact maximum of the combined
input. Behaves like `KLL_QUANTILES.MERGE_INT64` but for `FLOAT64`-typed
sketches.

## Signatures
- `KLL_QUANTILES.MERGE_FLOAT64(sketch, num_quantiles)`

## Behavior
- `sketch` is a `BYTES` KLL quantiles sketch initialized on the `FLOAT64`
  data type (e.g. produced by `KLL_QUANTILES.INIT_FLOAT64` or
  `KLL_QUANTILES.MERGE_PARTIAL` of `FLOAT64` sketches).
- Aggregates by merging all input sketches into a single sketch, then
  returns `num_quantiles` approximate quantiles bracketed by the exact
  observed minimum and maximum of the merged input.
- The result is an `ARRAY<FLOAT64>` of length `num_quantiles + 1`:
  the exact minimum of the merged input, each approximate quantile in
  ascending order, then the exact maximum.
- `num_quantiles` is an `INT64` specifying how many roughly equal-sized
  groups the captured input is divided into; the corresponding
  `KLL_QUANTILES.MERGE_INT64` documents a maximum value of 100,000.
- Values are ordered using the GoogleSQL floating-point sort order, so
  `NaN` orders before negative infinity.
- If the merged sketches were initialized with different precisions, the
  output precision is downgraded to the lowest precision involved in the
  merge — except when the aggregations are small enough to still capture
  the input exactly, in which case the mergee's precision is maintained.
- Returns an error if any input sketch's underlying type is not compatible
  with `FLOAT64`, or if any input is not a valid KLL quantiles sketch.

## Examples
```sql
SELECT KLL_QUANTILES.MERGE_FLOAT64(kll_sketch, 2) AS halves
FROM (SELECT KLL_QUANTILES.INIT_FLOAT64(x, 1000) AS kll_sketch
      FROM (SELECT 1.0 AS x UNION ALL
            SELECT 2.0 UNION ALL
            SELECT 3.0 UNION ALL
            SELECT 4.0 UNION ALL
            SELECT 5.0)
      UNION ALL
      SELECT KLL_QUANTILES.INIT_FLOAT64(x, 1000) AS kll_sketch
      FROM (SELECT 6.0 AS x UNION ALL
            SELECT 7.0 UNION ALL
            SELECT 8.0 UNION ALL
            SELECT 9.0 UNION ALL
            SELECT 10.0));
-- expected: halves = [1.0, 5.0, 10.0]
```

## Edge cases
- Returns an error if any input sketch's underlying type is not compatible
  with `FLOAT64` (e.g. mixing an `INT64`-initialized sketch with a
  `FLOAT64`-initialized one).
- Returns an error if any input is not a valid KLL quantiles sketch.
- `NaN` sorts before `-inf` in the floating-point ordering used by this
  function, which affects which value appears as the reported minimum.
- Mixing precisions across input sketches downgrades the merged precision
  to the lowest involved, unless the merged input is small enough to be
  captured exactly.
- Quantile values in the output are approximate; only the leading minimum
  and trailing maximum are exact extremes observed during sketch
  initialization.
- As an aggregate, the function follows the standard NULL handling for
  KLL aggregates: `NULL` sketches are ignored (see
  `KLL_QUANTILES.MERGE_PARTIAL` for the documented rule on the same
  family).

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_float64>.
