---
name: ANY_VALUE
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#any_value
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/any_value.yaml
---

# ANY_VALUE

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

## `ANY_VALUE`

```googlesql
ANY_VALUE(
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

Returns `expression` for some row chosen from the group. Which row is chosen is
nondeterministic, not random. Returns `NULL` when the input produces no
rows. Returns `NULL` when `expression`
or `having_expression` is
`NULL` for all rows in the group.

If `expression` contains any non-NULL values, then `ANY_VALUE` behaves as if
`IGNORE NULLS` is specified;
rows for which `expression` is `NULL` aren't considered and won't be
selected.

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

Any

**Returned Data Types**

Matches the input data type.

**Examples**

```googlesql
SELECT ANY_VALUE(fruit) as any_value
FROM UNNEST(["apple", "banana", "pear"]) as fruit;

/*-----------+
 | any_value |
 +-----------+
 | apple     |
 +-----------*/
```

```googlesql
SELECT
  fruit,
  ANY_VALUE(fruit) OVER (ORDER BY LENGTH(fruit) ROWS BETWEEN 1 PRECEDING AND CURRENT ROW) AS any_value
FROM UNNEST(["apple", "banana", "pear"]) as fruit;

/*--------+-----------+
 | fruit  | any_value |
 +--------+-----------+
 | pear   | pear      |
 | apple  | pear      |
 | banana | apple     |
 +--------+-----------*/
```

```googlesql
WITH
  Store AS (
    SELECT 20 AS sold, "apples" AS fruit
    UNION ALL
    SELECT 30 AS sold, "pears" AS fruit
    UNION ALL
    SELECT 30 AS sold, "bananas" AS fruit
    UNION ALL
    SELECT 10 AS sold, "oranges" AS fruit
  )
SELECT ANY_VALUE(fruit HAVING MAX sold) AS a_highest_selling_fruit FROM Store;

/*-------------------------+
 | a_highest_selling_fruit |
 +-------------------------+
 | pears                   |
 +-------------------------*/
```

```googlesql
WITH
  Store AS (
    SELECT 20 AS sold, "apples" AS fruit
    UNION ALL
    SELECT 30 AS sold, "pears" AS fruit
    UNION ALL
    SELECT 30 AS sold, "bananas" AS fruit
    UNION ALL
    SELECT 10 AS sold, "oranges" AS fruit
  )
SELECT ANY_VALUE(fruit HAVING MIN sold) AS a_lowest_selling_fruit FROM Store;

/*-------------------------+
 | a_lowest_selling_fruit  |
 +-------------------------+
 | oranges                 |
 +-------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
