---
name: LOAD_DATA
dialect: googlesql
category: syntax/io
status: implemented
source_url: docs/third_party/googlesql-docs/load-statements.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/load-statements.md
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/io/load_data.yaml
---

# `LOAD DATA`

## Summary

`LOAD DATA INTO <table> FROM FILES (...)` ingests rows from external files. Accepted-only in googlesqlite — actual ingestion is delegated to the emulator.

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

Apache 2.0 derivative of `docs/third_party/googlesql-docs/load-statements.md`.
