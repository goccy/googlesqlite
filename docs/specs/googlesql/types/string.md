---
name: STRING
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#string-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/string.yaml
---

# STRING

## Summary

`STRING` is a variable-length character (Unicode) data type. Input and output values are UTF-8 encoded, and string operations work on Unicode characters rather than raw bytes.

## Signatures

- `STRING`

## Behavior

- Stores variable-length Unicode character data.
- Input and output values must be UTF-8 encoded; alternate encodings such as CESU-8 and Modified UTF-8 are not treated as valid UTF-8.
- All functions and operators on strings operate on Unicode characters, not bytes; for example, `SUBSTR` and `LENGTH` count characters.
- Character ordering and comparison are based on Unicode code points: lower code points compare as less than higher code points.
- `STRING` and `BYTES` are distinct types with no implicit casting in either direction.
- Explicit casts between `STRING` and `BYTES` perform UTF-8 encoding and decoding.

## Examples

```sql
SELECT LENGTH('abc') AS len;
-- expected: 3
```

```sql
SELECT CAST(b'abc' AS STRING) AS s;
-- expected: 'abc'
```

## Edge cases

- Casting `BYTES` to `STRING` returns an error when the bytes are not valid UTF-8.
- Strings and bytes cannot be used interchangeably; mixing them without an explicit cast is a type error.
- Non-UTF-8 encodings (CESU-8, Modified UTF-8) are rejected as invalid input.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## String type 
<a id="string_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>STRING</code></td>
<td>Variable-length character (Unicode) data.</td>
</tr>
</tbody>
</table>

Input string values must be UTF-8 encoded and output string values will be UTF-8
encoded. Alternate encodings like CESU-8 and Modified UTF-8 aren't treated as
valid UTF-8.

All functions and operators that act on string values operate on Unicode
characters rather than bytes. For example, functions like `SUBSTR` and `LENGTH`
applied to string input count the number of characters, not bytes.

Each Unicode character has a numeric value called a code point assigned to it.
Lower code points are assigned to lower characters. When characters are
compared, the code points determine which characters are less than or greater
than other characters.

Most functions on strings are also defined on bytes. The bytes version
operates on raw bytes rather than Unicode characters. Strings and bytes are
separate types that can't be used interchangeably. There is no implicit casting
in either direction. Explicit casting between string and bytes does
UTF-8 encoding and decoding. Casting bytes to string returns an error if the
bytes aren't valid UTF-8.

To learn more about the literal representation of a string type,
see [String literals][string-literals].

[string-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#string_and_bytes_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
