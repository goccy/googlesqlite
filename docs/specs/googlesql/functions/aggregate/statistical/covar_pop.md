---
name: COVAR_POP
dialect: googlesql
category: functions/aggregate/statistical
status: implemented
source_url: docs/third_party/googlesql-docs/statistical_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/statistical_aggregate_functions.md#covar_pop
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/statistical/covar_pop.yaml
---

# COVAR_POP

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

Verbatim copy from `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `COVAR_POP`

```googlesql
COVAR_POP(
  X1, X2
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
[ OVER over_clause ]

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]
  [ ORDER BY expression [ { ASC | DESC }  ] [, ...] ]
  [ window_frame_clause ]

```

**Description**

Returns the population [covariance][stat-agg-link-to-covariance] of
a set of number pairs. The first number is the dependent variable; the second
number is the independent variable. The return result is between `-Inf` and
`+Inf`.

All numeric types are supported. If the
input is `NUMERIC` or `BIGNUMERIC` then the internal aggregation is
stable with the final output converted to a `DOUBLE`.
Otherwise the input is converted to a `DOUBLE`
before aggregation, resulting in a potentially unstable result.

This function ignores any input pairs that contain one or more `NULL` values. If
there is no input pair without `NULL` values, this function returns `NULL`.
If there is exactly one input pair without `NULL` values, this function returns
`0`.

`NaN` is produced if:

+ Any input value is `NaN`
+ Any input value is positive infinity or negative infinity.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

<!-- mdlint on -->

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Return Data Type**

`DOUBLE`

**Examples**

```googlesql
SELECT COVAR_POP(y, x) AS results
FROM
  UNNEST(
    [
      STRUCT(1.0 AS y, 1.0 AS x),
      (2.0, 6.0),
      (9.0, 3.0),
      (2.0, 6.0),
      (9.0, 3.0)])

/*---------------------+
 | results             |
 +---------------------+
 | -1.6800000000000002 |
 +---------------------*/
```

```googlesql
SELECT COVAR_POP(y, x) AS results
FROM UNNEST([STRUCT(1.0 AS y, NULL AS x),(9.0, 3.0)])

/*---------+
 | results |
 +---------+
 | 0       |
 +---------*/
```

```googlesql
SELECT COVAR_POP(y, x) AS results
FROM UNNEST([STRUCT(1.0 AS y, NULL AS x),(9.0, NULL)])

/*---------+
 | results |
 +---------+
 | NULL    |
 +---------*/
```

```googlesql
SELECT COVAR_POP(y, x) AS results
FROM
  UNNEST(
    [
      STRUCT(1.0 AS y, 1.0 AS x),
      (2.0, 6.0),
      (9.0, 3.0),
      (2.0, 6.0),
      (NULL, 3.0)])

/*---------+
 | results |
 +---------+
 | -1      |
 +---------*/
```

```googlesql
SELECT COVAR_POP(y, x) AS results
FROM
  UNNEST(
    [
      STRUCT(1.0 AS y, 1.0 AS x),
      (2.0, 6.0),
      (9.0, 3.0),
      (2.0, 6.0),
      (CAST('Infinity' as DOUBLE), 3.0)])

/*---------+
 | results |
 +---------+
 | NaN     |
 +---------*/
```

[stat-agg-link-to-covariance]: https://en.wikipedia.org/wiki/Covariance

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`.
