---
name: PATH_LAST
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns the ELEMENT_ID-shape string of the path's last node.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_last
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_last
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/path_last.yaml
---

# PATH_LAST

## Summary

Returns the last node of a path.

## Signatures

- `PATH_LAST(path)`

## Return type

`NODE`.

## Behavior

- Equivalent to `NODES(path)[ORDINAL(ARRAY_LENGTH(NODES(path)))]`.
- Returns `NULL` if `path` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH p = (a)-[*]-(b)
RETURN PATH_LAST(p);
```

## Edge cases

- Defined for zero-length paths (returns the singleton node).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#path_last>.
