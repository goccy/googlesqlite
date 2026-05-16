---
name: EXPORT_DATA
dialect: googlesql
category: syntax/io
status: implemented
source_url: docs/third_party/googlesql-docs/export-statements.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/export-statements.md
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/io/export_data.yaml
---

# `EXPORT DATA`

## Summary

`EXPORT DATA OPTIONS(...) AS SELECT ...` writes the inner query to an external destination. googlesqlite resolves the inner SELECT and returns its rows; the actual write is left to the emulator layered above.

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

Apache 2.0 derivative of `docs/third_party/googlesql-docs/export-statements.md`.
