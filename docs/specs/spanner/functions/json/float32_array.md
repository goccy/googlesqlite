---
name: FLOAT32_ARRAY
dialect: spanner
category: functions/json
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float32_array
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float32_array
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/json/float32_array.yaml
---

# FLOAT32_ARRAY

## Summary

Coerces a `JSON` array to `ARRAY<FLOAT32>`.

## Signatures

- `FLOAT32_ARRAY(json_expr)`
- `FLOAT32_ARRAY(json_expr, wide_number_mode)`

## Arguments

- `json_expr`: `JSON` containing an array of numbers.
- `wide_number_mode`: optional `STRING` (`"exact"` or `"round"`).

## Return type

`ARRAY<FLOAT32>`.

## Behavior

- Each element must be coercible to `FLOAT32` per `FLOAT32`s rules.
- A `JSON null` element becomes a SQL `NULL` element.

## Examples

```sql
SELECT FLOAT32_ARRAY(JSON [1.5, 2.25, -3.75]);   -- [1.5e0, 2.25e0, -3.75e0]
```

## Edge cases

- Heterogeneous arrays (mixed numeric and non-numeric) raise an error.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float32_array>.
