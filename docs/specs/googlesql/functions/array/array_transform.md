---
name: ARRAY_TRANSFORM
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_transform
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_transform.yaml
---

# ARRAY_TRANSFORM

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

## `ARRAY_TRANSFORM`

```googlesql
ARRAY_TRANSFORM(array_expression, lambda_expression)

lambda_expression:
  {
    element_alias -> transform_expression
    | (element_alias, index_alias) -> transform_expression
  }
```

**Description**

Takes an array, transforms the elements, and returns the results in a new array.
The output array always has the same length as the input array.

+   `array_expression`: The array to transform.
+   `lambda_expression`: Each element in `array_expression` is evaluated against
    the [lambda expression][lambda-definition]. The evaluation results are
    returned in a new array.
+   `element_alias`: An alias that represents an array element.
+   `index_alias`: An alias that represents the zero-based offset of the array
    element.
+   `transform_expression`: The expression used to transform the array elements.

Returns `NULL` if the `array_expression` is `NULL`.

**Return type**

`ARRAY`

**Example**

```googlesql
SELECT
  ARRAY_TRANSFORM([1, 4, 3], e -> e + 1) AS a1,
  ARRAY_TRANSFORM([1, 4, 3], (e, i) -> e + i) AS a2;

/*---------+---------+
 | a1      | a2      |
 +---------+---------+
 | [2,5,4] | [1,5,5] |
 +---------+---------*/
```

[lambda-definition]: https://github.com/google/googlesql/blob/master/docs/functions-reference.md#lambdas

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
