---
name: ARRAY_LENGTH
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_length
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_length.yaml
---

# ARRAY_LENGTH

## Summary

Returns the size of an array as an `INT64`.

## Signatures

- `ARRAY_LENGTH(array_expression)`

## Behavior

- Return type is `INT64`.
- Returns the number of elements in `array_expression`.
- `NULL` elements are counted as part of the array length.
- Returns `0` when the input array is empty.
- Returns `NULL` when `array_expression` itself is `NULL`.

## Examples

```googlesql
SELECT
  ARRAY_LENGTH(["coffee", NULL, "milk"]) AS size_a,
  ARRAY_LENGTH(["cake", "pie"]) AS size_b;
-- expected: size_a = 3, size_b = 2
```

## Edge cases

- An empty array (`[]`) yields `0`, not `NULL`.
- A `NULL` array argument yields `NULL`.
- `NULL` elements inside the array do not reduce the reported length.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY_LENGTH`

```googlesql
ARRAY_LENGTH(array_expression)
```

**Description**

Returns the size of the array. Returns 0 for an empty array. Returns `NULL` if
the `array_expression` is `NULL`.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT
  ARRAY_LENGTH(["coffee", NULL, "milk" ]) AS size_a,
  ARRAY_LENGTH(["cake", "pie"]) AS size_b;

/*--------+--------+
 | size_a | size_b |
 +--------+--------+
 | 3      | 2      |
 +--------+--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
