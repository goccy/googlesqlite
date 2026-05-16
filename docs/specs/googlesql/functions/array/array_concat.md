---
name: ARRAY_CONCAT
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_concat
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_concat.yaml
---

# ARRAY_CONCAT

## Summary

Concatenates one or more arrays that share the same element type into a single array.

## Signatures

- `ARRAY_CONCAT(array_expression[, ...])`

## Behavior

- Returns an `ARRAY` whose element type matches the inputs.
- Accepts one or more `array_expression` arguments; all inputs must share the same element type.
- Produces a single array containing the elements of every input array, in argument order.
- Returns `NULL` if any input argument is `NULL`.
- The `||` concatenation operator can be used as an equivalent alternative.

## Examples

```googlesql
SELECT ARRAY_CONCAT([1, 2], [3, 4], [5, 6]) AS count_to_six;
-- expected count_to_six: [1, 2, 3, 4, 5, 6]
```

## Edge cases

- Returns `NULL` whenever any input argument is `NULL`.
- All input arrays must share the same element type; mismatched element types are not supported.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY_CONCAT`

```googlesql
ARRAY_CONCAT(array_expression[, ...])
```

**Description**

Concatenates one or more arrays with the same element type into a single array.

The function returns `NULL` if any input argument is `NULL`.

Note: You can also use the [|| concatenation operator][array-link-to-operators]
to concatenate arrays.

**Return type**

`ARRAY`

**Examples**

```googlesql
SELECT ARRAY_CONCAT([1, 2], [3, 4], [5, 6]) as count_to_six;

/*--------------------------------------------------+
 | count_to_six                                     |
 +--------------------------------------------------+
 | [1, 2, 3, 4, 5, 6]                               |
 +--------------------------------------------------*/
```

[array-link-to-operators]: https://github.com/google/googlesql/blob/master/docs/operators.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
