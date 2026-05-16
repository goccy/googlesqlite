---
name: LEFT
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#left
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/left.yaml
---

# LEFT

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

## `LEFT`

```googlesql
LEFT(value, length)
```

**Description**

Returns a `STRING` or `BYTES` value that consists of the specified
number of leftmost characters or bytes from `value`. The `length` is an
`INT64` that specifies the length of the returned
value. If `value` is of type `BYTES`, `length` is the number of leftmost bytes
to return. If `value` is `STRING`, `length` is the number of leftmost characters
to return.

If `length` is 0, an empty `STRING` or `BYTES` value will be
returned. If `length` is negative, an error will be returned. If `length`
exceeds the number of characters or bytes from `value`, the original `value`
will be returned.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT LEFT('banana', 3) AS results

/*---------+
 | results |
  +--------+
 | ban     |
 +---------*/
```

```googlesql
SELECT LEFT(b'\xab\xcd\xef\xaa\xbb', 3) AS results

/*--------------+
 | results      |
 +--------------+
 | \xab\xcd\xef |
 +--------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
