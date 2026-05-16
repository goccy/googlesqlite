---
name: FLOAT64_ARRAY
dialect: spanner
category: functions/json
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float64_array
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float64_array
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/json/float64_array.yaml
---

# FLOAT64_ARRAY

## Summary

Coerces a `JSON` array to `ARRAY<FLOAT64>`.

## Signatures

- `FLOAT64_ARRAY(json_expr)`
- `FLOAT64_ARRAY(json_expr, wide_number_mode)`

## Arguments

- `json_expr`: `JSON` containing an array of numbers.
- `wide_number_mode`: optional `STRING` (`"exact"` or `"round"`).

## Return type

`ARRAY<FLOAT64>`.

## Behavior

- Each element must be coercible to `FLOAT64`.
- A `JSON null` element becomes a SQL `NULL` element.

## Examples

```sql
SELECT FLOAT64_ARRAY(JSON [1.5, 2.25]);   -- [1.5e0, 2.25e0]
```

## Edge cases

- For single-precision storage, use `FLOAT32_ARRAY`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float64_array>.
