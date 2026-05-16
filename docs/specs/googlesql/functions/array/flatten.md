---
name: FLATTEN
dialect: googlesql
category: functions/array
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#flatten
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/flatten.yaml
---

# FLATTEN

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

## `FLATTEN`

```googlesql
FLATTEN(array_elements_field_access_expression)
```

**Description**

Takes an array of nested data and flattens a specific part of it into a single,
flat array with the
[array elements field access operator][array-el-field-operator].
Returns `NULL` if the input value is `NULL`.
If `NULL` array elements are
encountered, they are added to the resulting array.

There are several ways to flatten nested data into arrays. To learn more, see
[Flattening nested data into an array][flatten-tree-to-array].

**Return type**

`ARRAY`

**Examples**

In the following example, all of the arrays for `v.sales.quantity` are
concatenated in a flattened array.

```googlesql
WITH t AS (
  SELECT
  [
    STRUCT([STRUCT([1,2,3] AS quantity), STRUCT([4,5,6] AS quantity)] AS sales),
    STRUCT([STRUCT([7,8] AS quantity), STRUCT([] AS quantity)] AS sales)
  ] AS v
)
SELECT FLATTEN(v.sales.quantity) AS all_values
FROM t;

/*--------------------------+
 | all_values               |
 +--------------------------+
 | [1, 2, 3, 4, 5, 6, 7, 8] |
 +--------------------------*/
```

In the following example, `OFFSET` gets the second value in each array and
concatenates them.

```googlesql
WITH t AS (
  SELECT
  [
    STRUCT([STRUCT([1,2,3] AS quantity), STRUCT([4,5,6] AS quantity)] AS sales),
    STRUCT([STRUCT([7,8,9] AS quantity), STRUCT([10,11,12] AS quantity)] AS sales)
  ] AS v
)
SELECT FLATTEN(v.sales.quantity[OFFSET(1)]) AS second_values
FROM t;

/*---------------+
 | second_values |
 +---------------+
 | [2, 5, 8, 11] |
 +---------------*/
```

In the following example, all values for `v.price` are returned in a
flattened array.

```googlesql
WITH t AS (
  SELECT
  [
    STRUCT(1 AS price, 2 AS quantity),
    STRUCT(10 AS price, 20 AS quantity)
  ] AS v
)
SELECT FLATTEN(v.price) AS all_prices
FROM t;

/*------------+
 | all_prices |
 +------------+
 | [1, 10]    |
 +------------*/
```

For more examples, including how to use protocol buffers with `FLATTEN`, see the
[array elements field access operator][array-el-field-operator].

[flatten-tree-to-array]: https://github.com/google/googlesql/blob/master/docs/arrays.md#flattening_nested_data_into_arrays

[array-el-field-operator]: https://github.com/google/googlesql/blob/master/docs/operators.md#array_el_field_operator

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
