---
name: ARRAY_AGG
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#array_agg
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/array_agg.yaml
---

# ARRAY_AGG

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

## `ARRAY_AGG`

```googlesql
ARRAY_AGG(
  [ DISTINCT ]
  expression
  [ { IGNORE | RESPECT } NULLS ]
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
  [ ORDER BY key [ { ASC | DESC } ] [, ... ] ]
  [ LIMIT n ]
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

Returns an ARRAY of `expression` values.

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

**Supported Argument Types**

All data types except ARRAY.

**Returned Data Types**

ARRAY

If there are zero input rows, this function returns `NULL`.

**Examples**

```googlesql
SELECT ARRAY_AGG(x) AS array_agg FROM UNNEST([2, 1,-2, 3, -2, 1, 2]) AS x;

/*-------------------------+
 | array_agg               |
 +-------------------------+
 | [2, 1, -2, 3, -2, 1, 2] |
 +-------------------------*/
```

```googlesql
SELECT ARRAY_AGG(DISTINCT x) AS array_agg
FROM UNNEST([2, 1, -2, 3, -2, 1, 2]) AS x;

/*---------------+
 | array_agg     |
 +---------------+
 | [2, 1, -2, 3] |
 +---------------*/
```

```googlesql
SELECT ARRAY_AGG(x IGNORE NULLS) AS array_agg
FROM UNNEST([NULL, 1, -2, 3, -2, 1, NULL]) AS x;

/*-------------------+
 | array_agg         |
 +-------------------+
 | [1, -2, 3, -2, 1] |
 +-------------------*/
```

```googlesql
SELECT ARRAY_AGG(x ORDER BY ABS(x)) AS array_agg
FROM UNNEST([2, 1, -2, 3, -2, 1, 2]) AS x;

/*-------------------------+
 | array_agg               |
 +-------------------------+
 | [1, 1, 2, -2, -2, 2, 3] |
 +-------------------------*/
```

```googlesql
SELECT ARRAY_AGG(x LIMIT 5) AS array_agg
FROM UNNEST([2, 1, -2, 3, -2, 1, 2]) AS x;

/*-------------------+
 | array_agg         |
 +-------------------+
 | [2, 1, -2, 3, -2] |
 +-------------------*/
```

```googlesql
WITH vals AS
  (
    SELECT 1 x UNION ALL
    SELECT -2 x UNION ALL
    SELECT 3 x UNION ALL
    SELECT -2 x UNION ALL
    SELECT 1 x
  )
SELECT ARRAY_AGG(DISTINCT x ORDER BY x) as array_agg
FROM vals;

/*------------+
 | array_agg  |
 +------------+
 | [-2, 1, 3] |
 +------------*/
```

```googlesql
WITH vals AS
  (
    SELECT 1 x, 'a' y UNION ALL
    SELECT 1 x, 'b' y UNION ALL
    SELECT 2 x, 'a' y UNION ALL
    SELECT 2 x, 'c' y
  )
SELECT x, ARRAY_AGG(y) as array_agg
FROM vals
GROUP BY x;

/*---------------+
 | x | array_agg |
 +---------------+
 | 1 | [a, b]    |
 | 2 | [a, c]    |
 +---------------*/
```

```googlesql
SELECT
  x,
  ARRAY_AGG(x) OVER (ORDER BY ABS(x)) AS array_agg
FROM UNNEST([2, 1, -2, 3, -2, 1, 2]) AS x;

/*----+-------------------------+
 | x  | array_agg               |
 +----+-------------------------+
 | 1  | [1, 1]                  |
 | 1  | [1, 1]                  |
 | 2  | [1, 1, 2, -2, -2, 2]    |
 | -2 | [1, 1, 2, -2, -2, 2]    |
 | -2 | [1, 1, 2, -2, -2, 2]    |
 | 2  | [1, 1, 2, -2, -2, 2]    |
 | 3  | [1, 1, 2, -2, -2, 2, 3] |
 +----+-------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
