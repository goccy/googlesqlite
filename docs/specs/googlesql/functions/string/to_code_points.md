---
name: TO_CODE_POINTS
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#to_code_points
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/to_code_points.yaml
---

# TO_CODE_POINTS

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

## `TO_CODE_POINTS`

```googlesql
TO_CODE_POINTS(value)
```

**Description**

Takes a `STRING` or `BYTES` value and returns an array of `INT64` values that
represent code points or extended ASCII character values.

+ If `value` is a `STRING`, each element in the returned array represents a
  [code point][string-link-to-code-points-wikipedia]. Each code point falls
  within the range of [0, 0xD7FF] and [0xE000, 0x10FFFF].
+ If `value` is `BYTES`, each element in the array is an extended ASCII
  character value in the range of [0, 255].

To convert from an array of code points to a `STRING` or `BYTES`, see
[CODE_POINTS_TO_STRING][string-link-to-codepoints-to-string] or
[CODE_POINTS_TO_BYTES][string-link-to-codepoints-to-bytes].

**Return type**

`ARRAY<INT64>`

**Examples**

The following examples get the code points for each element in an array of
words.

```googlesql
SELECT
  'foo' AS word,
  TO_CODE_POINTS('foo') AS code_points

/*---------+------------------------------------+
 | word    | code_points                        |
 +---------+------------------------------------+
 | foo     | [102, 111, 111]                    |
 +---------+------------------------------------*/
```

```googlesql
SELECT
  'bar' AS word,
  TO_CODE_POINTS('bar') AS code_points

/*---------+------------------------------------+
 | word    | code_points                        |
 +---------+------------------------------------+
 | bar     | [98, 97, 114]                      |
 +---------+------------------------------------*/
```

```googlesql
SELECT
  'baz' AS word,
  TO_CODE_POINTS('baz') AS code_points

/*---------+------------------------------------+
 | word    | code_points                        |
 +---------+------------------------------------+
 | baz     | [98, 97, 122]                      |
 +---------+------------------------------------*/
```

```googlesql
SELECT
  'giraffe' AS word,
  TO_CODE_POINTS('giraffe') AS code_points

/*---------+------------------------------------+
 | word    | code_points                        |
 +---------+------------------------------------+
 | giraffe | [103, 105, 114, 97, 102, 102, 101] |
 +---------+------------------------------------*/
```

```googlesql
SELECT
  'llama' AS word,
  TO_CODE_POINTS('llama') AS code_points

/*---------+------------------------------------+
 | word    | code_points                        |
 +---------+------------------------------------+
 | llama   | [108, 108, 97, 109, 97]            |
 +---------+------------------------------------*/
```

The following examples convert integer representations of `BYTES` to their
corresponding ASCII character values.

```googlesql
SELECT
  b'\x66\x6f\x6f' AS bytes_value,
  TO_CODE_POINTS(b'\x66\x6f\x6f') AS bytes_value_as_integer

/*------------------+------------------------+
 | bytes_value      | bytes_value_as_integer |
 +------------------+------------------------+
 | foo              | [102, 111, 111]        |
 +------------------+------------------------*/
```

```googlesql
SELECT
  b'\x00\x01\x10\xff' AS bytes_value,
  TO_CODE_POINTS(b'\x00\x01\x10\xff') AS bytes_value_as_integer

/*------------------+------------------------+
 | bytes_value      | bytes_value_as_integer |
 +------------------+------------------------+
 | \x00\x01\x10\xff | [0, 1, 16, 255]        |
 +------------------+------------------------*/
```

The following example demonstrates the difference between a `BYTES` result and a
`STRING` result. Notice that the character `Ā` is represented as a two-byte
Unicode sequence. As a result, the `BYTES` version of `TO_CODE_POINTS` returns
an array with two elements, while the `STRING` version returns an array with a
single element.

```googlesql
SELECT TO_CODE_POINTS(b'Ā') AS b_result, TO_CODE_POINTS('Ā') AS s_result;

/*------------+----------+
 | b_result   | s_result |
 +------------+----------+
 | [196, 128] | [256]    |
 +------------+----------*/
```

[string-link-to-code-points-wikipedia]: https://en.wikipedia.org/wiki/Code_point

[string-link-to-codepoints-to-string]: #code_points_to_string

[string-link-to-codepoints-to-bytes]: #code_points_to_bytes

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
