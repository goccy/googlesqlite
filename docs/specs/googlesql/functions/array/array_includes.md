---
name: ARRAY_INCLUDES
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_includes
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_includes.yaml
---

# ARRAY_INCLUDES

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

## `ARRAY_INCLUDES`

+   [Signature 1](#array_includes_signature1):
    `ARRAY_INCLUDES(array_to_search, search_value)`
+   [Signature 2](#array_includes_signature2):
    `ARRAY_INCLUDES(array_to_search, lambda_expression)`

#### Signature 1 
<a id="array_includes_signature1"></a>

```googlesql
ARRAY_INCLUDES(array_to_search, search_value)
```

**Description**

Takes an array and returns `TRUE` if there is an element in the array that is
equal to the search_value.

+   `array_to_search`: The array to search.
+   `search_value`: The element to search for in the array.

Returns `NULL` if `array_to_search` or `search_value` is `NULL`.

**Return type**

`BOOL`

**Example**

In the following example, the query first checks to see if `0` exists in an
array. Then the query checks to see if `1` exists in an array.

```googlesql
SELECT
  ARRAY_INCLUDES([1, 2, 3], 0) AS a1,
  ARRAY_INCLUDES([1, 2, 3], 1) AS a2;

/*-------+------+
 | a1    | a2   |
 +-------+------+
 | false | true |
 +-------+------*/
```

#### Signature 2 
<a id="array_includes_signature2"></a>

```googlesql
ARRAY_INCLUDES(array_to_search, lambda_expression)

lambda_expression: element_alias -> boolean_expression
```

**Description**

Takes an array and returns `TRUE` if the lambda expression evaluates to `TRUE`
for any element in the array.

+   `array_to_search`: The array to search.
+   `lambda_expression`: Each element in `array_to_search` is evaluated against
    the [lambda expression][lambda-definition].
+   `element_alias`: An alias that represents an array element.
+   `boolean_expression`: The predicate used to evaluate the array elements.

Returns `NULL` if `array_to_search` is `NULL`.

**Return type**

`BOOL`

**Example**

In the following example, the query first checks to see if any elements that are
greater than 3 exist in an array (`e > 3`). Then the query checks to see if any
elements that are greater than 0 exist in an array (`e > 0`).

```googlesql
SELECT
  ARRAY_INCLUDES([1, 2, 3], e -> e > 3) AS a1,
  ARRAY_INCLUDES([1, 2, 3], e -> e > 0) AS a2;

/*-------+------+
 | a1    | a2   |
 +-------+------+
 | false | true |
 +-------+------*/
```

[lambda-definition]: https://github.com/google/googlesql/blob/master/docs/functions-reference.md#lambdas

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
