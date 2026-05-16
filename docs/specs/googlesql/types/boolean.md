---
name: BOOLEAN
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#boolean-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/boolean.yaml
---

# BOOLEAN

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

## Boolean type 
<a id="boolean_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
  <code>BOOL</code><br/>
  <code>BOOLEAN</code>
</td>
<td>Boolean values are represented by the keywords <code>TRUE</code> and
<code>FALSE</code> (case-insensitive).</td>
</tr>
</tbody>
</table>

`BOOLEAN` is an alias for `BOOL`.

Boolean values are sorted in this order, from least to greatest:

  1. `NULL`
  1. `FALSE`
  1. `TRUE`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
