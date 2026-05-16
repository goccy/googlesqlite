---
name: FROM_HEX
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#from_hex
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/from_hex.yaml
---

# FROM_HEX

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

## `FROM_HEX`

```googlesql
FROM_HEX(string)
```

**Description**

Converts a hexadecimal-encoded `STRING` into `BYTES` format. Returns an error
if the input `STRING` contains characters outside the range
`(0..9, A..F, a..f)`. The lettercase of the characters doesn't matter. If the
input `STRING` has an odd number of characters, the function acts as if the
input has an additional leading `0`. To convert `BYTES` to a hexadecimal-encoded
`STRING`, use [TO_HEX][string-link-to-to-hex].

**Return type**

`BYTES`

**Example**

```googlesql
WITH Input AS (
  SELECT '00010203aaeeefff' AS hex_str UNION ALL
  SELECT '0AF' UNION ALL
  SELECT '666f6f626172'
)
SELECT hex_str, FROM_HEX(hex_str) AS bytes_str
FROM Input;

/*------------------+----------------------------------+
 | hex_str          | bytes_str                        |
 +------------------+----------------------------------+
 | 0AF              | \x00\xaf                         |
 | 00010203aaeeefff | \x00\x01\x02\x03\xaa\xee\xef\xff |
 | 666f6f626172     | foobar                           |
 +------------------+----------------------------------*/
```

[string-link-to-to-hex]: #to_hex

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
