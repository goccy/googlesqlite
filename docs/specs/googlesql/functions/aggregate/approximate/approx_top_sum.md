---
name: APPROX_TOP_SUM
dialect: googlesql
category: functions/aggregate/approximate
status: implemented
source_url: docs/third_party/googlesql-docs/approximate_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/approximate_aggregate_functions.md#approx_top_sum
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/approximate/approx_top_sum.yaml
---

# APPROX_TOP_SUM

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

Verbatim copy from `docs/third_party/googlesql-docs/approximate_aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `APPROX_TOP_SUM`

```googlesql
APPROX_TOP_SUM(
  expression, weight, number
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
```

**Description**

Returns the approximate top elements of `expression`, ordered by the sum of the
`weight` values provided for each unique value of `expression`. The `number`
parameter specifies the number of elements returned.

If the `weight` input is negative or `NaN`, this function returns an error.

The elements are returned as an array of `STRUCT`s.
Each `STRUCT` contains two fields: `value` and `sum`.
The `value` field contains the value of the input expression. The `sum` field is
the same type as `weight`, and is the approximate sum of the input weight
associated with the `value` field.

Returns `NULL` if there are zero input rows.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

+ `expression`: Any data type that the `GROUP BY` clause supports.
+ `weight`: One of the following:

  + `INT64`
  + `UINT64`
  + `NUMERIC`
  + `BIGNUMERIC`
  + `DOUBLE`
+ `number`: `INT64` literal or query parameter.

**Returned Data Types**

`ARRAY<STRUCT>`

**Examples**

```googlesql
SELECT APPROX_TOP_SUM(x, weight, 2) AS approx_top_sum FROM
UNNEST([
  STRUCT("apple" AS x, 3 AS weight),
  ("pear", 2),
  ("apple", 0),
  ("banana", 5),
  ("pear", 4)
]);

/*--------------------------+
 | approx_top_sum           |
 +--------------------------+
 | [{pear, 6}, {banana, 5}] |
 +--------------------------*/
```

**NULL handling**

`APPROX_TOP_SUM` doesn't ignore `NULL` values for the `expression` and `weight`
parameters.

```googlesql
SELECT APPROX_TOP_SUM(x, weight, 2) AS approx_top_sum FROM
UNNEST([STRUCT("apple" AS x, NULL AS weight), ("pear", 0), ("pear", NULL)]);

/*----------------------------+
 | approx_top_sum             |
 +----------------------------+
 | [{pear, 0}, {apple, NULL}] |
 +----------------------------*/
```

```googlesql
SELECT APPROX_TOP_SUM(x, weight, 2) AS approx_top_sum FROM
UNNEST([STRUCT("apple" AS x, 0 AS weight), (NULL, 2)]);

/*-------------------------+
 | approx_top_sum          |
 +-------------------------+
 | [{NULL, 2}, {apple, 0}] |
 +-------------------------*/
```

```googlesql
SELECT APPROX_TOP_SUM(x, weight, 2) AS approx_top_sum FROM
UNNEST([STRUCT("apple" AS x, 0 AS weight), (NULL, NULL)]);

/*----------------------------+
 | approx_top_sum             |
 +----------------------------+
 | [{apple, 0}, {NULL, NULL}] |
 +----------------------------*/
```

[hll-functions]: https://github.com/google/googlesql/blob/master/docs/hll_functions.md

[aggregate-functions-reference]: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md

[agg-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/approximate_aggregate_functions.md`.
