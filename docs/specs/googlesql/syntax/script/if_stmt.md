---
name: IF
dialect: googlesql
category: syntax/script
status: implemented
source_url: docs/third_party/googlesql-docs/procedural-language.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/procedural-language.md
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/script/if_stmt.yaml
---

# `IF` statement

## Summary

Conditional execution: `IF condition THEN ... [ELSEIF ... THEN ...] [ELSE ...] END IF`. Each THEN/ELSEIF branch is a script-level statement list.

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

Apache 2.0 derivative of `docs/third_party/googlesql-docs/procedural-language.md`.
