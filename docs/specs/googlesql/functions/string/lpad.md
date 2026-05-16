---
name: LPAD
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#lpad
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/lpad.yaml
---

# LPAD

## Summary

Returns a `STRING` or `BYTES` value consisting of `original_value` left-padded with `pattern` to the length specified by `return_length`. The default `pattern` is a single blank space.

## Signatures

- `LPAD(original_value, return_length[, pattern])`

## Behavior

- Returns `STRING` or `BYTES`, matching the type of `original_value`.
- `return_length` is an `INT64`; for `BYTES` it counts bytes, for `STRING` it counts characters.
- Prepends copies of `pattern` to the left of `original_value` until the result reaches `return_length`.
- The default `pattern` is a single blank space when omitted.
- Both `original_value` and `pattern` must be the same data type.
- If `return_length` is less than or equal to the length of `original_value`, the function returns `original_value` truncated to `return_length`.

## Examples

```googlesql
SELECT FORMAT('%T', LPAD('c', 5)) AS results;
-- expected: "    c"
```

```googlesql
SELECT LPAD('b', 5, 'a') AS results;
-- expected: aaaab
```

```googlesql
SELECT LPAD('abc', 10, 'ghd') AS results;
-- expected: ghdghdgabc
```

```googlesql
SELECT LPAD('abc', 2, 'd') AS results;
-- expected: ab
```

```googlesql
SELECT FORMAT('%T', LPAD(b'abc', 10, b'ghd')) AS results;
-- expected: b"ghdghdgabc"
```

## Edge cases

- Returns `NULL` if `original_value`, `return_length`, or `pattern` is `NULL`.
- Raises an error if `return_length` is negative.
- Raises an error if `pattern` is empty.
- Raises an error if `original_value` and `pattern` are not the same data type.
- When `return_length` is less than or equal to the input length, the result is `original_value` truncated to `return_length` rather than padded.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `LPAD`

```googlesql
LPAD(original_value, return_length[, pattern])
```

**Description**

Returns a `STRING` or `BYTES` value that consists of `original_value` prepended
with `pattern`. The `return_length` is an `INT64` that
specifies the length of the returned value. If `original_value` is of type
`BYTES`, `return_length` is the number of bytes. If `original_value` is
of type `STRING`, `return_length` is the number of characters.

The default value of `pattern` is a blank space.

Both `original_value` and `pattern` must be the same data type.

If `return_length` is less than or equal to the `original_value` length, this
function returns the `original_value` value, truncated to the value of
`return_length`. For example, `LPAD('hello world', 7);` returns `'hello w'`.

If `original_value`, `return_length`, or `pattern` is `NULL`, this function
returns `NULL`.

This function returns an error if:

+ `return_length` is negative
+ `pattern` is empty

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT FORMAT('%T', LPAD('c', 5)) AS results

/*---------+
 | results |
 +---------+
 | "    c" |
 +---------*/
```

```googlesql
SELECT LPAD('b', 5, 'a') AS results

/*---------+
 | results |
 +---------+
 | aaaab   |
 +---------*/
```

```googlesql
SELECT LPAD('abc', 10, 'ghd') AS results

/*------------+
 | results    |
 +------------+
 | ghdghdgabc |
 +------------*/
```

```googlesql
SELECT LPAD('abc', 2, 'd') AS results

/*---------+
 | results |
 +---------+
 | ab      |
 +---------*/
```

```googlesql
SELECT FORMAT('%T', LPAD(b'abc', 10, b'ghd')) AS results

/*---------------+
 | results       |
 +---------------+
 | b"ghdghdgabc" |
 +---------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
