---
name: RPAD
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#rpad
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/rpad.yaml
---

# RPAD

## Summary

Returns a `STRING` or `BYTES` value consisting of `original_value` appended on the right with `pattern` until the result reaches `return_length`. For `STRING` inputs `return_length` is measured in characters; for `BYTES` it is measured in bytes.

## Signatures

- `RPAD(original_value, return_length[, pattern])`

## Behavior

- Return type is `STRING` or `BYTES`, matching the type of `original_value`.
- Appends `pattern` to the right of `original_value` until the result reaches `return_length`.
- `return_length` is an `INT64`: number of characters when `original_value` is `STRING`, number of bytes when `original_value` is `BYTES`.
- The default value of `pattern` is a single blank space.
- Both `original_value` and `pattern` must be the same data type.
- If `return_length` is less than or equal to the length of `original_value`, the result is `original_value` truncated to `return_length`.
- If `pattern` is longer than the remaining padding space, only the leading portion of `pattern` needed to reach `return_length` is appended.

## Examples

```googlesql
SELECT FORMAT('%T', RPAD('c', 5)) AS results;
-- expected: "c    "
```

```googlesql
SELECT RPAD('b', 5, 'a') AS results;
-- expected: baaaa
```

```googlesql
SELECT RPAD('abc', 10, 'ghd') AS results;
-- expected: abcghdghdg
```

```googlesql
SELECT RPAD('abc', 2, 'd') AS results;
-- expected: ab
```

```googlesql
SELECT FORMAT('%T', RPAD(b'abc', 10, b'ghd')) AS results;
-- expected: b"abcghdghdg"
```

## Edge cases

- Returns `NULL` if any of `original_value`, `return_length`, or `pattern` is `NULL`.
- Raises an error if `return_length` is negative.
- Raises an error if `pattern` is empty.
- When `return_length` is less than or equal to the input length, the input is truncated rather than padded (e.g. `RPAD('hello world', 7)` returns `'hello w'`).
- `original_value` and `pattern` must share the same data type; mixing `STRING` and `BYTES` is not allowed.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `RPAD`

```googlesql
RPAD(original_value, return_length[, pattern])
```

**Description**

Returns a `STRING` or `BYTES` value that consists of `original_value` appended
with `pattern`. The `return_length` parameter is an
`INT64` that specifies the length of the
returned value. If `original_value` is `BYTES`,
`return_length` is the number of bytes. If `original_value` is `STRING`,
`return_length` is the number of characters.

The default value of `pattern` is a blank space.

Both `original_value` and `pattern` must be the same data type.

If `return_length` is less than or equal to the `original_value` length, this
function returns the `original_value` value, truncated to the value of
`return_length`. For example, `RPAD('hello world', 7);` returns `'hello w'`.

If `original_value`, `return_length`, or `pattern` is `NULL`, this function
returns `NULL`.

This function returns an error if:

+ `return_length` is negative
+ `pattern` is empty

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT FORMAT('%T', RPAD('c', 5)) AS results

/*---------+
 | results |
 +---------+
 | "c    " |
 +---------*/
```

```googlesql
SELECT RPAD('b', 5, 'a') AS results

/*---------+
 | results |
 +---------+
 | baaaa   |
 +---------*/
```

```googlesql
SELECT RPAD('abc', 10, 'ghd') AS results

/*------------+
 | results    |
 +------------+
 | abcghdghdg |
 +------------*/
```

```googlesql
SELECT RPAD('abc', 2, 'd') AS results

/*---------+
 | results |
 +---------+
 | ab      |
 +---------*/
```

```googlesql
SELECT FORMAT('%T', RPAD(b'abc', 10, b'ghd')) AS results

/*---------------+
 | results       |
 +---------------+
 | b"abcghdghdg" |
 +---------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
