---
name: BYTES
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#bytes-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/bytes.yaml
---

# BYTES

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

## Bytes type 
<a id="bytes_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>BYTES</code></td>
<td>Variable-length binary data.</td>
</tr>
</tbody>
</table>

String and bytes are separate types that can't be used interchangeably.
Most functions on strings are also defined on bytes. The bytes version
operates on raw bytes rather than Unicode characters. Casts between string and
bytes enforce that the bytes are encoded using UTF-8.

You can convert a base64-encoded `STRING` expression into the `BYTES` format
using the
[`FROM_BASE64` function][from-base].
You can also convert a sequence of `BYTES` into a base64-encoded `STRING`
expression using the
[`TO_BASE64` function][to-base].

To learn more about the literal representation of a bytes type,
see [Bytes literals][bytes-literals].

[bytes-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#string_and_bytes_literals

[from-base]: https://github.com/google/googlesql/blob/master/docs/string_functions.md#from_base64

[to-base]: https://github.com/google/googlesql/blob/master/docs/string_functions.md#to_base64

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
