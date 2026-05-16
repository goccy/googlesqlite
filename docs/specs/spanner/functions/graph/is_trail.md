---
name: IS_TRAIL
dialect: spanner
category: functions/graph
status: implemented
notes: |
  True when every edge alias in the path refers to a distinct rowid. Implemented by formatPathScalarFunc / tryFormatGraphElementFunc in internal/graph_scan_node.go on top of the multi-hop GRAPH MATCH lowering.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#is_trail
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#is_trail
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/is_trail.yaml
---

# IS_TRAIL

## Summary

Returns `TRUE` if a path uses no edge more than once (nodes may repeat).

## Signatures

- `IS_TRAIL(path)`

## Return type

`BOOL`.

## Behavior

- Stricter than acyclicity in edge usage but more permissive in node revisits.
- Returns `NULL` if `path` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH p = (a)-[*1..5]-(b)
WHERE IS_TRAIL(p)
RETURN p;
```

## Edge cases

- A zero-length path is trivially a trail.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#is_trail>.
