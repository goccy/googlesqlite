---
name: IS_ACYCLIC
dialect: spanner
category: functions/graph
status: implemented
notes: |
  True when every node alias in the path refers to a distinct rowid. Implemented by formatPathScalarFunc / tryFormatGraphElementFunc in internal/graph_scan_node.go on top of the multi-hop GRAPH MATCH lowering.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#is_acyclic
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#is_acyclic
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/is_acyclic.yaml
---

# IS_ACYCLIC

## Summary

Returns `TRUE` if a path visits no node more than once.

## Signatures

- `IS_ACYCLIC(path)`

## Return type

`BOOL`.

## Behavior

- Returns `NULL` if `path` is `NULL`.
- Useful as a `WHERE` filter to suppress cyclic walks in unbounded `*`-quantified patterns.

## Examples

```sql
GRAPH FinGraph
MATCH p = (a)-[*1..5]-(b)
WHERE IS_ACYCLIC(p)
RETURN p;
```

## Edge cases

- A zero-length path is acyclic by definition.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#is_acyclic>.
