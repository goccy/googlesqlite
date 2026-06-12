---
name: MERGE
dialect: googlesql
category: syntax/dml
status: implemented
source_url: docs/third_party/googlesql-docs/data-manipulation-language.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-manipulation-language.md
last_synced: 2026-06-12
testdata: testdata/specs/googlesql/syntax/dml/merge.yaml
---

# `MERGE`

## Summary

Merges rows from a source table or subquery into a target table by joining
them under a boolean `merge_condition` and applying per-row `INSERT`,
`UPDATE`, or `DELETE` actions selected by `WHEN MATCHED`,
`WHEN NOT MATCHED BY TARGET`, and `WHEN NOT MATCHED BY SOURCE` clauses.

## Signatures

See the upstream reference linked at the bottom of this spec.

## Behavior

See the upstream reference linked at the bottom of this spec. Two
documented behaviours that the spec tests pin:

- The canonical `Inventory` + `NewArrivals` walkthrough (`ON T.product
  = S.product`) — the post-MERGE row set in the upstream "after" table
  is what the spec test asserts.
- `ON FALSE` as a documented constant-false-predicate optimisation:
  paired with `WHEN NOT MATCHED BY SOURCE THEN DELETE` +
  `WHEN NOT MATCHED BY TARGET THEN INSERT` it implements REPLACE
  semantics (`DELETE` followed by `INSERT`).

The implementation accepts any boolean expression as the
`merge_condition`; the spec tests cover equality, multi-equality,
non-equality (`>`), and the `FALSE` / `TRUE` degenerate cases.

## Examples

See the upstream reference linked at the bottom of this spec and the
testdata YAML.

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-manipulation-language.md`.
