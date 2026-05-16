---
name: LABELS
dialect: spanner
category: functions/graph
status: implemented
notes: |
  Returns ARRAY<STRING> of label names declared on the element table. tryFormatGraphElementFunc in internal/graph_scan_node.go reads the metadata that property_graph.go snapshotted at SimplePropertyGraph construction time and emits a base64-encoded ARRAY<STRING> literal.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#labels
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#labels
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/graph/labels.yaml
---

# LABELS

## Summary

Returns the array of labels associated with a graph element.

## Signatures

- `LABELS(element)`

## Return type

`ARRAY<STRING>`.

## Behavior

- Returns `NULL` if `element` is `NULL`.

## Examples

```sql
GRAPH FinGraph
MATCH (n)
RETURN LABELS(n);
```

## Edge cases

- Order of labels is implementation-defined.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/graph_functions#labels>.
