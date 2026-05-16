---
name: APPROX_DOT_PRODUCT
dialect: spanner
category: functions/math
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_dot_product
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_dot_product
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/math/approx_dot_product.yaml
---

# APPROX_DOT_PRODUCT

## Summary

Approximate dot product. Like `DOT_PRODUCT` but performs the reduction in lower-precision arithmetic for speed.

## Signatures

- `APPROX_DOT_PRODUCT(a, b)`

## Return type

`FLOAT64`.

## Behavior

- See `DOT_PRODUCT` for argument shape; this variant trades a small amount of numerical accuracy for performance on vector workloads.
- Returns `NULL` if either argument is `NULL` or contains a `NULL` element.

## Examples

```sql
SELECT APPROX_DOT_PRODUCT([1.0, 2.0], [3.0, 4.0]);   -- ~11.0
```

## Edge cases

- Differences from `DOT_PRODUCT` are typically below `1e-3` for well-conditioned inputs.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#approx_dot_product>.
