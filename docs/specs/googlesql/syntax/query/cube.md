---
name: GROUP_BY_CUBE
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/query-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#group_by_cube
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/cube.yaml
---

# `GROUP BY CUBE`

## Summary

`CUBE(col1, ..., colN)` expands to every possible grouping subset of
its arguments (`2^N` grouping sets in total). Each subset produces its
own aggregate row.

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
