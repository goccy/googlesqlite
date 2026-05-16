---
name: CORR
dialect: googlesql
category: functions/aggregate/statistical
status: implemented
source_url: docs/third_party/googlesql-docs/statistical_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/statistical_aggregate_functions.md#corr
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/statistical/corr.yaml
---

# CORR

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

## `CORR`

```googlesql
CORR(
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

Returns the [Pearson coefficient][stat-agg-link-to-pearson-coefficient]
of correlation of a set of number pairs. For each number pair, the first number
is the dependent variable and the second number is the independent variable.
The return result is between `-1` and `1`. A result of `0` indicates no
correlation.

All numeric types are supported. If the
input is `NUMERIC` or `BIGNUMERIC` then the internal aggregation is
stable with the final output converted to a `DOUBLE`.
Otherwise the input is converted to a `DOUBLE`
before aggregation, resulting in a potentially unstable result.

This function ignores any input pairs that contain one or more `NULL` values. If
there are fewer than two input pairs without `NULL` values, this function
returns `NULL`.

`NaN` is produced if:

+ Any input value is `NaN`
+ Any input value is positive infinity or negative infinity.
+ The variance of `X1` or `X2` is `0`.
+ The covariance of `X1` and `X2` is `0`.

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
SELECT CORR(y, x) AS results
FROM
  UNNEST(
    [
      STRUCT(1.0 AS y, 5.0 AS x),
      (3.0, 9.0),
      (4.0, 7.0)]);

/*--------------------+
 | results            |
 +--------------------+
 | 0.6546536707079772 |
 +--------------------*/
```

```googlesql
SELECT CORR(y, x) AS results
FROM
  UNNEST(
    [
      STRUCT(1.0 AS y, 5.0 AS x),
      (3.0, 9.0),
      (4.0, NULL)]);

/*---------+
 | results |
 +---------+
 | 1       |
 +---------*/
```

```googlesql
SELECT CORR(y, x) AS results
FROM UNNEST([STRUCT(1.0 AS y, NULL AS x),(9.0, 3.0)])

/*---------+
 | results |
 +---------+
 | NULL    |
 +---------*/
```

```googlesql
SELECT CORR(y, x) AS results
FROM UNNEST([STRUCT(1.0 AS y, NULL AS x),(9.0, NULL)])

/*---------+
 | results |
 +---------+
 | NULL    |
 +---------*/
```

```googlesql
SELECT CORR(y, x) AS results
FROM
  UNNEST(
    [
      STRUCT(1.0 AS y, 5.0 AS x),
      (3.0, 9.0),
      (4.0, 7.0),
      (5.0, 1.0),
      (7.0, CAST('Infinity' as DOUBLE))])

/*---------+
 | results |
 +---------+
 | NaN     |
 +---------*/
```

```googlesql
SELECT CORR(x, y) AS results
FROM
  (
    SELECT 0 AS x, 0 AS y
    UNION ALL
    SELECT 0 AS x, 0 AS y
  )

/*---------+
 | results |
 +---------+
 | NaN     |
 +---------*/
```

[stat-agg-link-to-pearson-coefficient]: https://en.wikipedia.org/wiki/Pearson_product-moment_correlation_coefficient

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`.
