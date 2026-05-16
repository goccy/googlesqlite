---
name: PATH_FIRST
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns the ELEMENT_ID-shape string of the path's first node. ELEMENT_ID(PATH_FIRST(p)) round-trips through formatPathScalarFunc.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_first
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_first
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/path_first.yaml
---

# PATH_FIRST

## Summary

Returns the first node of a path.

## Signatures

- `PATH_FIRST(path)`

## Return type

`NODE`.

## Behavior

- Equivalent to `NODES(path)[OFFSET(0)]`.
- Returns `NULL` if `path` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH p = (a)-[*]-(b)
RETURN PATH_FIRST(p);
```

## Edge cases

- Defined for zero-length paths (returns the singleton node).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_first>.
