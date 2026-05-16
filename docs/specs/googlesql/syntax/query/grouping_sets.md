---
name: GROUPING_SETS
dialect: googlesql
category: syntax/query
status: implemented
source_url: docs/third_party/googlesql-docs/query-syntax.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#group_by_grouping_sets
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/grouping_sets.yaml
---

# `GROUP BY GROUPING SETS`

## Summary

`GROUPING SETS` lets a query compute multiple groupings in a single
pass. Each named group set produces one row per (group_set, distinct
key combination). The empty set `()` produces a single grand-total row.

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
