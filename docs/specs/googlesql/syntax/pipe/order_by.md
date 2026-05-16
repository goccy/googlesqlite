---
name: PIPE_ORDER_BY
dialect: googlesql
category: syntax/pipe
status: implemented
source_url: docs/third_party/googlesql-docs/pipe-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/pipe-syntax.md#order_by_pipe_operator
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/pipe/order_by.yaml
---

# `|> ORDER BY`

## Summary

Sorts the rows of the preceding pipe step.

## Signatures

- `|> ORDER BY expression [ASC|DESC] [, ...]`

## Examples

```sql
FROM (SELECT 3 AS x UNION ALL SELECT 1 UNION ALL SELECT 2)
|> ORDER BY x;
```

## Behavior

See the upstream reference linked at the bottom of this spec.

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/pipe-syntax.md`.
