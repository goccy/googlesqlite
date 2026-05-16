---
name: ARRAY_FILTER
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_filter
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_filter.yaml
---

# ARRAY_FILTER

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

## `ARRAY_FILTER`

```googlesql
ARRAY_FILTER(array_expression, lambda_expression)

lambda_expression:
  {
    element_alias -> boolean_expression
    | (element_alias, index_alias) -> boolean_expression
  }
```

**Description**

Takes an array, filters out unwanted elements, and returns the results in a new
array.

+   `array_expression`: The array to filter.
+   `lambda_expression`: Each element in `array_expression` is evaluated against
    the [lambda expression][lambda-definition]. If the expression evaluates to
    `FALSE` or `NULL`, the element is removed from the resulting array.
+   `element_alias`: An alias that represents an array element.
+   `index_alias`: An alias that represents the zero-based offset of the array
    element.
+   `boolean_expression`: The predicate used to filter the array elements.

Returns `NULL` if the `array_expression` is `NULL`.

**Return type**

ARRAY

**Example**

```googlesql
SELECT
  ARRAY_FILTER([1 ,2, 3], e -> e > 1) AS a1,
  ARRAY_FILTER([0, 2, 3], (e, i) -> e > i) AS a2;

/*-------+-------+
 | a1    | a2    |
 +-------+-------+
 | [2,3] | [2,3] |
 +-------+-------*/
```

[lambda-definition]: https://github.com/google/googlesql/blob/master/docs/functions-reference.md#lambdas

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
