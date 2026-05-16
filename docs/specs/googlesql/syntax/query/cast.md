---
name: CAST
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/conversion_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/conversion_functions.md
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/cast.yaml
---

# `CAST` and `SAFE_CAST`

## Summary

`CAST(x AS T)` converts `x` to type `T`, raising if the cast fails.
`SAFE_CAST(x AS T)` returns `NULL` instead of an error when the cast
fails. Both forms accept the same expression-and-type pair.

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

Apache 2.0 derivative of `docs/third_party/googlesql-docs/conversion_functions.md`.
