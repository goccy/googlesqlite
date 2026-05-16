---
name: ELEMENT_DEFINITION_NAME
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns the element table's alias as a STRING literal (e.g. `'P'` for `NODE TABLES (Persons AS P ...)`).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#element_definition_name
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#element_definition_name
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/element_definition_name.yaml
---

# ELEMENT_DEFINITION_NAME

## Summary

Returns the schema definition name (label) of a graph element.

## Signatures

- `ELEMENT_DEFINITION_NAME(element)`

## Return type

`STRING`.

## Behavior

- For nodes the result is the node label; for edges it is the edge label.
- Returns `NULL` if `element` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH (n)
RETURN ELEMENT_DEFINITION_NAME(n);
```

## Edge cases

- An element with multiple labels returns the primary label only; use `LABELS` for the full set.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#element_definition_name>.
