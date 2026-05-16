---
name: STRUCT_ARRAY_LITERALS
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/struct_array_literals.yaml
---

# `STRUCT` and `ARRAY` literals

## Summary

Inline `STRUCT(...)` and bracket `[...]` constructors materialise
ad-hoc composite values inside an expression. Both forms support
declared types, named fields, and arbitrary expressions for each
element.

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

Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
