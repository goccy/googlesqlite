---
name: GROUPING
dialect: googlesql
category: functions/aggregate
status: implemented
notes: |
  GROUP BY GROUPING SETS / ROLLUP / CUBE go through the AggregateScan
  formatter's GroupingSetList branch (internal/formatter.go), which
  expands the user query into a UNION ALL of per-grouping-set scans.
  Each scan substitutes the GROUPING(col) output column with the
  literal `0` or `1` based on whether `col` is in the current grouping
  set (groupingCallTargets lookup). The CLAUDE.md "tied-row order is
  implementation-defined" caveat means the upstream Examples need
  `unordered: true` because `ORDER BY product_name` alone does not
  break the within-group tie when product_name is NULL.
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#grouping
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/grouping.yaml
---

# GROUPING

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

## `GROUPING`

```googlesql
GROUPING(groupable_value)
```

**Description**

If a groupable item in the [`GROUP BY` clause][group-by-clause] is aggregated
(and thus not grouped), this function returns `1`. Otherwise,
this function returns `0`.

Definitions:

+ `groupable_value`: An expression that represents a value that can be grouped
  in the `GROUP BY` clause.

Details:

The `GROUPING` function is helpful if you need to determine which rows are
produced by which grouping sets. A grouping set is a group of columns by which
rows can be grouped together. So, if you need to filter rows by
a few specific grouping sets, you can use the `GROUPING` function to identify
which grouping sets grouped which rows by creating a matrix of the results.

In addition, you can use the `GROUPING` function to determine the type of
`NULL` produced by the `GROUP BY` clause. In some cases, the `GROUP BY` clause
produces a `NULL` placeholder. This placeholder represents all groupable items
that are aggregated (not grouped) in the current grouping set. This is different
from a standard `NULL`, which can also be produced by a query.

For more information, see the following examples.

**Returned Data Type**

`INT64`

**Examples**

In the following example, it's difficult to determine which rows are grouped by
the grouping value `product_type` or `product_name`. The `GROUPING` function
makes this easier to determine.

Pay close attention to what's in the `product_type_agg` and
`product_name_agg` column matrix. This determines how the rows are grouped.

`product_type_agg` | `product_name_agg` | Notes
------------------ | -------------------| ------
1                  | 0                  | Rows are grouped by `product_name`.
0                  | 1                  | Rows are grouped by `product_type`.
0                  | 0                  | Rows are grouped by `product_type` and `product_name`.
1                  | 1                  | Grand total row.

```googlesql
WITH
  Products AS (
    SELECT 'shirt' AS product_type, 't-shirt' AS product_name, 3 AS product_count UNION ALL
    SELECT 'shirt', 't-shirt', 8 UNION ALL
    SELECT 'shirt', 'polo', 25 UNION ALL
    SELECT 'pants', 'jeans', 6
  )
SELECT
  product_type,
  product_name,
  SUM(product_count) AS product_sum,
  GROUPING(product_type) AS product_type_agg,
  GROUPING(product_name) AS product_name_agg,
FROM Products
GROUP BY GROUPING SETS(product_type, product_name, ())
ORDER BY product_name;

/*--------------+--------------+-------------+------------------+------------------+
 | product_type | product_name | product_sum | product_type_agg | product_name_agg |
 +--------------+--------------+-------------+------------------+------------------+
 | NULL         | NULL         | 42          | 1                | 1                |
 | shirt        | NULL         | 36          | 0                | 1                |
 | pants        | NULL         | 6           | 0                | 1                |
 | NULL         | jeans        | 6           | 1                | 0                |
 | NULL         | polo         | 25          | 1                | 0                |
 | NULL         | t-shirt      | 11          | 1                | 0                |
 +--------------+--------------+-------------+------------------+------------------*/
```

In the following example, it's difficult to determine
if `NULL` represents a `NULL` placeholder or a standard `NULL` value in the
`product_type` column. The `GROUPING` function makes it easier to
determine what type of `NULL` is being produced. If
`product_type_is_aggregated` is `1`, the `NULL` value for
the `product_type` column is a `NULL` placeholder.

```googlesql
WITH
  Products AS (
    SELECT 'shirt' AS product_type, 't-shirt' AS product_name, 3 AS product_count UNION ALL
    SELECT 'shirt', 't-shirt', 8 UNION ALL
    SELECT NULL, 'polo', 25 UNION ALL
    SELECT 'pants', 'jeans', 6
  )
SELECT
  product_type,
  product_name,
  SUM(product_count) AS product_sum,
  GROUPING(product_type) AS product_type_is_aggregated
FROM Products
GROUP BY GROUPING SETS(product_type, product_name)
ORDER BY product_name;

/*--------------+--------------+-------------+----------------------------+
 | product_type | product_name | product_sum | product_type_is_aggregated |
 +--------------+--------------+-------------+----------------------------+
 | shirt        | NULL         | 11          | 0                          |
 | NULL         | NULL         | 25          | 0                          |
 | pants        | NULL         | 6           | 0                          |
 | NULL         | jeans        | 6           | 1                          |
 | NULL         | polo         | 25          | 1                          |
 | NULL         | t-shirt      | 11          | 1                          |
 +--------------+--------------+-------------+----------------------------*/
```

[group-by-clause]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#group_by_clause

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
