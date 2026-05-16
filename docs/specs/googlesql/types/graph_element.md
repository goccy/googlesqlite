---
name: GRAPH ELEMENT
dialect: googlesql
category: types
status: partial
notes: |
  TYPEOF(graph_element) is the blocker. The upstream Example asserts the
  exact rendering `GRAPH_NODE(FinGraph)<Id INT64, ..., DYNAMIC>`, which
  needs two pieces that the current pipeline does not produce:
    1. The graph-scan lowering in internal/graph_scan_node.go does not
       materialise the bound element variable `n` as a column at the
       SQL layer — only `n.<property>` and `ELEMENT_ID(n)` are
       reachable. The formatter therefore emits `typeof(n#1)` against
       a column SQLite cannot resolve ("no such column: n#1").
    2. The analyzer's TypeofFunction rewrite folds TYPEOF down to a
       string literal for ordinary types, but does not render
       `GRAPH_NODE(name)<properties..., DYNAMIC>` for graph-element
       types. A formatter-level special case would need to construct
       that exact string from the resolved graph-table schema.
  Closing this needs either (a) materialise the element variable as a
  structural column the formatter can hand to TYPEOF, or (b) a custom
  resolved-tree pass that folds `TYPEOF(graph_element_column_ref)` to
  the rendered string at compile time. Both are go-googlesql-level work.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#graph-element-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/graph_element.yaml
---

# GRAPH ELEMENT

## Summary

(TBD — refine from the upstream reference below.)

## Signatures

(TBD)

## Behavior

(TBD)

## Examples

(TBD)

## Edge cases

(TBD)

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## Graph element type 
<a id="graph_element_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>GRAPH_ELEMENT</code></td>
<td>
  An element in a property graph.
</td>
</tr>
</tbody>
</table>

A variable with a `GRAPH_ELEMENT` type is produced by a graph query.
The generated type has this format:

```
GRAPH_ELEMENT<T>
```

A graph element is either a node or an edge, representing data from a
matching node or edge table based on its label. Each graph element holds a
set of properties that can be accessed with a case-insensitive name,
similar to fields of a struct.

**Example**

In the following example, `n` represents a graph element in the
[`FinGraph`][fin-graph] property graph:

```googlesql
GRAPH FinGraph
MATCH (n:Person)
RETURN n.name
```

In the following example, the [`TYPEOF`][type-of] function is used to inspect the
set of properties defined in the graph element type.

```googlesql
GRAPH FinGraph
MATCH (n:Person)
RETURN TYPEOF(n) AS t
LIMIT 1

/*----------------------------------------------+
 | t                                            |
 +----------------------------------------------+
 | GRAPH_NODE(FinGraph)<Id INT64, ..., DYNAMIC> |
 +---------------------------------------------*/
```

[graph-query]: https://github.com/google/googlesql/blob/master/docs/graph-intro.md

[fin-graph]: https://github.com/google/googlesql/blob/master/docs/graph-schema-statements.md#fin_graph

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
