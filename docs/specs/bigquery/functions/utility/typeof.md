---
name: TYPEOF
dialect: bigquery
category: functions/utility
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/utility-functions#typeof
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/utility-functions#typeof
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/utility/typeof.yaml
---

# TYPEOF

## Summary

Returns the GoogleSQL type name of an expression as a `STRING`.

## Signatures

- `TYPEOF(expression)`

## Behavior

- Return type is `STRING`.
- For `NULL`, the supertype `INT64` is returned.
- Reports parameterised types in their declared form (e.g.
  `STRUCT<x INT64, y STRING>`).
- Field access returns the type of the resolved field
  (e.g. `TYPEOF(s.y)` for a struct field).

## Examples

```sql
SELECT TYPEOF(NULL), TYPEOF('hello'), TYPEOF(12+1), TYPEOF(4.7);
-- expected: INT64, STRING, INT64, FLOAT64
```

## Edge cases

- `NULL` collapses to the supertype `INT64` rather than reporting
  `NULL`-typed.
- Implicit-cast literals report the resolved type, not the source
  literal type.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/utility-functions#typeof>.
