---
name: DESTINATION_NODE_ID
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns the destination-side foreign-key value of an edge as a STRING. Implemented by formatPathScalarFunc / tryFormatGraphElementFunc in internal/graph_scan_node.go on top of the multi-hop GRAPH MATCH lowering.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#destination_node_id
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#destination_node_id
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/destination_node_id.yaml
---

# DESTINATION_NODE_ID

## Summary

Returns the identifier of the destination (head) node of a graph edge.

## Signatures

- `DESTINATION_NODE_ID(edge)`

## Arguments

- `edge`: a graph edge value bound in a `MATCH` graph pattern.

## Return type

The graphs node-identifier type (typically the primary-key tuple).

## Behavior

- For an undirected edge encountered in a directed traversal, the destination is the endpoint reached via the traversal.
- Returns `NULL` if `edge` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH (a)-[e:TRANSFERS]->(b)
RETURN DESTINATION_NODE_ID(e) AS to_id;
```

## Edge cases

- Companion of `SOURCE_NODE_ID`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#destination_node_id>.
