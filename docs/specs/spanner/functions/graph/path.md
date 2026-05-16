---
name: PATH
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Path construction syntax `MATCH p = (a)-[k]->(b)` is recognised by the analyzer once the SqlGraphPathType / SqlGraphPathMode language features are enabled, and graph_scan_node.go lowers it to a JOIN chain plus a per-path pathInfo map. The PATH() builtin itself is implicitly created by the analyzer; users reference the path column directly.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/path.yaml
---

# PATH

## Summary

Constructs a graph path value from explicit node/edge sequences.

## Signatures

- `PATH(node_or_edge_1, node_or_edge_2, ...)`

## Return type

A graph `PATH` value.

## Behavior

- Used inside graph queries to assemble path values when the natural pattern syntax is insufficient (e.g. dynamic-length traversals).
- Returns `NULL` if any input is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH (a:Account), (b:Account)
RETURN PATH(a, edge_to_b, b);
```

## Edge cases

- Inputs must alternate node/edge/node/...; mismatches raise an error.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path>.
