---
name: KLL_QUANTILES.INIT_FLOAT64
dialect: bigquery
category: functions/kll
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesinit_double
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesinit_double
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/kll/kll_quantiles_init_float64.yaml
---

# KLL_QUANTILES.INIT_FLOAT64

## Summary

Aggregate function that initialises a KLL sketch from `FLOAT64`
input values. Returns the sketch as `BYTES`.

## Signatures

- `KLL_QUANTILES.INIT_FLOAT64(input [, precision [, weight => input_weight]])`

## Behavior

- Equivalent to `KLL_QUANTILES.INIT_INT64` but accepts `FLOAT64`
  input.
- Orders values using the GoogleSQL floating-point sort order
  (`NaN` precedes `-inf`).
- `precision` defaults to 1000; valid range is 1..100,000. Higher
  values trade memory for accuracy.
- `weight` is a per-input multiplier (1..2,147,483,647) — use named
  argument syntax `weight => N`.
- Returns the sketch encoded as `BYTES`.

## Examples

```sql
SELECT KLL_QUANTILES.INIT_FLOAT64(x, 1000) AS kll_sketch
FROM UNNEST([1.0, 2.0, 3.0, 4.0, 5.0]) AS x;
```

## Edge cases

- `precision` outside `[1, 100000]` raises an error.
- `weight` outside its allowed range raises an error.
- `NaN` is sorted before `-inf` in the resulting sketch.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/kll_functions#kll_quantilesinit_double>.
