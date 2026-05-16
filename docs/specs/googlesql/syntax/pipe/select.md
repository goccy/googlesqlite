---
name: PIPE_SELECT
dialect: googlesql
category: syntax/pipe
status: implemented
source_url: docs/third_party/googlesql-docs/pipe-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/pipe-syntax.md#select_pipe_operator
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/pipe/select.yaml
---

# `|> SELECT`

## Summary

Pipe `SELECT` produces a new table with the listed columns. Similar to the
outermost `SELECT` clause in a standard subquery.

## Signatures

- `|> SELECT expression [[AS] alias] [, ...]`

## Behavior

- Operates on the table produced by the preceding pipe step.
- Does not perform aggregation; use the `AGGREGATE` operator instead.
- Supports window-function expressions, `SELECT AS STRUCT`, and `DISTINCT`.

## Examples

```sql
FROM (SELECT 'apples' AS item, 2 AS sales)
|> SELECT item AS fruit_name;
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/pipe-syntax.md`.
