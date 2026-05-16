---
name: ARRAY_TO_STRING
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_to_string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_to_string.yaml
---

# ARRAY_TO_STRING

## Summary

Returns a concatenation of the elements in `array_expression`, separated by
`delimiter`, as a `STRING` or `BYTES` value.

## Signatures

- `ARRAY_TO_STRING(array_expression, delimiter[, null_text])`

## Behavior

- Returns `STRING` for a `STRING`-typed input array, and `BYTES` for a
  `BYTES`-typed input array.
- `array_expression` must be an array of `STRING` or `BYTES` elements.
- Joins the array elements in order using `delimiter` between adjacent
  elements.
- When `null_text` is supplied, each `NULL` element is replaced by
  `null_text` before joining.
- When `null_text` is omitted, each `NULL` element and its preceding
  delimiter are dropped from the result.

## Examples

```googlesql
SELECT ARRAY_TO_STRING(['coffee', 'tea', 'milk', NULL], '--', 'MISSING') AS text
-- expected: coffee--tea--milk--MISSING
```

```googlesql
SELECT ARRAY_TO_STRING(['cake', 'pie', NULL], '--') AS text
-- expected: cake--pie
```

```googlesql
SELECT ARRAY_TO_STRING([b'prefix', b'middle', b'suffix', b'\x00'], b'--') AS data
-- expected: prefix--middle--suffix--\x00
```

## Edge cases

- When `null_text` is omitted, `NULL` array elements are silently
  skipped along with their preceding delimiter, which can yield a
  shorter result than the input length suggests.
- When `null_text` is provided, `NULL` elements are materialised as
  `null_text` rather than skipped.
- `array_expression` must be an array of `STRING` or `BYTES`; other
  element types are not accepted.
- Mixing string and bytes inputs (e.g. a `BYTES` array with a `STRING`
  delimiter) is not supported by the signature.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY_TO_STRING`

```googlesql
ARRAY_TO_STRING(array_expression, delimiter[, null_text])
```

**Description**

Returns a concatenation of the elements in `array_expression` as a `STRING`
or `BYTES` value. The value for `array_expression` can
either be an array of `STRING` or `BYTES` data type.

If the `null_text` parameter is used, the function replaces any `NULL` values in
the array with the value of `null_text`.

If the `null_text` parameter isn't used, the function omits the `NULL` value
and its preceding delimiter.

**Return type**

* `STRING` for a function signature with `STRING` input.
* `BYTES` for a function signature with `BYTES` input.

**Examples**

```googlesql
SELECT ARRAY_TO_STRING(['coffee', 'tea', 'milk', NULL], '--', 'MISSING') AS text

/*--------------------------------+
 | text                           |
 +--------------------------------+
 | coffee--tea--milk--MISSING     |
 +--------------------------------*/
```

```googlesql

SELECT ARRAY_TO_STRING(['cake', 'pie', NULL], '--', 'MISSING') AS text

/*--------------------------------+
 | text                           |
 +--------------------------------+
 | cake--pie--MISSING             |
 +--------------------------------*/
```

```googlesql

SELECT ARRAY_TO_STRING([b'prefix', b'middle', b'suffix', b'\x00'], b'--') AS data

/*--------------------------------+
 | data                           |
 +--------------------------------+
 | prefix--middle--suffix--\x00   |
 +--------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
