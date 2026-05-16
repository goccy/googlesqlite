---
name: KLL_QUANTILES.MERGE_INT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_int64
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_int64
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_merge_int64.yaml
---

# KLL_QUANTILES.MERGE_INT64

## Summary

Aggregate function that merges multiple `INT64`-initialised KLL
sketches and returns approximate quantile boundaries.

## Signatures

- `KLL_QUANTILES.MERGE_INT64(sketch, num_quantiles)`

## Behavior

- `sketch`: `BYTES` KLL sketch initialised on `INT64` data.
- `num_quantiles`: positive `INT64` (max 100,000) — the number of
  roughly equal-sized groups the merged distribution is split into.
- Returns `ARRAY<INT64>` of length `num_quantiles + 1`: exact min,
  the approximate quantile boundaries, and exact max.
- When merging sketches initialised with different `precision`
  values the merged precision drops to the lowest input precision —
  except when the merged data is small enough to be captured
  exactly, in which case the higher precision is preserved.

## Examples

```sql
SELECT KLL_QUANTILES.MERGE_INT64(kll_sketch, 2) AS halves
FROM (SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch
      FROM UNNEST([1, 2, 3, 4, 5]) AS x
      UNION ALL
      SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch
      FROM UNNEST([6, 7, 8, 9, 10]) AS x);
-- expected: [1, 5, 10]
```

## Edge cases

- Returns an error when the underlying type of any input sketch is
  not `INT64`.
- Returns an error when an input is not a valid KLL sketch.
- `num_quantiles > 100000` raises an error.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesmerge_int64>.
