---
name: QUALIFY
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/query-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#qualify_clause
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/qualify.yaml
---

# `QUALIFY`

## Summary

`QUALIFY` filters rows by the result of a window function call. It runs
after window functions are computed but before `ORDER BY`/`LIMIT`,
giving an easier way to express "top-N per partition" than a
subquery+`WHERE`.

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
