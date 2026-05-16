---
name: ARRAY_SUBSCRIPT
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/query-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#array_subscript_operator
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/array_subscript.yaml
---

# Array subscript operators

## Summary

Index into an array literal or array-typed value with one of four
modes:

- `OFFSET(n)` — 0-based; out-of-range raises.
- `ORDINAL(n)` — 1-based; out-of-range raises.
- `SAFE_OFFSET(n)` — 0-based; out-of-range returns `NULL`.
- `SAFE_ORDINAL(n)` — 1-based; out-of-range returns `NULL`.

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
