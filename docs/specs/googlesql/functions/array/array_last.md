---
name: ARRAY_LAST
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_last
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_last.yaml
---

# ARRAY_LAST

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

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY_LAST`

```googlesql
ARRAY_LAST(array_expression)
```

**Description**

Takes an array and returns the last element in the array.

Produces an error if the array is empty.

Returns `NULL` if `array_expression` is `NULL`.

Note: To get the first element in an array, see [`ARRAY_FIRST`][array-first].

**Return type**

Matches the data type of elements in `array_expression`.

**Example**

```googlesql
SELECT ARRAY_LAST(['a','b','c','d']) as last_element

/*---------------+
 | last_element  |
 +---------------+
 | d             |
 +---------------*/
```

[array-first]: #array_first

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
