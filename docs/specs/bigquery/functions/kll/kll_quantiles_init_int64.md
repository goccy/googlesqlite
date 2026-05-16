---
name: KLL_QUANTILES.INIT_INT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesinit_int64
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesinit_int64
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_init_int64.yaml
---

# KLL_QUANTILES.INIT_INT64

## Summary

Aggregate function that builds a KLL sketch from `INT64` input
values. Returns the sketch as `BYTES`.

## Signatures

- `KLL_QUANTILES.INIT_INT64(input [, precision [, weight => input_weight]])`

## Behavior

- `input`: `INT64` value being aggregated into the sketch.
- `precision`: `INT64` controlling sketch accuracy. Default 1000;
  valid range 1..100,000. Larger values capture quantiles more
  precisely at the cost of more memory.
- `input_weight` (named argument `weight`): `INT64` multiplier for
  each input (default 1, max 2,147,483,647). A weight of 3 makes
  one input row count as if three identical rows had been seen.
- Returns the sketch as `BYTES`. Pass to `EXTRACT_INT64` /
  `EXTRACT_POINT_INT64` / `MERGE_INT64` / `MERGE_POINT_INT64` to
  query approximate quantiles.

## Examples

```sql
SELECT KLL_QUANTILES.INIT_INT64(x, 1000) AS kll_sketch
FROM UNNEST([1, 2, 3, 4, 5]) AS x;
-- yields a BYTES sketch consumable by EXTRACT_*/MERGE_*
```

## Edge cases

- `precision` outside `[1, 100000]` raises an error.
- `input_weight` outside its allowed range raises an error.
- An empty input set yields a `NULL` sketch.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesinit_int64>.
