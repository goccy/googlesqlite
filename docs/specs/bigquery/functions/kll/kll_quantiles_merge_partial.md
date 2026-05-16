---
name: KLL_QUANTILES.MERGE_PARTIAL
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_partial
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_partial
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_merge_partial.yaml
---

# KLL_QUANTILES.MERGE_PARTIAL

## Summary

Aggregate function that merges KLL sketches of the same underlying
type into a new sketch (also `BYTES`), without extracting
quantiles.

## Signatures

- `KLL_QUANTILES.MERGE_PARTIAL(sketch)`

## Behavior

- All input sketches must share the same underlying data type
  (e.g. all `INT64`, or all `FLOAT64`).
- When sketches were initialised with different `precision`, the
  merged precision drops to the lowest input — unless the data is
  small enough for the higher precision to still be exact.
- `NULL` sketches are skipped. With zero rows or only `NULL`
  inputs, the function returns `NULL`.
- Returns the merged sketch as `BYTES`.

## Examples

```sql
SELECT KLL_QUANTILES.MERGE_PARTIAL(kll_sketch) AS merged_sketch
FROM (SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch
      FROM UNNEST([1, 2, 3, 4, 5]) AS x
      UNION ALL
      SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch
      FROM UNNEST([6, 7, 8, 9, 10]) AS x);
```

## Edge cases

- Mixing underlying types (e.g. `INT64` and `FLOAT64`) raises an
  error.
- Any input that isn't a valid KLL sketch raises an error.
- `NULL` inputs are ignored.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_partial>.
