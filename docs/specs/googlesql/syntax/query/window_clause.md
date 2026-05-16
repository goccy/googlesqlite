---
name: WINDOW_CLAUSE
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/query-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md#def_use_named_window
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/window_clause.yaml
---

# Named `WINDOW` clause

## Summary

The trailing `WINDOW name AS (window_spec)` clause defines named
windows that can be referenced from window-function calls in the
`SELECT` list. Identical to inline `OVER (...)` but lets callers reuse
the spec.

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
