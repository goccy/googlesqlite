---
name: UUID
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#uuid-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/uuid.yaml
---

# UUID

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

## UUID type 
<a id="uuid_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>UUID</code></td>
<td>A universally unique identifier (UUID) represented as a 128-bit number.</td>
</tr>
</tbody>
</table>

The following ASCII string format of lowercase hexadecimal digits is used to
represent a UUID:

`[8 digits]-[4 digits]-[4 digits]-[4 digits]-[12 digits]`

**Example**

`f81d4fae-7dec-11d0-a765-00a0c91e6bf6`

### Cast a UUID to a string

You can cast a UUID to a string by using the following syntax:

```googlesql
  SELECT CAST(NEW_UUID() AS STRING) AS UUID_STR;
```

You can also cast a string to a UUID, either explicitly or by using an
implicit coercion of a literal or parameter.

**Examples**

```googlesql
  SELECT UUID_id >= CAST("00000000-0000-0000-0000-000000000000" AS UUID) FROM T1;
```

```googlesql
  SELECT UUID_id >= "00000000-0000-0000-0000-000000000000" FROM T1;
```

### Cast a UUID to bytes

You can cast a UUID to bytes by using the following syntax:

```googlesql
  SELECT CAST(NEW_UUID() AS BYTES) AS UUID_BYTES;
```

You can also explicitly cast bytes to a UUID. Unlike strings, bytes can't be
implicitly coerced to a UUID.

##### Comparison operator examples

The comparison operator compares UUIDs using their internal representation.
However, the result is presented as if the comparison were performed on the
36-character lowercase ASCII string representation of the UUIDs,
using lexicographical order.

<table>
<thead>
<tr>
<th>Left term</th>
<th>Operator</th>
<th>Right term</th>
<th>Returns</th>
</tr>
</thead>
<tbody>
<tr>
<td>Any value</td>
<td><code>=</code></td>
<td><code>NULL</code></td>
<td><code>NULL</code></td>
</tr>
<tr>
<td><code>NULL</code></td>
<td><code>&lt;</code></td>
<td>Any value</td>
<td><code>NULL</code></td>
</tr>
<tr>
<td>00000000-0000-0000-0000-000000000000</td>
<td><code>&lt;</code></td>
<td>ffffffff-ffff-ffff-ffff-ffffffffffff</td>
<td><code>TRUE</code></td>
</tr>
<tr>
<td>00000000-0000-0000-0000-000000000000</td>
<td><code>=</code></td>
<td>00000000-0000-0000-0000-000000000000</td>
<td><code>TRUE</code></td>
</tr>
<tr>
<td>00000000-0000-0000-0000-000000000000</td>
<td><code>&gt;</code></td>
<td>ffffffff-ffff-ffff-ffff-ffffffffffff</td>
<td><code>FALSE</code></td>
</tr>
</tbody>
</table>

**Example**

```googlesql
  SELECT NEW_UUID() >= "00000000-0000-0000-0000-000000000000" AS Is_GE;

/*-------+
 | Is_GE |
 +-------+
 | true  |
 +-------*/
```

[array-nulls]: #array_nulls

[floating-point-types]: #floating_point_types

[lexical-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#literals

[join-types]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#join_types

[order-by-clause]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#order_by_clause

[st-equals]: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_equals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
