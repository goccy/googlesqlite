---
name: DOT_PRODUCT
dialect: spanner
category: functions/math
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#dot_product
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#dot_product
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/math/dot_product.yaml
---

# DOT_PRODUCT

## Summary

Returns the dot product of two equal-length numeric arrays.

## Signatures

- `DOT_PRODUCT(a, b)`

## Arguments

- `a`, `b`: `ARRAY<FLOAT32>` or `ARRAY<FLOAT64>` of equal length. Mixed element types raise an error.

## Return type

`FLOAT64`.

## Behavior

- Result is `SUM(a[i] * b[i])`.
- Returns `NULL` if either argument is `NULL` or contains a `NULL` element.
- Mismatched lengths raise an error.

## Examples

```sql
SELECT DOT_PRODUCT([1.0, 2.0, 3.0], [4.0, 5.0, 6.0]);   -- 32.0
```

## Edge cases

- For approximate (HW-accelerated) form, see `APPROX_DOT_PRODUCT`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/mathematical_functions#dot_product>.
