---
name: APPROX_QUANTILES
dialect: googlesql
category: functions/aggregate/approximate
status: implemented
source_url: docs/third_party/googlesql-docs/approximate_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/approximate_aggregate_functions.md#approx_quantiles
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/approximate/approx_quantiles.yaml
---

# APPROX_QUANTILES

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

## `APPROX_QUANTILES`

```googlesql
APPROX_QUANTILES(
  [ DISTINCT ]
  expression, number
  [ { IGNORE | RESPECT } NULLS ]
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
```

**Description**

Returns the approximate boundaries for a group of `expression` values, where
`number` represents the number of quantiles to create. This function returns an
array of `number` + 1 elements, sorted in ascending order, where the
first element is the approximate minimum and the last element is the approximate
maximum.

Returns `NULL` if there are zero input rows or `expression` evaluates to
`NULL` for all rows.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

+ `expression`: Any supported data type **except**:

  + `ARRAY`
  + `STRUCT`
  + `PROTO`
+ `number`: `INT64` literal or query parameter.

**Returned Data Types**

`ARRAY<T>` where `T` is the type specified by `expression`.

**Examples**

```googlesql
SELECT APPROX_QUANTILES(x, 2) AS approx_quantiles
FROM UNNEST([1, 1, 1, 4, 5, 6, 7, 8, 9, 10]) AS x;

/*------------------+
 | approx_quantiles |
 +------------------+
 | [1, 5, 10]       |
 +------------------*/
```

```googlesql
SELECT APPROX_QUANTILES(x, 100)[OFFSET(90)] AS percentile_90
FROM UNNEST([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) AS x;

/*---------------+
 | percentile_90 |
 +---------------+
 | 9             |
 +---------------*/
```

```googlesql
SELECT APPROX_QUANTILES(DISTINCT x, 2) AS approx_quantiles
FROM UNNEST([1, 1, 1, 4, 5, 6, 7, 8, 9, 10]) AS x;

/*------------------+
 | approx_quantiles |
 +------------------+
 | [1, 6, 10]       |
 +------------------*/
```

```googlesql
SELECT APPROX_QUANTILES(x, 2 RESPECT NULLS) AS approx_quantiles
FROM UNNEST([NULL, NULL, 1, 1, 1, 4, 5, 6, 7, 8, 9, 10]) AS x;

/*------------------+
 | approx_quantiles |
 +------------------+
 | [NULL, 4, 10]    |
 +------------------*/
```

```googlesql
SELECT APPROX_QUANTILES(DISTINCT x, 2 RESPECT NULLS) AS approx_quantiles
FROM UNNEST([NULL, NULL, 1, 1, 1, 4, 5, 6, 7, 8, 9, 10]) AS x;

/*------------------+
 | approx_quantiles |
 +------------------+
 | [NULL, 6, 10]    |
 +------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/approximate_aggregate_functions.md`.
