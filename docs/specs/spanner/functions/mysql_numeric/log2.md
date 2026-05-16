---
name: LOG2
dialect: spanner
category: functions/mysql_numeric
status: implemented
notes: |
  Returns base-2 logarithm. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_log2 in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#log2
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#log2
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_numeric/log2.yaml
---

# LOG2

## Summary

Returns the base-2 logarithm of `x`.

## Signatures

- `LOG2(x)`

## Arguments

- `x`: `FLOAT64`.

## Return type

`FLOAT64`.

## Behavior

- `LOG2(x) = LN(x) / LN(2)`.
- `x <= 0` raises an error (or returns `NULL` in SAFE mode).
- Returns `NULL` if `x` is `NULL`.

## Examples

```sql
SELECT LOG2(8);    -- 3.0
SELECT LOG2(1);    -- 0.0
```

## Edge cases

- For arbitrary bases, use `LOG(x, b)` (note argument order).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/numeric_functions#log2>.
