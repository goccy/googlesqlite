---
name: PATH_LENGTH
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns the static edge count of the matched path (number of hop traversals). Implemented by formatPathScalarFunc / tryFormatGraphElementFunc in internal/graph_scan_node.go on top of the multi-hop GRAPH MATCH lowering.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_length
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_length
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/path_length.yaml
---

# PATH_LENGTH

## Summary

Returns the number of edges in a path.

## Signatures

- `PATH_LENGTH(path)`

## Return type

`INT64`.

## Behavior

- Equivalent to `ARRAY_LENGTH(EDGES(path))`.
- Returns `NULL` if `path` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH p = (a)-[*]-(b)
RETURN PATH_LENGTH(p);
```

## Edge cases

- A zero-length path has length `0`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_length>.
