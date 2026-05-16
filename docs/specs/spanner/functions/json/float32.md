---
name: FLOAT32
dialect: spanner
category: functions/json
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float32
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float32
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/json/float32.yaml
---

# FLOAT32

## Summary

Coerces a `JSON` value to `FLOAT32`. Spanner-specific variant complementing the GoogleSQL `FLOAT64`/`INT64`/etc. extractors.

## Signatures

- `FLOAT32(json_expr)`
- `FLOAT32(json_expr, wide_number_mode)`

## Arguments

- `json_expr`: `JSON`.
- `wide_number_mode`: optional `STRING` `"exact"` (the default — error on out-of-range) or `"round"` (round to nearest representable `FLOAT32`).

## Return type

`FLOAT32`.

## Behavior

- Numeric JSON values are coerced; other JSON kinds raise an error.
- Returns `NULL` if `json_expr` is `JSON NULL` or SQL `NULL`.

## Examples

```sql
SELECT FLOAT32(JSON 3.14);   -- 3.14e0
```

## Edge cases

- Use `FLOAT64` for double-precision; `FLOAT32` is a memory-efficient choice for vector workloads.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/json_functions#float32>.
