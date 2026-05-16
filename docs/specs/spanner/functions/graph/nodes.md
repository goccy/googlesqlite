---
name: NODES
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns an ARRAY of node ELEMENT_IDs in path order, built via googlesqlite_make_array over the static node-alias list captured at GraphPathScan lowering.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#nodes
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#nodes
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/nodes.yaml
---

# NODES

## Summary

Returns the array of nodes in a graph path, in path order.

## Signatures

- `NODES(path)`

## Return type

`ARRAY<NODE>`.

## Behavior

- The first element is the path start node; the last element is the path end node.
- Returns `NULL` if `path` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH p = (a)-[*1..3]-(b)
RETURN NODES(p);
```

## Edge cases

- For zero-length paths (a single node), the array has one element.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#nodes>.
