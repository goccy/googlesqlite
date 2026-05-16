---
name: SUBSTR
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#substr
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/substr.yaml
---

# SUBSTR

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

## `SUBSTR`

```googlesql
SUBSTR(value, position[, length])
```

**Description**

Gets a portion (substring) of the supplied `STRING` or `BYTES` value.

The `position` argument is an integer specifying the starting position of the
substring.

+ If `position` is `1`, the substring starts from the first character or byte.
+ If `position` is `0` or less than `-LENGTH(value)`, `position` is set to `1`,
  and the substring starts from the first character or byte.
+ If `position` is greater than the length of `value`, the function produces
  an empty substring.
+ If `position` is negative, the function counts from the end of `value`,
  with `-1` indicating the last character or byte.

The `length` argument specifies the maximum number of characters or bytes to
return.

+ If `length` isn't specified, the function produces a substring that starts
  at the specified position and ends at the last character or byte of `value`.
+ If `length` is `0`, the function produces an empty substring.
+ If `length` is negative, the function produces an error.
+ The returned substring may be shorter than `length`, for example, when
  `length` exceeds the length of `value`, or when the starting position of the
  substring plus `length` is greater than the length of `value`.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT SUBSTR('apple', 2) AS example

/*---------+
 | example |
 +---------+
 | pple    |
 +---------*/
```

```googlesql
SELECT SUBSTR('apple', 2, 2) AS example

/*---------+
 | example |
 +---------+
 | pp      |
 +---------*/
```

```googlesql
SELECT SUBSTR('apple', -2) AS example

/*---------+
 | example |
 +---------+
 | le      |
 +---------*/
```

```googlesql
SELECT SUBSTR('apple', 1, 123) AS example

/*---------+
 | example |
 +---------+
 | apple   |
 +---------*/
```

```googlesql
SELECT SUBSTR('apple', 123) AS example

/*---------+
 | example |
 +---------+
 |         |
 +---------*/
```

```googlesql
SELECT SUBSTR('apple', 123, 5) AS example

/*---------+
 | example |
 +---------+
 |         |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
