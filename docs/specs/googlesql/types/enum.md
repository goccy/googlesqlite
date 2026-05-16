---
name: ENUM
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#enum-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/enum.yaml
---

# ENUM

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

## Enum type 
<a id="enum_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>ENUM</code></td>
<td>Named type that maps string constants to <code>INT32</code> constants.</td>
</tr>
</tbody>
</table>

An enum is a named type that enumerates a list of possible values, each of which
contains:

+ An integer value: Integers are used for comparison and ordering enum values.
There is no requirement that these integers start at zero or that they be
contiguous.
+ A string value for its name: Strings are case sensitive. In the case of
protocol buffer open enums, this name is optional.
+ Optional alias values: One or more additional string values that act as
aliases.

Enum values are referenced using their integer value or their string value.
You reference an enum type, such as when using CAST, by using its fully
qualified name.

You can't create new enum types using GoogleSQL.

To learn more about the literal representation of an enum type,
see [Enum literals][enum-literals].

[enum-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#enum_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
