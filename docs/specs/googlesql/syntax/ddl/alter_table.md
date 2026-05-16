---
name: ALTER_TABLE
dialect: googlesql
category: syntax/ddl
status: implemented
source_url: docs/third_party/googlesql-docs/data-definition-language.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-definition-language.md
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/ddl/alter_table.yaml
---

# `ALTER TABLE`

## Summary

Schema-level changes to an existing table — typically `ADD COLUMN` to grow the column list. googlesqlite accepts the statement for analyzer compatibility; the catalog is treated as metadata-only.

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

Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-definition-language.md`.
