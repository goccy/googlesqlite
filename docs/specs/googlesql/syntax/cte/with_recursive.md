---
name: WITH_RECURSIVE
dialect: googlesql
category: syntax/cte
status: implemented
source_url: docs/third_party/googlesql-docs/recursive-ctes.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/recursive-ctes.md
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/cte/with_recursive.yaml
---

# `WITH RECURSIVE`

## Summary

`WITH RECURSIVE` defines a recursive common table expression: a
self-referential CTE that iterates a base query and a recursive query
until the recursive query produces no new rows.

## Signatures

- `WITH RECURSIVE cte_name AS (base_query UNION ALL recursive_query) SELECT ... FROM cte_name`

## Behavior

- The base query is evaluated once. Its rows seed the CTE.
- The recursive query is evaluated repeatedly, with the CTE name bound
  to the rows produced by the most recent iteration.
- Iteration terminates when the recursive query produces no new rows.
- The column list and types of the base query and the recursive query
  must align.

## Examples

```sql
WITH RECURSIVE numbers AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM numbers WHERE n < 5
)
SELECT n FROM numbers;
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/recursive-ctes.md`.
