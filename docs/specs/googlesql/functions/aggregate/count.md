---
name: COUNT
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#count
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/count.yaml
---

# COUNT

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

Verbatim copy from `docs/third_party/googlesql-docs/aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `COUNT`

```googlesql
COUNT(*)
[ OVER over_clause ]
```

```googlesql
COUNT(
  [ DISTINCT ]
  expression
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

Gets the number of rows in the input or the number of rows with an
expression evaluated to any value other than `NULL`.

Note: If you're querying a large dataset, you can compute results faster and
save resources by using [HLL++ functions][hll-functions] for approximate
distinct counts. For more information, see
[Sketches][sketches].

**Definitions**

+ `*`: Use this value to get the number of all rows in the input.
+ `expression`: A value of any data type that represents the expression to
  evaluate. If `DISTINCT` is present,
  `expression` can only be a data type that is
  [groupable][groupable-data-types].
+   `DISTINCT`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `WHERE`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `HAVING { MAX | MIN }`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `OVER`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `over_clause`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `window_specification`: To learn more, see
    [Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

[agg-threshold-clause]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#agg_threshold_clause

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Details**

To count the number of distinct values of an expression for which a
certain condition is satisfied, you can use the following recipe:

```googlesql
COUNT(DISTINCT IF(condition, expression, NULL))
```

`IF` returns the value of `expression` if `condition` is `TRUE`, or
`NULL` otherwise. The surrounding `COUNT(DISTINCT ...)` ignores the `NULL`
values, so it counts only the distinct values of `expression` for which
`condition` is `TRUE`.

To count the number of non-distinct values of an expression for which a
certain condition is satisfied, consider using the
[`COUNTIF`][countif] function.

This function with `DISTINCT` supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

`COUNT` can be used with differential privacy. For more information, see
[Differentially private aggregate functions][dp-functions].

**Return type**

`INT64`

**Examples**

You can use the `COUNT` function to return the number of rows in a table or the
number of distinct values of an expression. For example:

```googlesql
SELECT
  COUNT(*) AS count_star,
  COUNT(DISTINCT x) AS count_dist_x
FROM UNNEST([1, 4, 4, 5]) AS x;

/*------------+--------------+
 | count_star | count_dist_x |
 +------------+--------------+
 | 4          | 3            |
 +------------+--------------*/
```

```googlesql
SELECT
  x,
  COUNT(*) OVER (PARTITION BY MOD(x, 3)) AS count_star,
  COUNT(DISTINCT x) OVER (PARTITION BY MOD(x, 3)) AS count_dist_x
FROM UNNEST([1, 4, 4, 5]) AS x;

/*------+------------+--------------+
 | x    | count_star | count_dist_x |
 +------+------------+--------------+
 | 1    | 3          | 2            |
 | 4    | 3          | 2            |
 | 4    | 3          | 2            |
 | 5    | 1          | 1            |
 +------+------------+--------------*/
```

```googlesql
SELECT
  x,
  COUNT(*) OVER (PARTITION BY MOD(x, 3)) AS count_star,
  COUNT(x) OVER (PARTITION BY MOD(x, 3)) AS count_x
FROM UNNEST([1, 4, NULL, 4, 5]) AS x;

/*------+------------+---------+
 | x    | count_star | count_x |
 +------+------------+---------+
 | NULL | 1          | 0       |
 | 1    | 3          | 3       |
 | 4    | 3          | 3       |
 | 4    | 3          | 3       |
 | 5    | 1          | 1       |
 +------+------------+---------*/
```

The following query counts the number of distinct positive values of `x`:

```googlesql
SELECT COUNT(DISTINCT IF(x > 0, x, NULL)) AS distinct_positive
FROM UNNEST([1, -2, 4, 1, -5, 4, 1, 3, -6, 1]) AS x;

/*-------------------+
 | distinct_positive |
 +-------------------+
 | 3                 |
 +-------------------*/
```

The following query counts the number of distinct dates on which a certain kind
of event occurred:

```googlesql
WITH Events AS (
  SELECT DATE '2021-01-01' AS event_date, 'SUCCESS' AS event_type
  UNION ALL
  SELECT DATE '2021-01-02' AS event_date, 'SUCCESS' AS event_type
  UNION ALL
  SELECT DATE '2021-01-02' AS event_date, 'FAILURE' AS event_type
  UNION ALL
  SELECT DATE '2021-01-03' AS event_date, 'SUCCESS' AS event_type
  UNION ALL
  SELECT DATE '2021-01-04' AS event_date, 'FAILURE' AS event_type
  UNION ALL
  SELECT DATE '2021-01-04' AS event_date, 'FAILURE' AS event_type
)
SELECT
  COUNT(DISTINCT IF(event_type = 'FAILURE', event_date, NULL))
    AS distinct_dates_with_failures
FROM Events;

/*------------------------------+
 | distinct_dates_with_failures |
 +------------------------------+
 | 2                            |
 +------------------------------*/
```

The following query counts the number of distinct `id`s that exist in both
the `customers` and `vendor` tables:

```googlesql
WITH
  customers AS (
    SELECT 1934 AS id, 'a' AS team UNION ALL
    SELECT 2991, 'b' UNION ALL
    SELECT 3988, 'c'),
  vendors AS (
    SELECT 1934 AS id, 'd' AS team UNION ALL
    SELECT 2991, 'e' UNION ALL
    SELECT 4366, 'f')
SELECT
  COUNT(DISTINCT IF(id IN (SELECT id FROM customers), id, NULL)) AS result
FROM vendors;

/*--------+
 | result |
 +--------+
 | 2      |
 +--------*/
```

[sketches]: https://github.com/google/googlesql/blob/master/docs/sketches.md

[hll-functions]: https://github.com/google/googlesql/blob/master/docs/hll_functions.md

[countif]: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#countif

[groupable-data-types]: https://github.com/google/googlesql/blob/master/docs/data-types.md#groupable_data_types

[dp-functions]: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
