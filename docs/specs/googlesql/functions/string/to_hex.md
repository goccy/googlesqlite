---
name: TO_HEX
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#to_hex
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/to_hex.yaml
---

# TO_HEX

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

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `TO_HEX`

```googlesql
TO_HEX(bytes)
```

**Description**

Converts a sequence of `BYTES` into a hexadecimal `STRING`. Converts each byte
in the `STRING` as two hexadecimal characters in the range
`(0..9, a..f)`. To convert a hexadecimal-encoded
`STRING` to `BYTES`, use [FROM_HEX][string-link-to-from-hex].

**Return type**

`STRING`

**Example**

```googlesql
SELECT
  b'\x00\x01\x02\x03\xAA\xEE\xEF\xFF' AS byte_string,
  TO_HEX(b'\x00\x01\x02\x03\xAA\xEE\xEF\xFF') AS hex_string

/*----------------------------------+------------------+
 | byte_string                      | hex_string       |
 +----------------------------------+------------------+
 | \x00\x01\x02\x03\xaa\xee\xef\xff | 00010203aaeeefff |
 +----------------------------------+------------------*/
```

[string-link-to-from-hex]: #from_hex

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
