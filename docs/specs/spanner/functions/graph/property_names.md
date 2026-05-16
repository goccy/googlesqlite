---
name: PROPERTY_NAMES
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns ARRAY<STRING> of property names declared on the element table.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#property_names
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#property_names
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/property_names.yaml
---

# PROPERTY_NAMES

## Summary

Returns the array of property names defined on a graph element.

## Signatures

- `PROPERTY_NAMES(element)`

## Return type

`ARRAY<STRING>`.

## Behavior

- Returns `NULL` if `element` is `NULL`.
- Order of property names is implementation-defined.

## Examples

```sql
GRAPH FinGraph
MATCH (n:Account)
RETURN PROPERTY_NAMES(n);
```

## Edge cases

- Use `n.<property>` to access values; this function only enumerates names.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#property_names>.
