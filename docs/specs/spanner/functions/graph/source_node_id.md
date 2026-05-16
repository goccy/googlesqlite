---
name: SOURCE_NODE_ID
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns the source-side foreign-key value of an edge as a STRING. Resolved via the edge-table metadata captured in property_graph.go (SrcFKCols). Implemented by formatPathScalarFunc / tryFormatGraphElementFunc in internal/graph_scan_node.go on top of the multi-hop GRAPH MATCH lowering.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#source_node_id
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#source_node_id
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/source_node_id.yaml
---

# SOURCE_NODE_ID

## Summary

Returns the identifier of the source (tail) node of a graph edge.

## Signatures

- `SOURCE_NODE_ID(edge)`

## Return type

The graphs node-identifier type.

## Behavior

- See `DESTINATION_NODE_ID` for the directionality model.
- Returns `NULL` if `edge` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH (a)-[e:TRANSFERS]->(b)
RETURN SOURCE_NODE_ID(e) AS from_id;
```

## Edge cases

- For self-loops, source and destination IDs are equal.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#source_node_id>.
