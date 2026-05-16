---
name: PIPE_WHERE
dialect: googlesql
category: syntax/pipe
status: implemented
source_url: docs/third_party/googlesql-docs/pipe-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/pipe-syntax.md#where_pipe_operator
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/pipe/where.yaml
---

# `|> WHERE`

## Summary

Filters the preceding pipe step's rows by a boolean predicate.

## Signatures

- `|> WHERE bool_expression`

## Behavior

- Behaves like the standard `WHERE` clause but applies to whatever pipeline
  step produced the input table.
- Multiple `|> WHERE` operators chain as logical `AND`.

## Examples

```sql
FROM (SELECT 1 AS x UNION ALL SELECT 2 UNION ALL SELECT 3)
|> WHERE x > 1
|> SELECT x;
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/pipe-syntax.md`.
