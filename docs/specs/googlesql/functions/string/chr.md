---
name: CHR
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#chr
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/chr.yaml
---

# CHR

## Summary

Takes a Unicode code point and returns the character that matches it.

## Signatures

- ```googlesql
  CHR(value)
  ```

## Behavior

- Return type is `STRING`.
- Accepts a Unicode code point and returns the matching character.
- Valid code points fall within the ranges `[0, 0xD7FF]` and `[0xE000, 0x10FFFF]`.
- Returns an empty string when the code point is `0`.
- To work with an array of Unicode code points, use `CODE_POINTS_TO_STRING`.

## Examples

```googlesql
SELECT CHR(65) AS A, CHR(255) AS B, CHR(513) AS C, CHR(1024) AS D;
-- expected A='A', B='ÿ', C='ȁ', D='Ѐ'
```

```googlesql
SELECT CHR(97) AS A, CHR(0xF9B5) AS B, CHR(0) AS C, CHR(NULL) AS D;
-- expected A='a', B='例', C='', D=NULL
```

## Edge cases

- Returns an empty string when the code point is `0`.
- Returns `NULL` when the input is `NULL`.
- Raises an error when an invalid Unicode code point is supplied (outside `[0, 0xD7FF]` and `[0xE000, 0x10FFFF]`).

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `CHR`

```googlesql
CHR(value)
```

**Description**

Takes a Unicode [code point][string-link-to-code-points-wikipedia] and returns
the character that matches the code point. Each valid code point should fall
within the range of [0, 0xD7FF] and [0xE000, 0x10FFFF]. Returns an empty string
if the code point is `0`. If an invalid Unicode code point is specified, an
error is returned.

To work with an array of Unicode code points, see
[`CODE_POINTS_TO_STRING`][string-link-to-codepoints-to-string]

**Return type**

`STRING`

**Examples**

```googlesql
SELECT CHR(65) AS A, CHR(255) AS B, CHR(513) AS C, CHR(1024)  AS D;

/*-------+-------+-------+-------+
 | A     | B     | C     | D     |
 +-------+-------+-------+-------+
 | A     | ÿ     | ȁ     | Ѐ     |
 +-------+-------+-------+-------*/
```

```googlesql
SELECT CHR(97) AS A, CHR(0xF9B5) AS B, CHR(0) AS C, CHR(NULL) AS D;

/*-------+-------+-------+-------+
 | A     | B     | C     | D     |
 +-------+-------+-------+-------+
 | a     | 例    |       | NULL  |
 +-------+-------+-------+-------*/
```

[string-link-to-code-points-wikipedia]: https://en.wikipedia.org/wiki/Code_point

[string-link-to-codepoints-to-string]: #code_points_to_string

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
