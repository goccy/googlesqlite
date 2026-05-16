---
name: EDGES
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns an ARRAY of edge ELEMENT_IDs in path order, built via googlesqlite_make_array over the static edge-alias list.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#edges
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#edges
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/edges.yaml
---

# EDGES

## Summary

Returns the array of edges in a graph path, in path order.

## Signatures

- `EDGES(path)`

## Return type

`ARRAY<EDGE>`.

## Behavior

- An n-length path has `n-1` edges (and `n` nodes).
- Returns `NULL` if `path` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH p = (a)-[*]-(b)
RETURN EDGES(p);
```

## Edge cases

- For zero-length paths, the array is empty.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#edges>.
