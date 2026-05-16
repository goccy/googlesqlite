---
name: PIPE_AGGREGATE
dialect: googlesql
category: syntax/pipe
status: implemented
source_url: docs/third_party/googlesql-docs/pipe-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/pipe-syntax.md#aggregate_pipe_operator
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/pipe/aggregate.yaml
---

# `|> AGGREGATE`

## Summary

Aggregates rows from the preceding pipe step. Replaces the standard
`SELECT ... GROUP BY` form.

## Signatures

- `|> AGGREGATE aggregate_expression [[AS] alias] [, ...]`
- `|> AGGREGATE aggregate_expression [, ...] GROUP BY group_expression [, ...]`

## Behavior

- Produces one row per group (or one row total if no `GROUP BY` is given).
- Aggregate expressions may use any standard aggregate function.

## Examples

```sql
FROM (SELECT 1 AS x UNION ALL SELECT 2 UNION ALL SELECT 3)
|> AGGREGATE COUNT(*) AS n;
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/pipe-syntax.md`.
