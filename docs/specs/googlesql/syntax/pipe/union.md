---
name: PIPE_UNION
dialect: googlesql
category: syntax/pipe
status: implemented
source_url: docs/third_party/googlesql-docs/pipe-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/pipe-syntax.md#union_pipe_operator
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/pipe/union.yaml
---

# `|> UNION`

## Summary

Combines the preceding pipe step's rows with the rows of one or more
input queries. `UNION ALL` retains duplicates; `UNION DISTINCT` removes
them.

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

Apache 2.0 derivative of `docs/third_party/googlesql-docs/pipe-syntax.md`.
