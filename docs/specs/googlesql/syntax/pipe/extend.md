---
name: PIPE_EXTEND
dialect: googlesql
category: syntax/pipe
status: implemented
source_url: docs/third_party/googlesql-docs/pipe-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/pipe-syntax.md#extend_pipe_operator
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/pipe/extend.yaml
---

# `|> EXTEND`

## Summary

Propagates the existing table and adds computed columns. Similar to
`SELECT *, expression AS alias` in standard syntax.

## Signatures

- `|> EXTEND expression [[AS] alias] [, ...]`

## Behavior

- Output columns are the input columns followed by every newly extended
  column, in declaration order.

## Examples

```sql
FROM (SELECT 1 AS x UNION ALL SELECT 2 UNION ALL SELECT 3)
|> EXTEND x*2 AS doubled
|> ORDER BY x;
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/pipe-syntax.md`.
