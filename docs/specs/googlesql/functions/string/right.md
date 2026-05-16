---
name: RIGHT
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#right
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/right.yaml
---

# RIGHT

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

## `RIGHT`

```googlesql
RIGHT(value, length)
```

**Description**

Returns a `STRING` or `BYTES` value that consists of the specified
number of rightmost characters or bytes from `value`. The `length` is an
`INT64` that specifies the length of the returned
value. If `value` is `BYTES`, `length` is the number of rightmost bytes to
return. If `value` is `STRING`, `length` is the number of rightmost characters
to return.

If `length` is 0, an empty `STRING` or `BYTES` value will be
returned. If `length` is negative, an error will be returned. If `length`
exceeds the number of characters or bytes from `value`, the original `value`
will be returned.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT 'apple' AS example, RIGHT('apple', 3) AS right_example

/*---------+---------------+
 | example | right_example |
 +---------+---------------+
 | apple   | ple           |
 +---------+---------------*/
```

```googlesql
SELECT b'apple' AS example, RIGHT(b'apple', 3) AS right_example

/*----------------------+---------------+
 | example              | right_example |
 +----------------------+---------------+
 | apple                | ple           |
 +----------------------+---------------*
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
