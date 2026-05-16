---
name: ELEMENT_ID
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns the row-unique ID derived from the underlying SQLite rowid (`CAST(alias.rowid AS STRING)`).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#element_id
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#element_id
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/element_id.yaml
---

# ELEMENT_ID

## Summary

Returns the identifier of a node or edge graph element.

## Signatures

- `ELEMENT_ID(element)`

## Return type

The graphs element-identifier type.

## Behavior

- Returns `NULL` if `element` is `NULL`.
- Stable across snapshots for the same logical element.

## Examples

```sql
GRAPH FinGraph
MATCH (n:Account)
RETURN ELEMENT_ID(n);
```

## Edge cases

- Element IDs are opaque; do not rely on their byte layout.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#element_id>.
