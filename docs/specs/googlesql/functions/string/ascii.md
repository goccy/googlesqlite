---
name: ASCII
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#ascii
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/ascii.yaml
---

# ASCII

## Summary

Returns the ASCII code for the first character or byte in `value`.

## Signatures

- `ASCII(value)`

## Behavior

- Return type: `INT64`.
- Returns the ASCII code of the first character or byte of the input `value`.
- Returns `0` when `value` is the empty string.
- Returns `0` when the first character or byte has an ASCII code of `0`.
- Returns `NULL` when `value` is `NULL`.

## Examples

```googlesql
SELECT ASCII('abcd') as A, ASCII('a') as B, ASCII('') as C, ASCII(NULL) as D;
-- expected: A=97, B=97, C=0, D=NULL
```

## Edge cases

- An empty `value` yields `0` rather than an error.
- A `NULL` input propagates as `NULL`.
- A leading character or byte whose ASCII code is `0` is indistinguishable in the output from the empty-string case (both return `0`).

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ASCII`

```googlesql
ASCII(value)
```

**Description**

Returns the ASCII code for the first character or byte in `value`. Returns
`0` if `value` is empty or the ASCII code is `0` for the first character
or byte.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT ASCII('abcd') as A, ASCII('a') as B, ASCII('') as C, ASCII(NULL) as D;

/*-------+-------+-------+-------+
 | A     | B     | C     | D     |
 +-------+-------+-------+-------+
 | 97    | 97    | 0     | NULL  |
 +-------+-------+-------+-------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
