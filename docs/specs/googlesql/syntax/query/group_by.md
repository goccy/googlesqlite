---
name: GROUP_BY
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/query-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#group_by_clause
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/group_by.yaml
---

# `GROUP BY`

## Summary

`GROUP BY` partitions a query's rows into groups by one or more
expressions. Each aggregation in the SELECT list collapses each group
to one row. Group keys include scalar columns, struct field
references, and full struct / array values.

## Signatures

See the upstream reference linked at the bottom of this spec.

## Behavior

See the upstream reference linked at the bottom of this spec.

## Examples

See the upstream reference linked at the bottom of this spec and the testdata YAML.

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/query-syntax.md`.
